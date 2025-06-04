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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"

	"fowergram-backend/internal/config"
	"fowergram-backend/internal/domain/post"
	"fowergram-backend/internal/domain/user"
	"fowergram-backend/internal/graphql"
	"fowergram-backend/internal/infra/cache"
	"fowergram-backend/internal/infra/database"
	"fowergram-backend/internal/infra/messaging"
	"fowergram-backend/internal/infra/storage"
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

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "::1"},
		ReadTimeout:             30 * time.Second,
		WriteTimeout:            30 * time.Second,
		IdleTimeout:             120 * time.Second,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   cfg.AppVersion,
		})
	})

	// SuperTokens middleware setup
	app.Use(authService.Middleware())

	// Authentication REST API endpoints
	app.Post("/api/auth/signup", func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Username string `json:"username"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if req.Email == "" || req.Password == "" || req.Username == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Email, password, and username are required"})
		}

		user, err := authService.CreateUser(c.Context(), req.Email, req.Password, req.Username)
		if err != nil {
			logger.Error("Failed to create user", "error", err)
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":       user.ID.String(),
				"email":    user.Email,
				"username": req.Username,
			},
			"message": "User created successfully",
		})
	})

	app.Post("/api/auth/signin", func(c *fiber.Ctx) error {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if req.Email == "" || req.Password == "" {
			return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
		}

		user, token, err := authService.SignIn(c.Context(), req.Email, req.Password)
		if err != nil {
			logger.Error("Failed to sign in", "error", err)
			return c.Status(401).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":    user.ID.String(),
				"email": user.Email,
			},
			"accessToken": token,
			"message":     "Signed in successfully",
		})
	})

	app.Post("/api/auth/signout", func(c *fiber.Ctx) error {
		// SuperTokens handles signout via session management
		return c.JSON(fiber.Map{
			"message": "Signed out successfully",
		})
	})

	app.Get("/api/auth/me", func(c *fiber.Ctx) error {
		// Get user directly from Fiber context locals
		user, ok := c.Locals("user").(*auth.User)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Not authenticated"})
		}

		return c.JSON(fiber.Map{
			"user": fiber.Map{
				"id":       user.ID.String(),
				"email":    user.Email,
				"username": user.Username,
			},
		})
	})

	// GraphQL endpoint
	app.All("/graphql", adaptor.HTTPHandler(gqlServer))

	// GraphQL playground (development only)
	if cfg.Environment == "development" {
		app.Get("/playground", adaptor.HTTPHandler(graphql.NewPlayground("/graphql")))
	}

	// Metrics endpoint for Prometheus
	app.Get("/metrics", adaptor.HTTPHandler(telemetry.PrometheusHandler()))

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
		port = "8080"
	}

	logger.Info("Starting server", "port", port, "environment", cfg.Environment)
	if err := app.Listen(":" + port); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}
