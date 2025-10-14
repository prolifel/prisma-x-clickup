package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"prisma-webhook/config"
	"prisma-webhook/models"
	"strings"
	"time"
)

type MSGraphClient struct {
	tenantID     string
	clientID     string
	clientSecret string
	siteID       string
	accessToken  string
	tokenExpiry  time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type createPageRequest struct {
	Name              string              `json:"name"`
	Title             string              `json:"title"`
	PageLayout        string              `json:"pageLayout"`
	PromotionKind     string              `json:"promotionKind,omitempty"`
	CanvasLayout      canvasLayout        `json:"canvasLayout"`
	TitleArea         *titleArea          `json:"titleArea,omitempty"`
}

type canvasLayout struct {
	HorizontalSections []horizontalSection `json:"horizontalSections"`
}

type horizontalSection struct {
	Layout  string   `json:"layout"`
	Columns []column `json:"columns"`
}

type column struct {
	Width    int        `json:"width"`
	WebParts []webPart  `json:"webParts"`
}

type webPart struct {
	ID         string                 `json:"id"`
	InnerHTML  string                 `json:"innerHtml"`
	WebPartType string                `json:"webPartType,omitempty"`
}

type titleArea struct {
	EnableGradientEffect bool   `json:"enableGradientEffect"`
	ImageWebUrl          string `json:"imageWebUrl,omitempty"`
	Layout               string `json:"layout"`
	ShowAuthor           bool   `json:"showAuthor"`
	ShowPublishedDate    bool   `json:"showPublishedDate"`
	ShowTextBlockAboveTitle bool `json:"showTextBlockAboveTitle"`
	TextAboveTitle       string `json:"textAboveTitle,omitempty"`
	TextAlignment        string `json:"textAlignment"`
}

type createPageResponse struct {
	ID      string `json:"id"`
	WebUrl  string `json:"webUrl"`
	Name    string `json:"name"`
	Title   string `json:"title"`
}

func NewMSGraphClient(cfg *config.Config) *MSGraphClient {
	return &MSGraphClient{
		tenantID:     cfg.AzureTenantID,
		clientID:     cfg.AzureClientID,
		clientSecret: cfg.AzureClientSecret,
		siteID:       cfg.SharePointSiteID,
	}
}

// IsEnabled returns true if the MSGraphClient is properly configured
func (c *MSGraphClient) IsEnabled() bool {
	return c.tenantID != "" && c.clientID != "" && c.clientSecret != "" && c.siteID != ""
}

// getAccessToken retrieves an access token using client credentials flow
func (c *MSGraphClient) getAccessToken() error {
	// Check if we have a valid token
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	tokenURL := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", c.tenantID)

	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("token request failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("failed to unmarshal token response: %w", err)
	}

	c.accessToken = tokenResp.AccessToken
	// Set expiry with 5 minute buffer
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-300) * time.Second)

	return nil
}

// CreateSharePointPage creates a documentation page in SharePoint with alert details
func (c *MSGraphClient) CreateSharePointPage(alert *models.PrismaAlert, clickupURL string) (string, error) {
	if !c.IsEnabled() {
		return "", fmt.Errorf("MSGraph client is not properly configured")
	}

	// Get access token
	if err := c.getAccessToken(); err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Build the page content
	pageTitle := alert.GetTaskTitle()
	pageName := c.sanitizePageName(pageTitle)

	// Create HTML content
	htmlContent := c.buildPageContent(alert, clickupURL)

	// Create page request
	pageReq := createPageRequest{
		Name:       pageName,
		Title:      pageTitle,
		PageLayout: "article",
		CanvasLayout: canvasLayout{
			HorizontalSections: []horizontalSection{
				{
					Layout: "oneColumn",
					Columns: []column{
						{
							Width: 12,
							WebParts: []webPart{
								{
									ID:        "1",
									InnerHTML: htmlContent,
								},
							},
						},
					},
				},
			},
		},
	}

	jsonData, err := json.Marshal(pageReq)
	if err != nil {
		return "", fmt.Errorf("failed to marshal page request: %w", err)
	}

	// Create the page
	apiURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/sites/%s/pages", c.siteID)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create page request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("page creation failed (status %d): %s", resp.StatusCode, string(body))
	}

	var pageResp createPageResponse
	if err := json.Unmarshal(body, &pageResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal page response: %w", err)
	}

	// Publish the page
	if err := c.publishPage(pageResp.ID); err != nil {
		// Log but don't fail if publish fails
		fmt.Printf("Warning: Failed to publish page: %v\n", err)
	}

	return pageResp.WebUrl, nil
}

