package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prisma-webhook/config"
	"prisma-webhook/models"
	"strconv"
	"strings"
	"time"
)

type TeamsClient struct {
	webhookURL string
}

// Adaptive Card structures for Power Automate
type teamsAdaptiveCardMessage struct {
	Type        string                        `json:"type"`
	Attachments []teamsAdaptiveCardAttachment `json:"attachments"`
}

type teamsAdaptiveCardAttachment struct {
	ContentType string                   `json:"contentType"`
	ContentURL  interface{}              `json:"contentUrl"`
	Content     teamsAdaptiveCardContent `json:"content"`
}

type teamsAdaptiveCardContent struct {
	Schema  string                     `json:"$schema"`
	Type    string                     `json:"type"`
	Version string                     `json:"version"`
	Body    []teamsAdaptiveCardElement `json:"body"`
	Actions []teamsAdaptiveCardAction  `json:"actions,omitempty"`
}

type teamsAdaptiveCardElement struct {
	Type      string                    `json:"type"`
	Text      string                    `json:"text,omitempty"`
	Size      string                    `json:"size,omitempty"`
	Weight    string                    `json:"weight,omitempty"`
	Color     string                    `json:"color,omitempty"`
	Wrap      bool                      `json:"wrap,omitempty"`
	Separator bool                      `json:"separator,omitempty"`
	Spacing   string                    `json:"spacing,omitempty"`
	Columns   []teamsAdaptiveCardColumn `json:"columns,omitempty"`
}

type teamsAdaptiveCardColumn struct {
	Type  string                     `json:"type"`
	Width string                     `json:"width,omitempty"`
	Items []teamsAdaptiveCardElement `json:"items"`
}

type teamsAdaptiveCardAction struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

// Legacy MessageCard structures (kept for backwards compatibility)
type teamsMessageCard struct {
	Type            string                 `json:"@type"`
	Context         string                 `json:"@context"`
	ThemeColor      string                 `json:"themeColor"`
	Summary         string                 `json:"summary"`
	Sections        []teamsSection         `json:"sections"`
	PotentialAction []teamsPotentialAction `json:"potentialAction,omitempty"`
}

type teamsSection struct {
	ActivityTitle    string      `json:"activityTitle,omitempty"`
	ActivitySubtitle string      `json:"activitySubtitle,omitempty"`
	ActivityImage    string      `json:"activityImage,omitempty"`
	Facts            []teamsFact `json:"facts,omitempty"`
	Markdown         bool        `json:"markdown,omitempty"`
	Text             string      `json:"text,omitempty"`
}

type teamsFact struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type teamsPotentialAction struct {
	Type    string              `json:"@type"`
	Name    string              `json:"name"`
	Targets []teamsActionTarget `json:"targets,omitempty"`
}

type teamsActionTarget struct {
	OS  string `json:"os"`
	URI string `json:"uri"`
}

func NewTeamsClient(cfg *config.Config) *TeamsClient {
	return &TeamsClient{
		webhookURL: cfg.TeamsWebhookURL,
	}
}

// IsEnabled returns true if the TeamsClient is properly configured
func (t *TeamsClient) IsEnabled() bool {
	return t.webhookURL != ""
}

