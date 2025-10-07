# Prisma Cloud to ClickUp Webhook Integration

A lightweight Go Fiber HTTP backend that receives webhook notifications from Prisma Cloud and automatically creates ClickUp tasks with assignees.

## Features

- **Webhook Endpoint**: Receives Prisma Cloud alert webhooks
- **Automatic Task Creation**: Creates ClickUp tasks with detailed information
- **Priority Mapping**: Maps Prisma Cloud severity (high/medium/low) to ClickUp priority
- **Assignee Support**: Automatically assigns tasks to specified team members
- **Dockerized**: Easy deployment with Docker and Docker Compose
- **Health Check**: Built-in health check endpoint for monitoring

## Prerequisites

- Go 1.21 or later (for local development)
- Docker and Docker Compose (for containerized deployment)
- ClickUp API Token
- ClickUp List ID
- Prisma Cloud instance with webhook integration

## Quick Start

### 1. Clone and Configure

```bash
# Copy the example environment file
cp .env.example .env

# Edit .env with your credentials
nano .env
```

### 2. Get ClickUp Credentials

**API Token:**
1. Go to https://app.clickup.com/settings/apps
2. Generate a new API token
3. Copy the token to `CLICKUP_API_TOKEN` in `.env`

**List ID:**
1. Navigate to your ClickUp list
2. The List ID can be found in the URL or via API
3. Or use the API: `GET https://api.clickup.com/api/v2/team/{team_id}/space/{space_id}/list`

**Assignee IDs:**
1. Get team member IDs via API: `GET https://api.clickup.com/api/v2/team`
2. Add comma-separated IDs to `CLICKUP_ASSIGNEES` (e.g., `183,245,678`)

### 3. Run with Docker Compose

```bash
# Build and start the container
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the container
docker-compose down
```

### 4. Run Locally (Development)

```bash
# Install dependencies
go mod download

# Run the application
go run main.go
```

## Configuration

### Environment Variables

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `PORT` | No | Server port (default: 8080) | `8080` |
| `CLICKUP_API_TOKEN` | Yes | ClickUp API token | `pk_xxxxx` |
| `CLICKUP_LIST_ID` | Yes | Target ClickUp list ID | `123456789` |
| `CLICKUP_ASSIGNEES` | No | Comma-separated user IDs | `183,245,678` |

## API Endpoints

### `GET /`
Health check endpoint returning service information.

**Response:**
```json
{
  "service": "Prisma Cloud to ClickUp Webhook",
  "version": "1.0.0",
  "status": "running"
}
```

### `GET /health`
Health check for container orchestration.

**Response:**
```json
{
  "status": "healthy"
}
```

### `POST /webhook`
Receives Prisma Cloud alert webhooks and creates ClickUp tasks.

**Request Body (Example):**
```json
{
  "resourceId": "i-0123456789abcdef0",
  "alertRuleName": "High Severity Alerts",
  "accountName": "Production AWS",
  "cloudType": "aws",
  "severity": "high",
  "policyName": "S3 bucket is publicly accessible",
  "resourceName": "my-public-bucket",
  "policyDescription": "S3 bucket allows public read access",
  "alertId": "A-12345",
  "resourceRegion": "us-east-1",
  "resourceType": "S3 Bucket"
}
```

**Response (Success):**
```json
{
  "received": 1,
  "tasks_created": 1,
  "task_ids": ["abc123"],
  "status": "success"
}
```

## Prisma Cloud Configuration

### 1. Create Webhook Integration

1. Log in to Prisma Cloud
2. Go to **Settings** → **Integrations**
3. Click **Add Integration** → **Webhook**
4. Configure:
   - **Integration Name**: ClickUp Webhook
   - **URL**: `http://your-server:8080/webhook`
   - **Custom Payload**: Enable and use this template:

```json
{
  "resourceId": "${ResourceId}",
  "alertRuleName": "${AlertRuleName}",
  "accountName": "${AccountName}",
  "cloudType": "${CloudType}",
  "severity": "${Severity}",
  "policyName": "${PolicyName}",
  "resourceName": "${ResourceName}",
  "policyDescription": "${PolicyDescription}",
  "alertId": "${AlertId}",
  "resourceRegion": "${ResourceRegion}",
  "resourceType": "${ResourceType}"
}
```

### 2. Create Alert Rule

1. Go to **Alerts** → **Alert Rules**
2. Click **Add Alert Rule**
3. Configure the alert rule with your desired policies
4. In **Notifications**, select your webhook integration

## Severity to Priority Mapping

| Prisma Severity | ClickUp Priority |
|-----------------|------------------|
| Critical/High   | 1 (Urgent)       |
| Medium          | 2 (High)         |
| Low             | 3 (Normal)       |
| Default         | 4 (Low)          |

## Project Structure

```
prisma-webhook/
├── main.go                  # Application entry point
├── config/
│   └── config.go           # Configuration management
├── models/
│   └── prisma.go           # Prisma Cloud alert models
├── services/
│   └── clickup.go          # ClickUp API client
├── handlers/
│   └── webhook.go          # Webhook handler
├── Dockerfile              # Docker build configuration
├── docker-compose.yml      # Docker Compose configuration
├── go.mod                  # Go module dependencies
├── .env.example            # Environment variable template
└── README.md               # This file
```

## Development

### Build Binary

```bash
go build -o prisma-webhook
./prisma-webhook
```

### Run Tests (if tests are added)

```bash
go test ./...
```

### Build Docker Image

```bash
docker build -t prisma-webhook:latest .
```

## Troubleshooting

### Issue: "CLICKUP_API_TOKEN is required"
- Ensure `.env` file exists and contains valid `CLICKUP_API_TOKEN`

### Issue: "ClickUp API error (status 401)"
- Verify your API token is valid
- Check token hasn't expired

### Issue: "ClickUp API error (status 404)"
- Verify `CLICKUP_LIST_ID` is correct
- Ensure the list exists and you have access

### Issue: Tasks created without assignees
- Verify `CLICKUP_ASSIGNEES` contains valid user IDs
- Check user IDs have access to the list

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
