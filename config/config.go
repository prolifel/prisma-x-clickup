package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port             string
	ClickUpAPIToken  string
	ClickUpListID    string
	ClickUpAssignees []int
	WebhookAPIKey    string
	AllowedIPs       []string

	// Azure AD / Microsoft Graph
	AzureTenantID     string
	AzureClientID     string
	AzureClientSecret string
	SharePointSiteID  string

	// Microsoft Teams
	TeamsWebhookURL string
}

func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	clickUpToken := os.Getenv("CLICKUP_API_TOKEN")
	if clickUpToken == "" {
		log.Fatal("CLICKUP_API_TOKEN is required")
	}

	clickUpListID := os.Getenv("CLICKUP_LIST_ID")
	if clickUpListID == "" {
		log.Fatal("CLICKUP_LIST_ID is required")
	}

	assigneesStr := os.Getenv("CLICKUP_ASSIGNEES")
	var assignees []int
	if assigneesStr != "" {
		assigneeStrs := strings.Split(assigneesStr, ",")
		for _, idStr := range assigneeStrs {
			id, err := strconv.Atoi(strings.TrimSpace(idStr))
			if err != nil {
				log.Printf("Warning: Invalid assignee ID '%s', skipping", idStr)
				continue
			}
			assignees = append(assignees, id)
		}
	}

	webhookAPIKey := os.Getenv("WEBHOOK_API_KEY")
	if webhookAPIKey == "" {
		log.Fatal("WEBHOOK_API_KEY is required")
	}

	var allowedIPs []string
	allowedIPsStr := os.Getenv("ALLOWED_IPS")
	if allowedIPsStr != "" {
		ipStrs := strings.Split(allowedIPsStr, ",")
		for _, ip := range ipStrs {
			allowedIPs = append(allowedIPs, strings.TrimSpace(ip))
		}
		log.Printf("IP allowlist enabled with %d IP(s)", len(allowedIPs))
	} else {
		log.Println("Warning: No IP allowlist configured. All IPs will be allowed.")
	}

	// Azure AD / Microsoft Graph (optional)
	azureTenantID := os.Getenv("AZURE_TENANT_ID")
	azureClientID := os.Getenv("AZURE_CLIENT_ID")
	azureClientSecret := os.Getenv("AZURE_CLIENT_SECRET")
	sharePointSiteID := os.Getenv("SHAREPOINT_SITE_ID")

	if azureTenantID != "" && azureClientID != "" && azureClientSecret != "" && sharePointSiteID != "" {
		log.Println("SharePoint integration enabled")
	} else if azureTenantID != "" || azureClientID != "" || azureClientSecret != "" || sharePointSiteID != "" {
		log.Println("Warning: SharePoint integration partially configured. All Azure AD fields are required.")
	}

	// Microsoft Teams Webhook (optional)
	teamsWebhookURL := os.Getenv("TEAMS_WEBHOOK_URL")
	if teamsWebhookURL != "" {
		log.Println("Teams webhook integration enabled")
	}

	return &Config{
		Port:              port,
		ClickUpAPIToken:   clickUpToken,
		ClickUpListID:     clickUpListID,
		ClickUpAssignees:  assignees,
		WebhookAPIKey:     webhookAPIKey,
		AllowedIPs:        allowedIPs,
		AzureTenantID:     azureTenantID,
		AzureClientID:     azureClientID,
		AzureClientSecret: azureClientSecret,
		SharePointSiteID:  sharePointSiteID,
		TeamsWebhookURL:   teamsWebhookURL,
	}
}