// publishPage publishes a SharePoint page
func (c *MSGraphClient) publishPage(pageID string) error {
	apiURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/sites/%s/pages/%s/publish", c.siteID, pageID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create publish request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to publish page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// sanitizePageName creates a safe page name from the title
func (c *MSGraphClient) sanitizePageName(title string) string {
	// Remove special characters and replace spaces with dashes
	name := strings.ToLower(title)
	name = strings.ReplaceAll(name, " ", "-")
	// Remove brackets and other special chars
	name = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, name)

	// Limit length and add timestamp to ensure uniqueness
	if len(name) > 50 {
		name = name[:50]
	}

	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.aspx", name, timestamp)
}

// buildPageContent creates HTML content for the SharePoint page
func (c *MSGraphClient) buildPageContent(alert *models.PrismaAlert, clickupURL string) string {
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

	timestamp := alert.AlertTime.Format("2006-01-02 15:04:05 MST")

	policyDesc := alert.Policy.Description
	if policyDesc == "" {
		policyDesc = alert.PolicyDescription
	}

	recommendation := alert.Policy.Recommendation

	// Build HTML content
	var html strings.Builder

	html.WriteString(`<div style="font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;">`)

	// Severity badge
	severityColor := c.getSeverityColor(severity)
	html.WriteString(fmt.Sprintf(`<div style="margin-bottom: 20px;"><span style="background-color: %s; color: white; padding: 5px 15px; border-radius: 3px; font-weight: bold;">%s</span></div>`, severityColor, strings.ToUpper(severity)))

	// Alert details
	html.WriteString(`<h2>Alert Details</h2>`)
	html.WriteString(`<table style="border-collapse: collapse; width: 100%; margin-bottom: 20px;">`)

	c.addTableRow(&html, "Policy Name", policyName)
	c.addTableRow(&html, "Severity", severity)
	c.addTableRow(&html, "Resource Name", resourceName)
	c.addTableRow(&html, "Account", accountName)
	c.addTableRow(&html, "Region", region)
	c.addTableRow(&html, "Alert Time", timestamp)

	if alert.AlertID != "" {
		c.addTableRow(&html, "Alert ID", alert.AlertID)
	}

	html.WriteString(`</table>`)

	// Policy Description
	if policyDesc != "" {
		html.WriteString(`<h2>Description</h2>`)
		html.WriteString(fmt.Sprintf(`<p style="margin-bottom: 20px;">%s</p>`, c.escapeHTML(policyDesc)))
	}

	// Recommendation
	if recommendation != "" {
		html.WriteString(`<h2>Recommendation</h2>`)
		html.WriteString(fmt.Sprintf(`<p style="margin-bottom: 20px;">%s</p>`, c.escapeHTML(recommendation)))
	}

	// Links
	html.WriteString(`<h2>Related Links</h2>`)
	html.WriteString(`<ul>`)
	if clickupURL != "" {
		html.WriteString(fmt.Sprintf(`<li><a href="%s" target="_blank">View ClickUp Task</a></li>`, clickupURL))
	}
	html.WriteString(`</ul>`)

	html.WriteString(`</div>`)

	return html.String()
}

// addTableRow adds a row to the HTML table
func (c *MSGraphClient) addTableRow(html *strings.Builder, label, value string) {
	html.WriteString(`<tr>`)
	html.WriteString(fmt.Sprintf(`<td style="padding: 8px; border: 1px solid #ddd; background-color: #f5f5f5; font-weight: bold; width: 200px;">%s</td>`, label))
	html.WriteString(fmt.Sprintf(`<td style="padding: 8px; border: 1px solid #ddd;">%s</td>`, c.escapeHTML(value)))
	html.WriteString(`</tr>`)
}

// getSeverityColor returns a color for the severity level
func (c *MSGraphClient) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "#8B0000" // Dark red
	case "high":
		return "#D32F2F" // Red
	case "medium":
		return "#F57C00" // Orange
	case "low":
		return "#FBC02D" // Yellow
	default:
		return "#757575" // Gray
	}
}

// escapeHTML escapes HTML special characters
func (c *MSGraphClient) escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
