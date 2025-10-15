package main

import (
	"log"
	"os"
	"prisma-webhook/config"
	"prisma-webhook/handlers"
	"prisma-webhook/middleware"
	"prisma-webhook/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize services
	clickUpClient := services.NewClickUpClient(cfg)
	teamsClient := services.NewTeamsClient(cfg)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(clickUpClient, teamsClient)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Prisma Cloud to ClickUp Webhook",
	})

	// Middleware
	// Custom File Writer
	file, err := os.OpenFile("./webhook.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer file.Close()
	app.Use(logger.New(logger.Config{
		Output: file,
	}))
	app.Use(recover.New())

	// Routes
	// Health check - no rate limit for monitoring
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
		})
	})

	// Root endpoint - with rate limit
	app.Get("/", middleware.GeneralRateLimit(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"service": "Prisma Cloud to ClickUp Webhook",
			"version": "1.2.0",
			"status":  "running",
		})
	})

	// Webhook endpoint - with IP allowlist, API key auth, and rate limit
	app.Post("/webhook",
		middleware.IPAllowlist(cfg.AllowedIPs),
		middleware.APIKeyAuth(cfg.WebhookAPIKey),
		middleware.WebhookRateLimit(),
		webhookHandler.HandlePrismaWebhook,
	)

	// Start server
	log.Printf("Starting server on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
