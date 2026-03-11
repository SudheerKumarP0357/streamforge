package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// Logger is a custom chi middleware that logs method, path, status code, duration, and userID (if authenticated)
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		defer func() {
			duration := time.Since(start)
			status := ww.Status()
			if status == 0 {
				status = http.StatusOK
			}

			userID := GetUserID(r)
			if userID == "" {
				userID = "-"
			}

			log.Printf(
				"[%s] %s %s | Status: %d | Duration: %v",
				userID,
				r.Method,
				r.URL.Path,
				status,
				duration,
			)
		}()

		next.ServeHTTP(ww, r)
	})
}
