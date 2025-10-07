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

- Go 1.23 or later (for local development)
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
| `WEBHOOK_API_KEY` | Yes | API key for webhook authentication | `generated_key_here` |
| `ALLOWED_IPS` | No | Comma-separated allowed IPs | `203.0.113.1,198.51.100.0` |

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

**Security:**
- Requires `X-API-Key` header with valid API key
- IP allowlist validation (if configured)
- Rate limit: 100 requests per minute per IP

**Headers:**
```
X-API-Key: your_webhook_api_key
Content-Type: application/json
```

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
   - **Custom Headers**: Add `X-API-Key` header with your webhook API key
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
├── .github/
│   └── workflows/
│       ├── deploy.yml      # CI/CD workflow
│       └── README.md       # Workflow documentation
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

## CI/CD Deployment

This project includes GitHub Actions for automated deployment. See [.github/workflows/README.md](.github/workflows/README.md) for setup instructions.

**Quick setup:**
1. Configure GitHub Secrets (Docker Hub credentials, SSH details)
2. Push to `main` branch to trigger automatic deployment
3. Workflow builds Docker image and deploys to your server

## Security

### Multi-Layer Security Protection

This application implements multiple security layers:

1. **API Key Authentication**: All webhook requests must include `X-API-Key` header
2. **IP Allowlisting**: Restrict access to known Prisma Cloud IP addresses
3. **Rate Limiting**: Prevent abuse with configurable rate limits

### Generating API Key

```bash
# Generate a secure random API key
openssl rand -hex 32
```

Add the generated key to your `.env` file and Prisma Cloud webhook configuration.

### Rate Limits

| Endpoint | Rate Limit | Notes |
|----------|------------|-------|
| `/webhook` | 100 req/min per IP | For handling burst alerts |
| `/` | 60 req/min per IP | General endpoint |
| `/health` | No limit | For monitoring systems |

### Getting Prisma Cloud IPs

To configure IP allowlist:
1. Contact Prisma Cloud support or check documentation for webhook source IPs
2. Add IPs to `ALLOWED_IPS` in `.env` (comma-separated)
3. Leave empty to allow all IPs (not recommended for production)

## Troubleshooting

### Issue: "WEBHOOK_API_KEY is required"
- Ensure `.env` file contains `WEBHOOK_API_KEY`
- Generate key with: `openssl rand -hex 32`

### Issue: "Unauthorized: Invalid or missing API key"
- Verify `X-API-Key` header is included in webhook requests
- Ensure the key matches `WEBHOOK_API_KEY` in `.env`
- Check Prisma Cloud webhook configuration has correct header

### Issue: "Access denied: IP not allowed"
- Verify the IP is in `ALLOWED_IPS` list
- Check if you're behind a proxy/load balancer (IP may differ)
- Leave `ALLOWED_IPS` empty for testing (not recommended for production)

### Issue: "Rate limit exceeded"
- Default limits: 100 req/min for webhook, 60 req/min for general
- Wait for the time window to reset
- Adjust limits in `middleware/ratelimit.go` if needed

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
