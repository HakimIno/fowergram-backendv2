package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"fowergram-backend/internal/config"
	"fowergram-backend/internal/domain/post"
	"fowergram-backend/internal/domain/user"
	"fowergram-backend/internal/graphql"
	"fowergram-backend/internal/handlers"
	"fowergram-backend/internal/infra/cache"
	"fowergram-backend/internal/infra/database"
	"fowergram-backend/internal/infra/messaging"
	"fowergram-backend/internal/infra/storage"
	"fowergram-backend/internal/routes"
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/logger"
	"fowergram-backend/pkg/telemetry"
)

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize logger
	logger := logger.NewZapLogger()
	defer logger.Sync()

	// Load configuration
	cfg := config.Load()

	// Initialize telemetry
	telemetry, err := telemetry.NewTelemetry(cfg.AppName, cfg.AppVersion)
	if err != nil {
		logger.Fatal("Failed to initialize telemetry", "error", err)
	}
	defer telemetry.Shutdown()

	// Initialize database
	db, err := database.NewPostgreSQLDB(cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Initialize cache
	cacheClient, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", "error", err)
	}
	defer cacheClient.Close()

	// Initialize storage
	storageClient, err := storage.NewMinIOStorage(cfg.Storage)
	if err != nil {
		logger.Fatal("Failed to initialize MinIO storage", "error", err)
	}

	// Initialize messaging
	msgClient, err := messaging.NewNATSClient(cfg.NatsURL)
	if err != nil {
		logger.Fatal("Failed to connect to NATS", "error", err)
	}
	defer msgClient.Close()

	// Initialize authentication with JWT
	userRepo := user.NewRepository(db)
	authService := auth.NewJWTAuth(auth.JWTConfig{
		Secret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		AccessExpiry:  time.Hour * 1,       // 1 hour
		RefreshExpiry: time.Hour * 24 * 30, // 30 days
	}, userRepo)

	// Initialize repositories
	postRepo := post.NewRepository(db)

	// Initialize services
	userService := user.NewService(userRepo, cacheClient, authService, logger)
	postService := post.NewService(postRepo, userRepo, storageClient, cacheClient, msgClient, logger)

	// Initialize GraphQL server
	gqlServer := graphql.NewServer(userService, postService, authService, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	healthHandler := handlers.NewHealthHandler(cfg.AppVersion)
	postHandler := handlers.NewPostHandler(postService, logger)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1"},
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		IdleTimeout:             120 * time.Second,
	})

	// Setup routes
	routes.SetupRoutes(app, routes.Config{
		AuthHandler:    authHandler,
		HealthHandler:  healthHandler,
		PostHandler:    postHandler,
		AuthService:    authService,
		GQLHandler:     adaptor.HTTPHandler(gqlServer),
		MetricsHandler: adaptor.HTTPHandler(telemetry.PrometheusHandler()),
		AllowedOrigins: cfg.AllowedOrigins,
	})

	// Setup development routes if in development mode
	if cfg.Environment == "development" {
		routes.SetupDevelopmentRoutes(app, adaptor.HTTPHandler(graphql.NewPlayground("/graphql")))
	}

	// Graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			logger.Error("Server forced to shutdown", "error", err)
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	logger.Info("Starting server", "port", port, "environment", cfg.Environment)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
