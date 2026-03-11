package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"streamforge/transcoder/internal/config"
	"streamforge/transcoder/internal/consumer"
	"streamforge/transcoder/internal/db"
	"streamforge/transcoder/internal/ffmpeg"
	"streamforge/transcoder/internal/metrics" // Add this import
	"streamforge/transcoder/internal/pipeline"
	"streamforge/transcoder/internal/storage"

	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Dev only: load .env file
	_ = godotenv.Load()

	metrics.Init()

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		slog.Info("Starting Prometheus metrics server on :9090")
		if err := http.ListenAndServe(":9090", mux); err != nil {
			slog.Error("Metrics server failed", "error", err)
		}
	}()

	cfg := config.Load()

	// Initialize Storage
	blobClient := storage.NewBlobClient(cfg)
	if blobClient == nil {
		slog.Error("Blob storage initialization failed")
		os.Exit(1)
	}

	// Initialize Database
	dbClient := db.NewPostgresClient(cfg)
	defer dbClient.Close()

	// Initialize FFmpeg Transcoder wrapper
	transcoder := ffmpeg.NewTranscoder("", "") // uses defaults: "ffmpeg", os.TempDir()

	// Initialize RabbitMQ Consumer
	queueName := "transcoder.jobs"
	rmqConsumer, err := consumer.NewConsumer(cfg.RabbitMQURL, queueName)
	if err != nil {
		slog.Error("Failed to initialize RabbitMQ consumer", "error", err)
		os.Exit(1)
	}
	defer rmqConsumer.Close()

	// Initialize Dead Letter Queue (DLQ) Consumer
	dlqName := "transcoder.jobs.dlq"
	dlqConsumer, err := consumer.NewDLQConsumer(cfg.RabbitMQURL, dlqName)
	if err != nil {
		slog.Warn("Failed to initialize DLQ consumer, continuing without DLQ processing", "error", err)
	} else {
		defer dlqConsumer.Close()
	}

	// Initialize worker Pipeline
	pipe := pipeline.NewPipeline(rmqConsumer, transcoder, blobClient, dbClient)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("Received termination signal, shutting down worker...")
		cancel() // trigger context cancellation to gracefully halt the consumer blocked loop
	}()

	slog.Info("StreamForge Transcoder Worker started.")
	
	// 1. Start RabbitMQ Queue Monitoring routines
	go rmqConsumer.MonitorQueues(ctx, dlqName)

	// 2. Start DLQ Worker if successfully dialed
	if dlqConsumer != nil {
		go func() {
			err := dlqConsumer.Start(ctx, func(jobCtx context.Context, job consumer.TranscodeJob, failCount int) error {
				slog.Warn("DLQ retry attempt", "video_id", job.VideoID, "fail_count", failCount)
				return pipe.Process(jobCtx, job) // recycle back to the main transcoder routine
			})
			if err != nil && ctx.Err() == nil {
				slog.Error("DLQ consumer exited with error", "error", err)
			}
		}()
	}

	// 3. Start Main Worker Consumer Blocks
	if err := rmqConsumer.Start(ctx, pipe.Process); err != nil {
		slog.Error("Main Worker loop exited", "error", err)
	}

	slog.Info("Worker exited gracefully.")
}
