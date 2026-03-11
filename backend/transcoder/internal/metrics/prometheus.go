package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	TranscodeJobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "transcode_jobs_total",
			Help: "Total number of completed transcode jobs",
		},
		[]string{"status"}, // "success", "failed"
	)

	TranscodeDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "transcode_duration_seconds",
			Help:    "Duration of successful video transcodes in seconds",
			Buckets: []float64{10, 30, 60, 120, 300, 600, 1800},
		},
	)

	RabbitMQConsumedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_consumed_total",
			Help: "Total number of messages consumed from RabbitMQ",
		},
		[]string{"queue"},
	)
)

func Init() {
	// Registered automatically by promauto
}
