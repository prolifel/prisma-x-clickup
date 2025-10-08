package models

import "time"

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
	SentTs            string    `json:"sentTs"`
	Message           string    `json:"message"`
}

// GetPriority maps Prisma Cloud severity to ClickUp priority
func (p *PrismaAlert) GetPriority() int {
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
		return "[Prisma Cloud] " + p.PolicyName
	}
	return "[Prisma Cloud] Security Alert"
}

// GetTaskDescription generates a detailed task description from the alert
func (p *PrismaAlert) GetTaskDescription() string {
	desc := "## Prisma Cloud Security Alert\n\n"

	if p.PolicyName != "" {
		desc += "**Policy:** " + p.PolicyName + "\n\n"
	}

	if p.PolicyDescription != "" {
		desc += "**Description:** " + p.PolicyDescription + "\n\n"
	}

	desc += "**Severity:** " + p.Severity + "\n\n"

	if p.ResourceName != "" {
		desc += "**Resource:** " + p.ResourceName + "\n\n"
	}

	if p.ResourceType != "" {
		desc += "**Resource Type:** " + p.ResourceType + "\n\n"
	}

	if p.AccountName != "" {
		desc += "**Account:** " + p.AccountName + "\n\n"
	}

	if p.CloudType != "" {
		desc += "**Cloud:** " + p.CloudType + "\n\n"
	}

	if p.ResourceRegion != "" {
		desc += "**Region:** " + p.ResourceRegion + "\n\n"
	}

	if p.AlertID != "" {
		desc += "**Alert ID:** " + p.AlertID + "\n\n"
	}

	if p.CallbackURL != "" {
		desc += "**View in Prisma Cloud:** " + p.CallbackURL + "\n\n"
	}

	return desc
}
