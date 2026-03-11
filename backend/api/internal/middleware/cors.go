package middleware

import (
	"net/http"
	"strings"
)

type CORSMiddleware struct {
	allowedOrigin string
}

func NewCORSMiddleware(allowedOrigin string) *CORSMiddleware {
	return &CORSMiddleware{
		allowedOrigin: strings.TrimRight(allowedOrigin, "/"),
	}
}

func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", m.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