// SendTeamsNotification sends an Adaptive Card notification to Microsoft Teams via webhook (Power Automate)
func (t *TeamsClient) SendTeamsNotification(alert *models.PrismaAlert, clickupURL string, sharepointURL string) error {
	if !t.IsEnabled() {
		return fmt.Errorf("Teams client is not properly configured")
	}

	// Extract alert details with fallbacks
	severity := alert.Policy.Severity
	if severity == "" {
		severity = alert.Severity
	}

	policyName := alert.Policy.Name
	if policyName == "" {
		policyName = alert.PolicyName
	}

	resourceName := alert.Resource.ResourceName
	if resourceName == "" {
		resourceName = alert.ResourceName
	}

	accountName := alert.Account.Name
	if accountName == "" {
		accountName = alert.AccountName
	}

	region := alert.Region
	if region == "" {
		region = alert.ResourceRegion
	}

	cloudType := alert.Account.CloudType
	if cloudType == "" {
		cloudType = alert.CloudType
	}

	// Determine severity color
	severityColor := t.getSeverityColorName(severity)

	// parse alertTs
	i, err := strconv.ParseInt(alert.AlertTs, 10, 64)
	if err != nil {
		return fmt.Errorf("Failed to convert alert timestamp to int: %v", err)
	}
	alertTime := time.UnixMilli(i)

	// Build Adaptive Card body
	body := []teamsAdaptiveCardElement{
		{
			Type:   "TextBlock",
			Text:   "🔔 Prisma Cloud Security Alert",
			Size:   "Large",
			Weight: "Bolder",
			Wrap:   true,
		},
		{
			Type:      "TextBlock",
			Text:      fmt.Sprintf("**Severity:** %s", strings.ToUpper(severity)),
			Color:     severityColor,
			Size:      "Medium",
			Weight:    "Bolder",
			Wrap:      true,
			Separator: true,
		},
		{
			Type:    "TextBlock",
			Text:    "**Policy Violation Details**",
			Weight:  "Bolder",
			Spacing: "Medium",
			Wrap:    true,
		},
		// Policy Name
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Policy:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: policyName,
							Wrap: true,
						},
					},
				},
			},
		},
		// Resource Name
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Resource:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: resourceName,
							Wrap: true,
						},
					},
				},
			},
		},
		// Account
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Account:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: accountName,
							Wrap: true,
						},
					},
				},
			},
		},
		// Cloud Type
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Cloud:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: cloudType,
							Wrap: true,
						},
					},
				},
			},
		},
		// Region
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Region:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: region,
							Wrap: true,
						},
					},
				},
			},
		},
		// Alert Time
		{
			Type: "ColumnSet",
			Columns: []teamsAdaptiveCardColumn{
				{
					Type:  "Column",
					Width: "auto",
					Items: []teamsAdaptiveCardElement{
						{
							Type:   "TextBlock",
							Text:   "**Alert Time:**",
							Weight: "Bolder",
							Wrap:   true,
						},
					},
				},
				{
					Type:  "Column",
					Width: "stretch",
					Items: []teamsAdaptiveCardElement{
						{
							Type: "TextBlock",
							Text: alertTime.Format("2006-01-02 15:04:05 +0700"),
							Wrap: true,
						},
					},
				},
			},
		},
	}

	// Build actions
	var actions []teamsAdaptiveCardAction
	if clickupURL != "" {
		actions = append(actions, teamsAdaptiveCardAction{
			Type:  "Action.OpenUrl",
			Title: "View ClickUp Task",
			URL:   clickupURL,
		})
	}

	if sharepointURL != "" {
		actions = append(actions, teamsAdaptiveCardAction{
			Type:  "Action.OpenUrl",
			Title: "View Documentation",
			URL:   sharepointURL,
		})
	}

	// Build the Adaptive Card
	adaptiveCard := teamsAdaptiveCardMessage{
		Type: "message",
		Attachments: []teamsAdaptiveCardAttachment{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				ContentURL:  nil,
				Content: teamsAdaptiveCardContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.4",
					Body:    body,
					Actions: actions,
				},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(adaptiveCard)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams adaptive card: %w", err)
	}

	// Send the webhook
	req, err := http.NewRequest("POST", t.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create Teams webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Teams webhook: %w", err)
	}
	defer resp.Body.Close()

	body2, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read Teams response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Teams webhook failed (status %d): %s", resp.StatusCode, string(body2))
	}

	return nil
}

// getSeverityColor returns a hex color for the severity level (legacy MessageCard)
func (t *TeamsClient) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "8B0000" // Dark red
	case "high":
		return "D32F2F" // Red
	case "medium":
		return "F57C00" // Orange
	case "low":
		return "FBC02D" // Yellow
	default:
		return "757575" // Gray
	}
}

// getSeverityColorName returns an Adaptive Card color name for the severity level
func (t *TeamsClient) getSeverityColorName(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "Attention" // Red
	case "high":
		return "Attention" // Red
	case "medium":
		return "Warning" // Orange/Yellow
	case "low":
		return "Good" // Green
	default:
		return "Default" // Default text color
	}
}
