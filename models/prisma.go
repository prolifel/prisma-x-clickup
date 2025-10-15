package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// PrismaAlert represents the webhook payload from Prisma Cloud
type PrismaAlert struct {
	ResourceID        string    `json:"resourceId"`
	AlertRuleName     string    `json:"alertRuleName"`
	AccountName       string    `json:"accountName"`
	CloudType         string    `json:"cloudType"`
	Severity          string    `json:"severity"`
	PolicyName        string    `json:"policyName"`
	ResourceName      string    `json:"resourceName"`
	PolicyDescription string    `json:"policyDescription"`
	AlertID           string    `json:"alertId"`
	AlertStatus       string    `json:"alertStatus"`
	AlertTime         time.Time `json:"alertTime"`
	ResourceRegion    string    `json:"resourceRegion"`
	ResourceType      string    `json:"resourceType"`
	PolicyID          string    `json:"policyId"`
	PolicyType        string    `json:"policyType"`
	RiskRating        string    `json:"riskRating"`
	CallbackURL       string    `json:"callbackUrl"`
	Sender            string    `json:"sender"`
	SentTs            int64     `json:"sentTs"`
	Message           string    `json:"message"`
	Account           Account   `json:"account"`
	Region            string    `json:"region"`
	ResourceRegionId  string    `json:"resourceRegionId"`
	Policy            Policy    `json:"policy"`
	Resource          Resource  `json:"resource"`
	Metadata          Metadata  `json:"metadata"`
	Reason            string    `json:"reason"`
	AlertRuleId       string    `json:"alertRuleId"`
	AlertTs           string    `json:"alertTs"`
	Firstseen         string    `json:"firstSeen"`
	Lastseen          string    `json:"lastSeen"`
	Service           string    `json:"service"`
}

type CustomPrismaAlert struct {
	Message                        string      `json:"message"`
	ResourceId                     string      `json:"resourceId"`
	AlertRuleName                  string      `json:"alertRuleName"`
	Anomaly                        fiber.Map   `json:"anomaly"`
	AccountName                    string      `json:"accountName"`
	HasFinding                     bool        `json:"hasFinding"`
	ResourceRegionId               string      `json:"resourceRegionId"`
	AlertRemediationCli            string      `json:"alertRemediationCli"`
	AlertRemediationCliDescription string      `json:"alertRemediationCliDescription"`
	AlertRemediationImpact         string      `json:"alertRemediationImpact"`
	Source                         string      `json:"source"`
	CloudType                      string      `json:"cloudType"`
	ComplianceMetadata             []fiber.Map `json:"complianceMetadata"`
	CallbackUrl                    string      `json:"callbackUrl"`
	AlertId                        string      `json:"alertId"`
	PolicyLabels                   []string    `json:"policyLabels"`
	AlertAttribution               fiber.Map   `json:"alertAttribution"`
	Severity                       string      `json:"severity"`
	PolicyName                     string      `json:"policyName"`
	Resource                       fiber.Map   `json:"resource"`
	ResourceName                   string      `json:"resourceName"`
	ResourceRegion                 string      `json:"resourceRegion"`
	PolicyDescription              string      `json:"policyDescription"`
	PolicyRecommendation           string      `json:"policyRecommendation"`
	AccountId                      string      `json:"accountId"`
	PolicyId                       string      `json:"policyId"`
	ResourceCloudService           string      `json:"resourceCloudService"`
	AlertTs                        int64       `json:"alertTs"`
	FirstSeen                      int64       `json:"firstSeen"`
	LastSeen                       int64       `json:"lastSeen"`
	ResourceType                   string      `json:"resourceType"`
	AdditionalInfo                 fiber.Map   `json:"additionalInfo"`
	Reason                         string      `json:"reason"`
	AlertStatus                    string      `json:"alertStatus"`
	AlertDismissalNote             string      `json:"alertDismissalNote"`
	AlertRuleId                    string      `json:"alertRuleId"`
	Tags                           []fiber.Map `json:"tags"`
	FindingSummary                 fiber.Map   `json:"findingSummary"`
	PolicyType                     string      `json:"policyType"`
	AccountOwners                  string      `json:"accountOwners"`
	AccountAncestors               string      `json:"accountAncestors"`
}

type Account struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	CloudType string `json:"cloudType"`
}

type Policy struct {
	Name           string   `json:"name"`
	Id             string   `json:"id"`
	Description    string   `json:"description"`
	Recommendation string   `json:"recommendation"`
	Severity       string   `json:"severity"`
	Labels         []string `json:"labels"`
	PolicyTs       string   `json:"policyTs"`
	PolicyType     string   `json:"policyType"`
}

type Resource struct {
	ResourceId   string `json:"resourceId"`
	ResourceName string `json:"resourceName"`
	ResourceTs   string `json:"resourceTs"`
}

type Metadata struct {
	Remediable bool `json:"remediable"`
}

