package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

// RequestIDKey is the context key for the per-request UUID.
const RequestIDKey contextKey = "request_id"

// GetRequestID retrieves the request ID from the request context.
func GetRequestID(r *http.Request) string {
	id, _ := r.Context().Value(RequestIDKey).(string)
	return id
}

// Logger is a structured logging middleware that captures method, path, query,
// status, latency, client IP, user ID, response size, referer, and user-agent
// for every HTTP request. It uses slog with level-based logging:
//   - INFO  for 2xx/3xx responses
//   - WARN  for 4xx responses
//   - ERROR for 5xx responses
//
// A UUID request_id is generated per request, injected into context, and
// returned in the X-Request-ID response header so downstream handlers and
// callers can correlate logs.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a unique request ID and propagate it.
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Set the request ID on the response so clients can reference it.
		w.Header().Set("X-Request-ID", requestID)

		// Wrap the response writer to capture status code and bytes written.
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			latencyMs := float64(time.Since(start).Microseconds()) / 1000.0

			userID := GetUserID(r)
			if userID == "" {
				userID = "-"
			}

			attrs := []slog.Attr{
				slog.String("request_id", requestID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("query", r.URL.RawQuery),
				slog.Int("status", status),
				slog.Float64("latency_ms", latencyMs),
				slog.String("client_ip", clientIP(r)),
				slog.String("user_id", userID),
				slog.Int("response_size", ww.BytesWritten()),
				slog.String("referer", r.Referer()),
				slog.String("user_agent", r.UserAgent()),
			}

			// Choose log level based on status code range.
			level := slog.LevelInfo
			switch {
			case status >= 500:
				level = slog.LevelError
			case status >= 400:
				level = slog.LevelWarn
			}

			slog.LogAttrs(r.Context(), level, "http request", attrs...)
		}()

		next.ServeHTTP(ww, r)
	})
}

// clientIP returns the client's IP address, respecting the X-Forwarded-For
// header for proxied requests. If the header contains a comma-separated list,
// the first (left-most) entry is used as it represents the original client.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// The first IP in the chain is the original client.
		if ip, _, found := strings.Cut(xff, ","); found {
			return strings.TrimSpace(ip)
		}
		return strings.TrimSpace(xff)
	}
	return r.RemoteAddr
}
