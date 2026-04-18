package handlers

import (
	"encoding/json"
	"log/slog"
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
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "AuthHandler.Register")
	logger.Info("starting")

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger = logger.With("email", req.Email)

	if err := auth.ValidateRegistration(req.Email, req.Password); err != nil {
		logger.Warn("validation failed", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	existingUser, _ := h.db.GetUserByEmail(r.Context(), req.Email)
	if existingUser != nil {
		logger.Warn("email already exists", "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusConflict, "Email already exists")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		logger.Error("failed to hash password", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	userID, err := h.db.CreateUser(r.Context(), req.Email, hash)
	if err != nil {
		logger.Error("failed to create user", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to create user")
		return
	}

	respond(w, http.StatusCreated, map[string]string{
		"user_id": userID,
		"email":   req.Email,
	})
	logger.Info("completed", "outcome", "success", "user_id", userID, "duration_ms", time.Since(start).Milliseconds())
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "AuthHandler.Login")
	logger.Info("starting")

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	logger = logger.With("email", req.Email)

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user == nil {
		logger.Warn("invalid credentials", "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		logger.Warn("invalid credentials", "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("failed to generate access token", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	refreshToken, err := h.jwt.GenerateRefreshToken()
	if err != nil {
		logger.Error("failed to generate refresh token", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to generate refresh token")
		return
	}

	tokenHash := h.jwt.HashRefreshToken(refreshToken)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	if err := h.db.CreateRefreshToken(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		logger.Error("failed to store refresh token", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to store refresh token")
		return
	}

	session := cache.SessionData{UserID: user.ID, Email: user.Email, Role: user.Role}
	if err := h.cache.SetSession(r.Context(), user.ID, session, 15*time.Minute); err != nil {
		logger.Error("failed to set session", "error", err.Error())
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600,
	})
	logger.Info("completed", "outcome", "success", "user_id", user.ID, "duration_ms", time.Since(start).Milliseconds())
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "AuthHandler.Refresh")
	logger.Info("starting")

	var req models.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid request body", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.RefreshToken == "" {
		logger.Warn("missing refresh token", "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusBadRequest, "Refresh token is required")
		return
	}

	tokenHash := h.jwt.HashRefreshToken(req.RefreshToken)

	dbToken, err := h.db.GetRefreshToken(r.Context(), tokenHash)
	if err != nil || dbToken == nil {
		logger.Warn("invalid refresh token", "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusUnauthorized, "Invalid refresh token")
		return
	}

	if dbToken.Revoked {
		logger.Warn("revoked refresh token used", "user_id", dbToken.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusUnauthorized, "Refresh token is revoked")
		return
	}

	if time.Now().After(dbToken.ExpiresAt) {
		logger.Warn("expired refresh token used", "user_id", dbToken.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusUnauthorized, "Refresh token is expired")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), dbToken.UserID)
	if err != nil || user == nil {
		logger.Error("user not found for refresh", "user_id", dbToken.UserID, "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "User not found")
		return
	}

	accessToken, err := h.jwt.GenerateAccessToken(user.ID, user.Email, user.Role)
	if err != nil {
		logger.Error("failed to generate access token", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
		respondError(w, http.StatusInternalServerError, "Failed to generate access token")
		return
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"expires_in":   3600,
	})
	logger.Info("completed", "outcome", "success", "user_id", user.ID, "duration_ms", time.Since(start).Milliseconds())
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "AuthHandler.Logout")

	userID := middleware.GetUserID(r)
	if userID == "" {
		respondError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	logger.Info("starting", "user_id", userID)

	if err := h.cache.DeleteSession(r.Context(), userID); err != nil {
		logger.Error("failed to delete session", "error", err.Error())
	}

	respond(w, http.StatusOK, map[string]string{"message": "logged out"})
	logger.Info("completed", "outcome", "success", "user_id", userID, "duration_ms", time.Since(start).Milliseconds())
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
