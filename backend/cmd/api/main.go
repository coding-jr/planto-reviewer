package main

import (
	"log"

	"github.com/coding-jr/planto-reviewer/backend/internal/config"
	"github.com/coding-jr/planto-reviewer/backend/internal/database"
	"github.com/coding-jr/planto-reviewer/backend/internal/handlers"
	"github.com/coding-jr/planto-reviewer/backend/internal/middleware"
	"github.com/coding-jr/planto-reviewer/backend/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL, cfg.Env == "development")
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	// Initialize repositories
	orgRepo := repositories.NewOrganizationRepository(db)

	// Initialize handlers
	orgHandler := handlers.NewOrganizationHandler(orgRepo)
	metricsHandler := handlers.NewMetricsHandler(db)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Health check (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	// API routes (with auth)
	api := app.Group("/api")
	if cfg.APIKey != "" {
		api.Use(middleware.APIKeyAuth(cfg.APIKey))
	}

	// Organization routes
	api.Post("/organizations", orgHandler.Create)
	api.Get("/organizations", orgHandler.List)
	api.Get("/organizations/:id", orgHandler.Get)
	api.Put("/organizations/:id", orgHandler.Update)
	api.Delete("/organizations/:id", orgHandler.Delete)

	// Metrics routes
	api.Get("/metrics/developer/:id", metricsHandler.GetDeveloperMetrics)
	api.Get("/metrics/organization/:id/summary", metricsHandler.GetOrgSummary)
	api.Get("/metrics/organization/:id/leaderboard", metricsHandler.GetLeaderboard)
	api.Get("/metrics/organization/:id/top-issues", metricsHandler.GetTopIssues)

	// Start server
	log.Printf("🚀 API server starting on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}
