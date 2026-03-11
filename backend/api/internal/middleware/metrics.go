package middleware

import (
	"net/http"
	"strconv"
	"time"

	"streamforge/api/internal/metrics"
)

// responseWriter captures the HTTP status code for metrics labeling.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// MetricsMiddleware tracks the active connections, request duration, and totals.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		metrics.ActiveConnections.Inc()
		defer metrics.ActiveConnections.Dec()

		rw := &responseWriter{w, http.StatusOK}
		
		// Serve the request
		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		// To prevent high cardinality, you might resolve path parameters (e.g. /video/123 -> /video/:id) here. 
		// For simplicity, we use the raw URL Path but it should ideally be normalized.
		path := r.URL.Path

		metrics.HTTPRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
	})
}
