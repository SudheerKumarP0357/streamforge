package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"streamforge/api/internal/auth"
	"streamforge/api/internal/cache"
	"streamforge/api/internal/config"
	"streamforge/api/internal/db"
	"streamforge/api/internal/handlers"
	"streamforge/api/internal/metrics"
	customMiddleware "streamforge/api/internal/middleware"
	queuepkg "streamforge/api/internal/queue"
	storagepkg "streamforge/api/internal/storage"
)

func New(cfg config.Config, dbClient *db.DB, cosmosClient *db.CosmosDB, rmq *queuepkg.Publisher, blob *storagepkg.BlobStorage, cacheClient *cache.Cache, jwtService *auth.JWTService) *chi.Mux {
	metrics.Init()

	r := chi.NewRouter()

	// Chi middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(customMiddleware.Logger)
	r.Use(customMiddleware.MetricsMiddleware) // Add metrics middleware
	r.Use(middleware.Recoverer)

	// Custom CORS and RateLimit middlewares
	corsMiddleware := customMiddleware.NewCORSMiddleware(cfg.CORSAllowedOrigin)
	r.Use(corsMiddleware.Handler)

	r.Use(customMiddleware.NewRateLimiter(cacheClient, 200))

	// Auth Middleware
	authMiddleware := customMiddleware.NewAuthMiddleware(jwtService, cacheClient)

	// Expose Prometheus Metrics endpoint
	r.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Health check endpoint
	healthHandler := handlers.NewHealthHandler(dbClient, cosmosClient, cacheClient, rmq)
	r.Get("/healthz", healthHandler.Health)

	// API groups
	r.Route("/api/v1", func(r chi.Router) {
		
		// Auth routes
		authHandler := handlers.NewAuthHandler(dbClient, cacheClient, jwtService)
		r.Route("/auth", func(r chi.Router) {
			r.Use(customMiddleware.NewRateLimiter(cacheClient, 10))
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			
			// Protected Auth route
			r.With(authMiddleware.Authenticate).Post("/logout", authHandler.Logout)
		})

		// Admin routes
		adminHandler := handlers.NewAdminHandler(dbClient)
		r.Route("/admin", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Use(authMiddleware.RequireRole("admin"))
			r.Get("/videos", adminHandler.GetVideos)
			r.Get("/stats", adminHandler.GetStats)
		})

		// Videos routes
		videoHandler := handlers.NewVideoHandler(cfg, dbClient, rmq, blob, cacheClient)
		r.Route("/videos", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Get("/", videoHandler.List)
			r.Get("/search", videoHandler.Search)
			r.Post("/", videoHandler.Upload)
			r.Get("/{id}", videoHandler.GetByID)
			r.Delete("/{id}", videoHandler.Delete)
			r.Get("/{id}/stream", videoHandler.GetStreamURL)
		})

		// Watch history routes
		watchHandler := handlers.NewWatchHandler(cosmosClient, dbClient)
		r.Route("/watch", func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)
			r.Post("/{video_id}/event", watchHandler.SaveEvent)
			r.Get("/history", watchHandler.GetHistory)
		})
	})

	return r
}
