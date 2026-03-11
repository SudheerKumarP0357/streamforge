package middleware

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"streamforge/api/internal/cache"
)

func NewRateLimiter(cacheClient *cache.Cache, limit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				var err error
				ip, _, err = net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					ip = r.RemoteAddr
				}
			}

			// Handle comma-separated list of IPs in X-Forwarded-For
			if strings.Contains(ip, ",") {
				ip = strings.Split(ip, ",")[0]
			}
			ip = strings.TrimSpace(ip)

			routeCtx := chi.RouteContext(r.Context())
			path := r.URL.Path
			if routeCtx != nil && routeCtx.RoutePattern() != "" {
				path = routeCtx.RoutePattern()
			}

			count, err := cacheClient.IncrRateLimit(r.Context(), ip, path)
			if err != nil {
				log.Printf("Rate limit error: %v", err)
				// Proceed if cache is down to prevent complete outage
			} else {
				remaining := limit - int(count)
				if remaining < 0 {
					remaining = 0
				}
				
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
				w.Header().Set("X-RateLimit-Reset", "60")

				if count > int64(limit) {
					w.Header().Set("Retry-After", "60")
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]any{
						"error":       "rate limit exceeded",
						"retry_after": 60,
					})
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
