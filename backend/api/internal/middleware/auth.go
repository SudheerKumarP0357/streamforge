package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"streamforge/api/internal/auth"
	"streamforge/api/internal/cache"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
	UserRoleKey  contextKey = "user_role"
)

type AuthMiddleware struct {
	jwt   *auth.JWTService
	cache *cache.Cache
}

func NewAuthMiddleware(jwtService *auth.JWTService, cache *cache.Cache) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:   jwtService,
		cache: cache,
	}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			respondError(w, http.StatusUnauthorized, "Missing or invalid Authorization header")
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := m.jwt.ValidateAccessToken(tokenString)
		if err != nil {
			respondError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Verify session exists in Redis
		session, err := m.cache.GetSession(r.Context(), claims.UserID)
		if err != nil || session == nil {
			respondError(w, http.StatusUnauthorized, "Session expired or invalid")
			return
		}

		// Inject user context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := GetUserRole(r)
			if userRole != role {
				respondError(w, http.StatusForbidden, "Forbidden: insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID retrieves the user ID from the request context
func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(UserIDKey).(string)
	return userID
}

// GetUserRole retrieves the user role from the request context
func GetUserRole(r *http.Request) string {
	role, _ := r.Context().Value(UserRoleKey).(string)
	return role
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