// GetPriority maps Prisma Cloud severity to ClickUp priority
func (p *PrismaAlert) GetPriority() int {
	switch p.Policy.Severity {
	case "high", "critical":
		return 1 // Urgent
	case "medium":
		return 2 // High
	case "low":
		return 3 // Normal
	default:
		return 4 // Low
	}
}

func (p *CustomPrismaAlert) GetPriority() int {
	switch p.Severity {
	case "high", "critical":
		return 1 // Urgent
	case "medium":
		return 2 // High
	case "low":
		return 3 // Normal
	default:
		return 4 // Low
	}
}

// GetTaskTitle generates a task title from the alert
func (p *PrismaAlert) GetTaskTitle() string {
	if p.PolicyName != "" {
		return fmt.Sprintf("[%s] - %s", strings.ToUpper(p.Severity), p.PolicyName)
	}
	if p.Policy.Name != "" {
		return fmt.Sprintf("[%s] - %s", strings.ToUpper(p.Policy.Severity), p.Policy.Name)
	}
	return "[Prisma Cloud] Security Alert"
}

func (p *CustomPrismaAlert) GetTaskTitle() string {
	if p.PolicyName != "" {
		return fmt.Sprintf("[%s] - %s", strings.ToUpper(p.Severity), p.PolicyName)
	}
	// if p.Policy.Name != "" {
	// 	return fmt.Sprintf("[%s] - %s", strings.ToUpper(p.Policy.Severity), p.Policy.Name)
	// }
	return "[Prisma Cloud] Security Alert"
}

// GetTaskDescription generates a detailed task description from the alert
func (p *PrismaAlert) GetTaskDescription() string {
	desc := "\n\n"

	if p.PolicyDescription != "" {
		desc += "Description: " + p.PolicyDescription + "\n\n"
	}

	if p.Policy.Description != "" {
		desc += "Description: " + p.Policy.Description + "\n\n"
	}

	if p.AlertID != "" {
		desc += "Alert ID: " + p.AlertID + "\n\n"
	}

	if p.Policy.Severity != "" {
		desc += "Severity: " + p.Policy.Severity + "\n\n"
	}

	if p.PolicyName != "" {
		desc += "Policy: " + p.PolicyName + "\n\n"
	}

	if p.Policy.Name != "" {
		desc += "Policy: " + p.Policy.Name + "\n\n"
	}

	if p.Policy.Recommendation != "" {
		desc += "Recommendation: " + p.Policy.Recommendation + "\n\n"
	}

	if p.ResourceName != "" {
		desc += "Resource: " + p.ResourceName + "\n\n"
	}

	if p.Resource.ResourceName != "" {
		desc += "Resource: " + p.Resource.ResourceName + "\n\n"
	}

	if p.AccountName != "" {
		desc += "Account: " + p.AccountName + "\n\n"
	}

	if p.Account.Name != "" {
		desc += "Account: " + p.Account.Name + "\n\n"
	}

	if p.CloudType != "" {
		desc += "Cloud: " + p.CloudType + "\n\n"
	}

	if p.Account.CloudType != "" {
		desc += "Cloud: " + p.Account.CloudType + "\n\n"
	}

	if p.ResourceRegion != "" {
		desc += "Region: " + p.ResourceRegion + "\n\n"
	}

	if p.Region != "" {
		desc += "Region: " + p.Region + "\n\n"
	}

	return desc
}

func (p *PrismaAlert) GetTaskDescriptionV2() string {
	desc := "# Prisma Cloud Alert Summary\n"
	desc += "## Alerts Detail\n"
	desc += "| **Field** | **Detail** |\n"
	desc += "| ------ | ------ |\n"

	if p.AlertID != "" {
		desc += "| **Alert ID** | " + p.AlertID + " |\n"
	}

	if p.Policy.Name != "" {
		desc += "| **Policy Name** | " + p.Policy.Name + " |\n"
	}

	if p.Policy.PolicyType != "" {
		desc += "| **Policy Type** | " + p.Policy.PolicyType + " |\n"
	}

	if p.Policy.Severity != "" {
		desc += "| **Severity** | " + p.getSeverityColor(p.Policy.Severity) + " |\n"
	}

	if p.Account.Name != "" {
		desc += "| **Cloud Account** | " + p.Account.Name + " |\n"
	}

	if p.Resource.ResourceName != "" {
		desc += "| **Resource Type** | " + p.Resource.ResourceName + " |\n"
	}

	if p.Region != "" {
		desc += "| **Region** | " + p.Region + " |\n"
	}

	if p.AlertStatus != "" {
		desc += "| **Status** | " + p.AlertStatus + " |\n"
	}

	desc += "---\n"

	desc += "## Description\n"
	desc += p.Policy.Description + "\n"

	desc += "## Remediation Recommendation\n"
	desc += p.Policy.Recommendation + "\n"

	desc += "---\n"

	desc += "[View Alert on Prisma](https://app.id.prismacloud.io/alerts/overview?viewId=default&filters={\"alert.id\":[\"" + p.AlertID + "\"]})\n"

	return desc
}

