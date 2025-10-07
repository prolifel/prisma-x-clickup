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

	return &Config{
		Port:             port,
		ClickUpAPIToken:  clickUpToken,
		ClickUpListID:    clickUpListID,
		ClickUpAssignees: assignees,
	}
}
