package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections/requests being processed",
		},
	)

	BlobUploadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "blob_upload_duration_seconds",
			Help:    "Duration of video file uploads to blob storage in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
		},
		[]string{"container"},
	)

	RabbitMQPublishedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_published_total",
			Help: "Total number of messages published to RabbitMQ",
		},
		[]string{"queue", "status"},
	)
)

// Init registers all custom metrics. promauto actually registers them automatically
// against the default registry, but we leave this here in case we want to customize.
func Init() {
	// Custom registrars if needed
}