func (p *CustomPrismaAlert) GetTaskDescriptionV2() string {
	desc := "# Prisma Cloud Alert Summary\n"
	desc += "## Alerts Detail\n"
	desc += "| **Field** | **Detail** |\n"
	desc += "| ------ | ------ |\n"

	if p.AlertId != "" {
		desc += "| **Alert ID** | " + p.AlertId + " |\n"
	}

	if p.AlertRuleId != "" {
		desc += "| **Alert Rule ID** | " + p.AlertRuleId + " |\n"
	}

	if p.AlertRuleName != "" {
		desc += "| **Alert Rule Name** | " + p.AlertRuleName + " |\n"
	}

	if p.PolicyName != "" {
		desc += "| **Policy Name** | " + p.PolicyName + " |\n"
	}

	if p.PolicyType != "" {
		desc += "| **Policy Type** | " + p.PolicyType + " |\n"
	}

	if p.Severity != "" {
		desc += "| **Severity** | " + p.getSeverityColor(p.Severity) + " |\n"
	}

	if p.CloudType != "" {
		desc += "| **Cloud Provider** | " + p.CloudType + " |\n"
	}

	if p.AccountName != "" {
		desc += "| **Cloud Account** | " + p.AccountName + " |\n"
	}

	if p.ResourceId != "" {
		desc += "| **Resource ID** | " + p.ResourceId + " |\n"
	}

	if p.ResourceName != "" {
		desc += "| **Resource Name** | " + p.ResourceName + " |\n"
	}

	if p.ResourceCloudService != "" {
		desc += "| **Resource Cloud Service** | " + p.ResourceCloudService + " |\n"
	}

	if p.ResourceType != "" {
		desc += "| **Resource Type** | " + p.ResourceType + " |\n"
	}

	if p.ResourceRegion != "" {
		desc += "| **Region** | " + p.ResourceRegion + " |\n"
	}

	if p.AlertStatus != "" {
		desc += "| **Status** | " + p.AlertStatus + " |\n"
	}

	desc += "---\n"

	desc += "## Description\n"
	desc += p.PolicyDescription + "\n"

	desc += "## Remediation Recommendation\n"
	desc += p.PolicyRecommendation + "\n"

	desc += "---\n"

	if p.Tags != nil {
		desc += "## Tags\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.Tags)
		desc += "```\n"
		desc += "---\n"
	}

	if p.FindingSummary != nil {
		desc += "## Finding Summary\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.FindingSummary)
		desc += "```\n"
		desc += "---\n"
	}

	if p.Anomaly != nil {
		desc += "## Anomaly\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.Anomaly)
		desc += "```\n"
		desc += "---\n"
	}

	if p.AlertRemediationCli != "" {
		desc += "## Remediation via CLI\n"

		if p.AlertRemediationCliDescription != "" {
			desc += p.AlertRemediationCliDescription
		}

		if p.AlertRemediationImpact != "" {
			desc += p.AlertRemediationImpact
		}

		desc += "\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.AlertRemediationCli)
		desc += "```\n"
		desc += "---\n"
	}

	if p.ComplianceMetadata != nil {
		desc += "## Compliance\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.ComplianceMetadata)
		desc += "```\n"
		desc += "---\n"
	}

	if p.AlertAttribution != nil {
		desc += "## Alert Attribution\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.AlertAttribution)
		desc += "```\n"
		desc += "---\n"
	}

	if p.Resource != nil {
		desc += "## Resource\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.Resource)
		desc += "```\n"
		desc += "---\n"
	}

	if p.AdditionalInfo != nil {
		desc += "## Additional Info\n"
		desc += "```json\n"
		desc += fmt.Sprintf("%v", p.AdditionalInfo)
		desc += "```\n"
		desc += "---\n"
	}

	// desc += "[View Alert on Prisma](https://app.id.prismacloud.io/alerts/overview?viewId=default&filters={\"alert.id\":[\"" + p.AlertID + "\"]})\n"
	desc += "[View Alert on Prisma](" + p.CallbackUrl + ")\n"

	return desc
}

func (p *PrismaAlert) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴 Critical" // Red
	case "high":
		return "🟠 High" // Red
	case "medium":
		return "🟡 Medium" // Orange/Yellow
	case "low":
		return "🟢 Low" // Green
	default:
		return "" // Default text color
	}
}

func (p *CustomPrismaAlert) getSeverityColor(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "🔴 Critical" // Red
	case "high":
		return "🟠 High" // Red
	case "medium":
		return "🟡 Medium" // Orange/Yellow
	case "low":
		return "🟢 Low" // Green
	default:
		return "" // Default text color
	}
}
