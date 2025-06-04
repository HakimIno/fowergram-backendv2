package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
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
	"fowergram-backend/pkg/email"
	"fowergram-backend/pkg/logger"
	"fowergram-backend/pkg/middleware"
	"fowergram-backend/pkg/telemetry"
)

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

	// Initialize email service
	emailService := email.NewSMTPEmailService(email.EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvAsInt("SMTP_PORT", 587),
		SMTPUsername: getEnv("SMTP_USERNAME", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("SMTP_FROM_EMAIL", "noreply@fowergram.com"),
		FromName:     getEnv("SMTP_FROM_NAME", "Fowergram"),
		BaseURL:      getEnv("APP_URL", "http://localhost:3000"),
	})

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		RedisClient: cacheClient.GetClient(),
		MaxRequests: 5,           // 5 requests
		Window:      time.Minute, // per minute
	})

	// Initialize repositories
	userRepo := user.NewPostgresRepository(db)
	verificationRepo := user.NewPostgresVerificationRepository(db)
	postRepo := post.NewRepository(db)

	// Initialize authentication with JWT
	authService := auth.NewJWTAuth(
		getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		time.Hour*1,     // 1 hour access token
		time.Hour*24*30, // 30 days refresh token
		userRepo,
		verificationRepo,
		emailService,
	)

	// Initialize services
	userService := user.NewService(userRepo, cacheClient, authService, logger)
	postService := post.NewService(postRepo, userRepo, storageClient, cacheClient, msgClient, logger)

	// Initialize GraphQL server
	gqlServer := graphql.NewServer(userService, postService, authService, logger)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, emailService, logger)
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
		AllowedOrigins: getEnv("ALLOWED_ORIGINS", "*"),
		RateLimiter:    rateLimiter,
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
	port := getEnv("PORT", "8000")
	logger.Info("Starting server", "port", port)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Failed to start server", "error", err)
	}
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
