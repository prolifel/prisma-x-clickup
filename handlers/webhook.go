package handlers

import (
	"log"
	"prisma-webhook/models"
	"prisma-webhook/services"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	clickUpClient *services.ClickUpClient
}

func NewWebhookHandler(clickUpClient *services.ClickUpClient) *WebhookHandler {
	return &WebhookHandler{
		clickUpClient: clickUpClient,
	}
}

// HandlePrismaWebhook processes incoming Prisma Cloud webhook alerts
func (h *WebhookHandler) HandlePrismaWebhook(c *fiber.Ctx) error {
	// Log the incoming request
	log.Printf("Received webhook from %s", c.IP())

	// Parse the request body
	var alerts []models.PrismaAlert
	var singleAlert models.PrismaAlert

	// Try to parse as array first
	if err := c.BodyParser(&alerts); err != nil {
		// If array parsing fails, try single object
		if err := c.BodyParser(&singleAlert); err != nil {
			log.Printf("Failed to parse webhook payload: %v", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}
		// If single object parsed successfully, add it to alerts slice
		alerts = []models.PrismaAlert{singleAlert}
	}

	// If no alerts received
	if len(alerts) == 0 {
		log.Println("No alerts in webhook payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No alerts in payload",
		})
	}

	log.Printf("Processing %d alert(s)", len(alerts))

	// Create ClickUp task for each alert
	var createdTasks []string
	var errors []string

	for i, alert := range alerts {
		log.Printf("Processing alert %d: %s (Severity: %s)", i+1, alert.PolicyName, alert.Severity)

		task, err := h.clickUpClient.CreateTask(&alert)
		if err != nil {
			errMsg := "Failed to create task for alert: " + err.Error()
			log.Printf("Error for alert %d: %s", i+1, errMsg)
			errors = append(errors, errMsg)
			continue
		}

		log.Printf("Created ClickUp task: %s (ID: %s)", task.Name, task.ID)
		createdTasks = append(createdTasks, task.ID)
	}

	// Build response
	response := fiber.Map{
		"received":      len(alerts),
		"tasks_created": len(createdTasks),
		"task_ids":      createdTasks,
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["status"] = "partial_success"
	} else {
		response["status"] = "success"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
