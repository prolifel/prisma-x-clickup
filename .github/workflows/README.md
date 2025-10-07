# GitHub Actions CI/CD Setup

This workflow automatically builds and deploys the Prisma Cloud webhook service when code is pushed to the `main` branch.

## Workflow Steps

1. **Build Go Application** - Compiles the Go backend
2. **Run Tests** - Executes test suite (if available)
3. **Build Docker Image** - Creates Docker image and pushes to GitHub Container Registry (GHCR)
4. **Deploy to Server** - SSHs to server and updates the running service

## Required GitHub Secrets

Configure these secrets in your GitHub repository settings (`Settings` → `Secrets and variables` → `Actions`):

### SSH Server Access (Required)
- `SSH_HOST` - Server IP address or hostname (e.g., `192.168.1.100` or `server.example.com`)
- `SSH_USERNAME` - SSH username (e.g., `ubuntu`, `root`, `deploy`)
- `SSH_PRIVATE_KEY` - SSH private key for authentication
- `SSH_PORT` - SSH port (default: `22`)

## Setup Instructions

### 1. Generate SSH Key Pair (if needed)

On your local machine:
```bash
ssh-keygen -t ed25519 -C "github-actions" -f ~/.ssh/github-actions
```

Copy the public key to your server:
```bash
ssh-copy-id -i ~/.ssh/github-actions.pub user@your-server
```

Copy the private key content:
```bash
cat ~/.ssh/github-actions
```

Add this private key to GitHub Secrets as `SSH_PRIVATE_KEY`.

### 2. Configure GitHub Container Registry (GHCR)

No additional configuration needed! The workflow uses `GITHUB_TOKEN` which is automatically provided by GitHub Actions.

**Important:** Make sure your repository's package is set to public, or configure GHCR permissions:
1. Go to your repository → Settings → Actions → General
2. Scroll to "Workflow permissions"
3. Select "Read and write permissions"
4. Save

To make the container image public:
1. Go to your GitHub profile → Packages
2. Find `prisma-webhook` package
3. Package settings → Change visibility → Public

### 3. Prepare Server

SSH to your server and prepare the deployment directory:

```bash
# Create deployment directory
mkdir -p /path/to/prisma-webhook
cd /path/to/prisma-webhook

# Create .env file with your configuration
nano .env
```

Add to `.env`:
```env
PORT=8080
CLICKUP_API_TOKEN=your_token
CLICKUP_LIST_ID=your_list_id
CLICKUP_ASSIGNEES=123456789
DOCKER_IMAGE=ghcr.io/your-github-username/prisma-webhook:latest
```

Copy `docker-compose.yml` to server:
```bash
# On your local machine
scp docker-compose.yml user@server:/path/to/prisma-webhook/
```

### 4. Update Deployment Path

Edit `.github/workflows/deploy.yml` and update the deployment path:
```yaml
script: |
  cd /path/to/prisma-webhook  # Change this to your actual path
  echo ${{ secrets.GITHUB_TOKEN }} | docker login ghcr.io -u ${{ github.actor }} --password-stdin
  docker-compose pull
  docker-compose down
  docker-compose up -d
  docker system prune -f
```

## Testing the Workflow

1. Make a commit to the `main` branch
2. Push to GitHub: `git push origin main`
3. Check the Actions tab in GitHub to see the workflow running
4. Verify deployment on your server: `docker ps`

## Troubleshooting

### SSH Connection Failed
- Verify SSH credentials are correct
- Test SSH connection manually: `ssh -i private_key user@host`
- Check server firewall allows SSH connections
- Ensure private key format is correct (should include `-----BEGIN OPENSSH PRIVATE KEY-----`)

### Docker Build Failed
- Check Dockerfile syntax
- Verify Go build succeeds locally
- Review GitHub Actions logs for specific errors

### Docker Pull Failed
- Verify GHCR authentication is working (check GitHub Actions logs)
- Check image name matches in workflow and docker-compose.yml
- Ensure server can access ghcr.io (network/firewall)
- Verify GITHUB_TOKEN has package write permissions
- Check if the package is public or server has authentication

### Container Not Starting
- Check server logs: `docker logs prisma-webhook`
- Verify `.env` file exists on server
- Check environment variables are set correctly
- Verify port 8080 is not already in use

## Local Testing

Test the workflow steps locally before pushing:

```bash
# Build
go build -o prisma-webhook .

# Build Docker image
docker build -t prisma-webhook:latest .

# Run with docker-compose
docker-compose up -d

# Check logs
docker-compose logs -f

# Stop
docker-compose down
```

## Security Notes

- **Never commit secrets** to the repository
- Use GitHub Secrets for all sensitive data
- Rotate SSH keys regularly
- GITHUB_TOKEN is automatically rotated by GitHub
- Use non-root user for deployments when possible
- Keep SSH port non-standard (not 22) if possible
- Enable SSH key-only authentication (disable password auth)
- Set container images to public or configure GHCR authentication on server
