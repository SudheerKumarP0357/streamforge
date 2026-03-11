package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"streamforge/api/internal/auth"
	"streamforge/api/internal/cache"
	"streamforge/api/internal/config"
	"streamforge/api/internal/db"
	queuepkg "streamforge/api/internal/queue"
	"streamforge/api/internal/router"
	storagepkg "streamforge/api/internal/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env only in development
	env := os.Getenv("ENV")
	if env == "" || env == "development" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, falling back to environment variables")
		}
	}

	cfg := config.Load()

	// Initialize DB Client
	dbURL := "postgres://" + cfg.PostgresUser + ":" + cfg.PostgresPassword + "@" + cfg.PostgresHost + ":" + cfg.PostgresPort + "/" + cfg.PostgresDB + "?sslmode=" + cfg.PostgresSSLMode
	dbClient, err := db.NewDB(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	// Run Database migrations
	if err := db.RunMigrations(dbURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initializing remaining dependencies
	cosmosClient, err := db.NewCosmos(context.Background(), cfg.CosmosConnectionString, cfg.CosmosDatabase)
	if err != nil {
		log.Fatalf("Failed to initialize Cosmos DB: %v", err)
	}
	defer cosmosClient.Close(context.Background())
	
	redisClient, err := cache.NewRedis(cfg.RedisURL, cfg.RedisPassword)
	if err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer redisClient.Close()

	jwtService := auth.NewJWTService(cfg.JWTSecret)
	
	rabbitMQ, err := queuepkg.NewPublisher(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	defer rabbitMQ.Close()
	
	blobClient, err := storagepkg.NewBlobStorage(cfg.AzureStorageName, cfg.AzureStorageKey, cfg.AzureStorageEndpoint, cfg.AzureBlobRawContainer, cfg.AzureBlobHLSContainer)
	if err != nil {
		log.Fatalf("Failed to initialize Blob Storage: %v", err)
	}

	r := router.New(cfg, dbClient, cosmosClient, rabbitMQ, blobClient, redisClient, jwtService)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	// Graceful shutdown setup
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Received termination signal, shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(idleConnsClosed)
	}()

	log.Printf("Starting server on port %s", cfg.Port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

	<-idleConnsClosed
	log.Println("HTTP server shutdown complete")
}
