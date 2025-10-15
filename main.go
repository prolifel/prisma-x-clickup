package main

import (
	"io"
	"os"
	"prisma-webhook/config"
	"prisma-webhook/handlers"
	"prisma-webhook/middleware"
	"prisma-webhook/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// default logger
	file, err := os.OpenFile("/logs/webhook.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	iw := io.MultiWriter(os.Stdout, file)
	log.SetOutput(iw)
	defer file.Close()

	// Initialize services
	clickUpClient := services.NewClickUpClient(cfg)
	teamsClient := services.NewTeamsClient(cfg)

	// Initialize handlers
	webhookHandler := handlers.NewWebhookHandler(clickUpClient, teamsClient)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Prisma Cloud to ClickUp Webhook",
	})

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
			"version": "1.3.0",
			"status":  "running",
		})
	})

	// Webhook endpoint - with IP allowlist, API key auth, and rate limit
	app.Post("/webhook",
		middleware.IPAllowlist(cfg.AllowedIPs),
		middleware.APIKeyAuth(cfg.WebhookAPIKey),
		middleware.WebhookRateLimit(),
		func(c *fiber.Ctx) error {
			xType := c.Get("X-Type")

			// choose handler based on header
			if xType != "alerta" && xType != "mandatory" {
				log.Debugf("Unknown X-Type: %s", xType)
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid or missing type header",
				})
			}

			return c.Next()
		},
		webhookHandler.HandlePrismaWebhook,
	)

	// Start server
	log.Debugf("Starting server on port %s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
