package handlers

import (
	"github.com/gofiber/fiber/v2/log"

	"prisma-webhook/models"
	"prisma-webhook/services"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type WebhookHandler struct {
	clickUpClient *services.ClickUpClient
	teamsClient   *services.TeamsClient
}

func NewWebhookHandler(
	clickUpClient *services.ClickUpClient,
	teamsClient *services.TeamsClient,
) *WebhookHandler {
	return &WebhookHandler{
		clickUpClient: clickUpClient,
		teamsClient:   teamsClient,
	}
}

// HandlePrismaWebhook processes incoming Prisma Cloud webhook alerts
func (h *WebhookHandler) HandlePrismaWebhook(c *fiber.Ctx) error {
	// Log the incoming request
	log.Infof("Received webhook from %s", c.IP())
	log.Infof("Payload: %v", string(c.Request().Body()))
	log.Infof("Received type: %s", c.Get("X-Type"))

	// Parse the request body
	var alerts []models.CustomPrismaAlert
	var singleAlert models.CustomPrismaAlert

	// Try to parse as array first
	if err := c.BodyParser(&alerts); err != nil {
		// If array parsing fails, try single object
		if err := c.BodyParser(&singleAlert); err != nil {
			log.Infof("Failed to parse webhook payload: %v, request: %v", err, string(c.Body()))
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request payload",
			})
		}
		// If single object parsed successfully, add it to alerts slice
		alerts = []models.CustomPrismaAlert{singleAlert}
	}

	// If no alerts received
	if len(alerts) == 0 {
		log.Info("No alerts in webhook payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No alerts in payload",
		})
	}

	log.Infof("Processing %d alert(s)", len(alerts))

	// Create ClickUp task for each alert
	var createdTasks []string
	var errors []string
	var isTestMessage bool
	var sharepointPages []string
	var teamsNotifications []string

	for i, alert := range alerts {
		if strings.HasPrefix(alert.Message, "This is a test message from Prisma Cloud initiated") {
			isTestMessage = true
			break
		}

		log.Infof("Processing alert %d: %s (Severity: %s)", i+1, alert.PolicyName, alert.Severity)

		// Step 1: Create ClickUp task
		task, err := h.clickUpClient.CreateTask(&alert, c.Get("X-Type"))
		if err != nil {
			errMsg := "Failed to create task for alert: " + err.Error()
			log.Infof("Error for alert %d: %s", i+1, errMsg)
			errors = append(errors, errMsg)
			continue
		}

		log.Infof("Created ClickUp task: %s (ID: %s)", task.Name, task.ID)
		createdTasks = append(createdTasks, task.ID)

		clickupURL := task.URL

		prismaURL := ""
		if alert.CallbackUrl != "" {
			prismaURL = alert.CallbackUrl
		}
		// prismaURL := "https://app.id.prismacloud.io/alerts/overview?viewId=default&filters={\"alert.id\":[\"" + alert.AlertID + "\"]}\n"

		// Step 2: Send Teams notification (if enabled)
		if h.teamsClient.IsEnabled() {
			err = h.teamsClient.SendTeamsNotificationV2(&alert, clickupURL, prismaURL, c.Get("X-Type"))
			if err != nil {
				errMsg := "Failed to send Teams notification: " + err.Error()
				log.Infof("Warning for alert %d: %s", i+1, errMsg)
				errors = append(errors, errMsg)
			} else {
				log.Infof("Sent Teams notification for alert %d", i+1)
				teamsNotifications = append(teamsNotifications, "sent")
			}
		}
	}

	if isTestMessage {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Test webhook received",
		})
	}

	// Build response
	response := fiber.Map{
		"received":      len(alerts),
		"tasks_created": len(createdTasks),
		"task_ids":      createdTasks,
	}

	if len(sharepointPages) > 0 {
		response["sharepoint_pages_created"] = len(sharepointPages)
		response["sharepoint_urls"] = sharepointPages
	}

	if len(teamsNotifications) > 0 {
		response["teams_notifications_sent"] = len(teamsNotifications)
	}

	if len(errors) > 0 {
		response["errors"] = errors
		response["status"] = "partial_success"
	} else {
		response["status"] = "success"
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
