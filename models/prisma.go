package models

import (
	"fmt"
	"strings"
	"time"
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
