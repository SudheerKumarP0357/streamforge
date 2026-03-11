package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"streamforge/api/internal/auth"
	"streamforge/api/internal/cache"
	"streamforge/api/internal/db"
	"streamforge/api/internal/middleware"
	"streamforge/api/internal/models"
)

type AuthHandler struct {
	db    *db.DB
	cache *cache.Cache
	jwt   *auth.JWTService
}

func NewAuthHandler(db *db.DB, cache *cache.Cache, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		db:    db,
		cache: cache,
		jwt:   jwtService,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := auth.ValidateRegistration(req.Email, req.Password); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	existingUser, _ := h.db.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		respondError(w, http.StatusConflict, "Email already exists")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("hash password err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	userID, err := h.db.CreateUser(r.Context(), req.Email, hash)
	if err != nil {
		log.Printf("create user err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respond(w, http.StatusCreated, map[string]string{
		"user_id": userID,
		"email":   req.Email,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("generate access token err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwt.GenerateRefreshToken()
	if err != nil {
		log.Printf("generate refresh token err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	tokenHash := h.jwt.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour) // matching JWTService refreshExpiry
	
	if err := h.db.CreateRefreshToken(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		log.Printf("create refresh token err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	session := cache.SessionData{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}
	if err := h.cache.SetSession(r.Context(), user.ID, session, 15*time.Minute); err != nil {
		log.Printf("set session err: %v", err)
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
	})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	tokenHash := h.jwt.HashRefreshToken(req.RefreshToken)
	
	dbToken, err := h.db.GetRefreshToken(r.Context(), tokenHash)
	if err != nil || dbToken == nil {
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if dbToken.Revoked {
		respondError(w, http.StatusUnauthorized, "Refresh token is revoked")
		return
	}

	if time.Now().After(dbToken.ExpiresAt) {
		respondError(w, http.StatusUnauthorized, "Refresh token is expired")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), dbToken.UserID)
	if err != nil || user == nil {
		respondError(w, http.StatusInternalServerError, "User not found")
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		log.Printf("generate access token err: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   3600,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	
	if err := h.cache.DeleteSession(r.Context(), userID); err != nil {
		log.Printf("delete session err: %v", err)
	}

	respond(w, http.StatusOK, map[string]string{
		"message": "logged out",
	})
}

func respond(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}
