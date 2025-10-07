package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prisma-webhook/config"
	"prisma-webhook/models"
)

type ClickUpClient struct {
	apiToken  string
	listID    string
	assignees []int
}

type CreateTaskRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Assignees   []int  `json:"assignees,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Status      string `json:"status,omitempty"`
}

type CreateTaskResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Status string `json:"status"`
	} `json:"status"`
	URL string `json:"url"`
}

func NewClickUpClient(cfg *config.Config) *ClickUpClient {
	return &ClickUpClient{
		apiToken:  cfg.ClickUpAPIToken,
		listID:    cfg.ClickUpListID,
		assignees: cfg.ClickUpAssignees,
	}
}

func (c *ClickUpClient) CreateTask(alert *models.PrismaAlert) (*CreateTaskResponse, error) {
	url := fmt.Sprintf("https://api.clickup.com/api/v2/list/%s/task", c.listID)

	taskReq := CreateTaskRequest{
		Name:        alert.GetTaskTitle(),
		Description: alert.GetTaskDescription(),
		Assignees:   c.assignees,
		Priority:    alert.GetPriority(),
		Status:      "Open",
	}

	jsonData, err := json.Marshal(taskReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal task request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ClickUp API error (status %d): %s", resp.StatusCode, string(body))
	}

	var taskResp CreateTaskResponse
	if err := json.Unmarshal(body, &taskResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &taskResp, nil
}
