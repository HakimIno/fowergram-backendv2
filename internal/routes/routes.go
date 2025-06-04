package routes

import (
	"fowergram-backend/internal/handlers"
	"fowergram-backend/pkg/auth"
	"fowergram-backend/pkg/middleware"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Config holds dependencies for route setup
type Config struct {
	AuthHandler    *handlers.AuthHandler
	HealthHandler  *handlers.HealthHandler
	PostHandler    *handlers.PostHandler
	AuthService    auth.AuthService
	GQLHandler     fiber.Handler
	MetricsHandler fiber.Handler
	AllowedOrigins string
	RateLimiter    *middleware.RateLimiter
}

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, cfg Config) {
	// Middleware
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: true,
	}))

	// Health check endpoint
	app.Get("/health", cfg.HealthHandler.Health)

	// API Documentation (Stoplight Elements) - static files
	app.Static("/docs", "./api", fiber.Static{
		Index:  "stoplight.html",
		Browse: true,
	})

	// Root redirect to docs
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs/")
	})

	// API routes
	api := app.Group("/api")

	// Authentication routes
	auth := api.Group("/auth")

	// Public auth routes with rate limiting
	auth.Post("/signup", cfg.RateLimiter.Middleware(), cfg.AuthHandler.Signup)
	auth.Post("/signin", cfg.RateLimiter.Middleware(), cfg.AuthHandler.Signin)
	auth.Post("/signout", cfg.AuthHandler.Signout)

	// Email verification routes
	auth.Post("/verify-email", cfg.RateLimiter.Middleware(), cfg.AuthHandler.VerifyEmail)
	auth.Post("/request-password-reset", cfg.RateLimiter.Middleware(), cfg.AuthHandler.RequestPasswordReset)
	auth.Post("/reset-password", cfg.RateLimiter.Middleware(), cfg.AuthHandler.ResetPassword)

	// Protected routes
	protected := api.Group("/auth")
	protected.Use(cfg.AuthService.Middleware())
	protected.Get("/me", cfg.AuthHandler.Me)

	// Posts routes (protected)
	if cfg.PostHandler != nil {
		posts := api.Group("/posts")
		posts.Use(cfg.AuthService.Middleware())
		posts.Post("/", cfg.PostHandler.CreatePost)
		posts.Get("/", cfg.PostHandler.GetPosts)
		posts.Get("/:id", cfg.PostHandler.GetPost)
		posts.Put("/:id", cfg.PostHandler.UpdatePost)
		posts.Delete("/:id", cfg.PostHandler.DeletePost)
	}

	// GraphQL endpoint
	if cfg.GQLHandler != nil {
		app.Post("/graphql", cfg.GQLHandler)
	}

	// Metrics endpoint
	if cfg.MetricsHandler != nil {
		app.Get("/metrics", cfg.MetricsHandler)
	}
}

// SetupDevelopmentRoutes adds development-only routes
func SetupDevelopmentRoutes(app *fiber.App, playgroundHandler fiber.Handler) {
	// GraphQL playground (development only)
	app.Get("/playground", playgroundHandler)
}
