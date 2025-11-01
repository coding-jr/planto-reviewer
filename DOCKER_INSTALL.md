# Docker Installation Guide

Quick and easy installation using Docker Compose.

## ⚡ Prerequisites

- Docker Desktop (macOS/Windows) or Docker Engine + Docker Compose (Linux)
- AWS Bedrock API Key
- GitHub Personal Access Token

## 🚀 Quick Install (5 minutes)

### 1. Clone the repository

```bash
git clone https://github.com/YOUR_USERNAME/code-quality-reviewer.git
cd code-quality-reviewer
```

### 2. Run the installation script

```bash
./install.sh
```

The script will:
- Check prerequisites
- Create `.env` file from template
- Build Docker images
- Start all services (API, Worker, Database, Frontend)
- Verify everything is running

### 3. Edit `.env` file

When prompted, edit the `.env` file:

```bash
# Required credentials
AWS_BEARER_TOKEN_BEDROCK=your-bedrock-api-key-here
BEDROCK_MODEL_ARN=arn:aws:bedrock:ap-south-1:YOUR_ACCOUNT_ID:inference-profile/...
API_KEY=your-secure-random-key-here
```

**Get AWS Bedrock API Key:**
1. Go to AWS Console → Bedrock → API Keys
2. Create new API key
3. Copy the bearer token (starts with `ABSK...`)

**Generate API Key:**
```bash
openssl rand -hex 32
```

### 4. Access the services

- **API**: http://localhost:3000
- **Dashboard**: http://localhost:3001
- **Health Check**: http://localhost:3000/health

## 📝 Manual Installation

If you prefer manual setup:

### 1. Create `.env` file

```bash
cp .env.docker .env
```

Edit `.env` with your credentials.

### 2. Start services

```bash
docker-compose up -d --build
```

### 3. Check logs

```bash
docker-compose logs -f
```

### 4. Verify health

```bash
curl http://localhost:3000/health
```

## 🔧 Configuration

All configuration is done via the `.env` file:

```bash
# Database
DB_ROOT_PASSWORD=rootpassword
DB_USER=codeuser
DB_PASSWORD=codepassword

# AWS Bedrock
AWS_REGION=ap-south-1
AWS_BEARER_TOKEN_BEDROCK=your-key-here
BEDROCK_MODEL=global.anthropic.claude-sonnet-4-5-20250929-v1:0
BEDROCK_MODEL_ARN=arn:aws:bedrock:...
BEDROCK_MAX_TOKENS=16192
BEDROCK_TEMPERATURE=0.3

# Security
API_KEY=your-secure-key

# Worker
POLLING_INTERVAL=30
```

## 🎯 Usage

### Add an organization

```bash
API_KEY=$(grep "^API_KEY=" .env | cut -d '=' -f2)

curl -X POST http://localhost:3000/api/organizations \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "github_org_name": "my-github-org",
    "github_token": "ghp_your_github_token",
    "repos": ["repo1", "repo2"]
  }'
```

### View metrics

```bash
# Get organizations
curl http://localhost:3000/api/organizations \
  -H "Authorization: Bearer $API_KEY"

# Get org summary
curl http://localhost:3000/api/metrics/organization/1/summary \
  -H "Authorization: Bearer $API_KEY"
```

### View dashboard

Open http://localhost:3001 in your browser.

## 🛠️ Management Commands

### View logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f api
docker-compose logs -f worker
docker-compose logs -f db
```

### Restart services

```bash
# All services
docker-compose restart

# Specific service
docker-compose restart api
docker-compose restart worker
```

### Stop services

```bash
docker-compose stop
```

### Start services

```bash
docker-compose start
```

### Rebuild and restart

```bash
docker-compose up -d --build
```

### Remove everything (including data)

```bash
docker-compose down -v
```

## 🔍 Troubleshooting

### API not starting

```bash
# Check logs
docker-compose logs api

# Common issues:
# - Missing AWS credentials in .env
# - Invalid bearer token
# - Database not ready
```

### Worker not processing PRs

```bash
# Check logs
docker-compose logs worker

# Common issues:
# - Invalid GitHub token
# - Repository doesn't exist
# - No pull requests to process
```

### Database connection error

```bash
# Check database
docker-compose logs db

# Restart database
docker-compose restart db
```

### Can't access services

```bash
# Check if containers are running
docker ps

# Expected output:
# - code-quality-api (port 3000)
# - code-quality-worker
# - code-quality-db (port 3306)
# - code-quality-frontend (port 3001)

# Check if ports are in use
lsof -i :3000
lsof -i :3001
```

## 🔐 Security Notes

1. **Never commit `.env` file** - It contains sensitive credentials
2. **Use strong API keys** - Generate with `openssl rand -hex 32`
3. **Restrict database access** - Change default passwords
4. **Use HTTPS in production** - Add reverse proxy (nginx/traefik)
5. **Limit API access** - Use firewall rules

## 🚀 Production Deployment

For production use:

### 1. Use environment-specific .env

```bash
# .env.production
DB_ROOT_PASSWORD=strong-random-password
API_KEY=strong-random-api-key
```

### 2. Add reverse proxy

Use nginx or traefik for HTTPS:

```yaml
# docker-compose.prod.yml
services:
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
```

### 3. Enable HTTPS

- Get SSL certificate (Let's Encrypt)
- Configure nginx for HTTPS
- Redirect HTTP to HTTPS

### 4. Set up backups

```bash
# Backup database
docker exec code-quality-db mysqldump -u root -p code_quality_dev > backup.sql

# Restore database
docker exec -i code-quality-db mysql -u root -p code_quality_dev < backup.sql
```

### 5. Monitor services

- Set up health check monitoring
- Configure alerting (Slack, email)
- Use logging service (ELK, Datadog)

## 📚 Additional Resources

- [Complete Setup Guide](SETUP.md)
- [Quick Start Guide](QUICKSTART.md)
- [Security Guide](SECURITY.md)
- [API Documentation](backend/README.md)

## 🆘 Need Help?

- Check logs: `docker-compose logs -f`
- Verify configuration: `cat .env`
- Test API: `curl http://localhost:3000/health`
- Check GitHub issues for similar problems
