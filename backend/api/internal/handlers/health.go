package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"streamforge/api/internal/cache"
	"streamforge/api/internal/db"
	"streamforge/api/internal/middleware"
	"streamforge/api/internal/queue"
)

var startTime = time.Now()

type HealthHandler struct {
	dbClient     *db.DB
	cosmosClient *db.CosmosDB
	cacheClient  *cache.Cache
	rmq          *queue.Publisher
}

func NewHealthHandler(dbClient *db.DB, cosmosClient *db.CosmosDB, cacheClient *cache.Cache, rmq *queue.Publisher) *HealthHandler {
	return &HealthHandler{
		dbClient:     dbClient,
		cosmosClient: cosmosClient,
		cacheClient:  cacheClient,
		rmq:          rmq,
	}
}

type HealthResponse struct {
	Status        string            `json:"status"`
	Version       string            `json:"version"`
	UptimeSeconds int64             `json:"uptime_seconds"`
	Dependencies  map[string]string `json:"dependencies"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	logger := slog.With("request_id", middleware.GetRequestID(r), "handler", "HealthHandler.Health")
	logger.Info("starting")

	ctx := r.Context()
	deps := make(map[string]string)
	status := "ok"

	// Check PostgreSQL
	if err := h.dbClient.Ping(ctx); err != nil {
		deps["db"] = fmt.Sprintf("error: %v", err)
		status = "degraded"
	} else {
		deps["db"] = "ok"
	}

	// Check Redis
	if err := h.cacheClient.Ping(ctx); err != nil {
		deps["redis"] = "error"
		status = "degraded"
	} else {
		deps["redis"] = "ok"
	}

	// Check RabbitMQ
	if err := h.rmq.IsHealthy(); err != nil {
		deps["rabbitmq"] = "error"
		status = "degraded"
	} else {
		deps["rabbitmq"] = "ok"
	}

	// Check Cosmos DB
	if err := h.cosmosClient.Ping(ctx); err != nil {
		deps["cosmos"] = "error"
		status = "degraded"
	} else {
		deps["cosmos"] = "ok"
	}

	httpStatus := http.StatusOK
	if status == "degraded" {
		httpStatus = http.StatusServiceUnavailable
	}

	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "dev"
	}

	uptime := int64(time.Since(startTime).Seconds())

	resp := HealthResponse{
		Status:        status,
		Version:       version,
		UptimeSeconds: uptime,
		Dependencies:  deps,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(resp)

	if status == "degraded" {
		logger.Warn("completed", "outcome", "degraded", "dependencies", deps, "duration_ms", time.Since(start).Milliseconds())
	} else {
		logger.Info("completed", "outcome", "success", "duration_ms", time.Since(start).Milliseconds())
	}
}
