# Setup Guide - Code Quality Tracker

Complete guide to set up the Code Quality Tracker with Next.js dashboard and AWS Bedrock Claude Sonnet.

## Architecture Overview

- **Backend**: Go Fiber + MySQL (API server + Background worker)
- **Frontend**: Next.js 14 + TypeScript + Tailwind CSS
- **AI**: AWS Bedrock Claude 3.5 Sonnet (or OpenAI/Anthropic/Google)
- **Database**: MySQL 8.0+

## Prerequisites

- Go 1.21+
- Node.js 18+
- MySQL 8.0+
- AWS Account with Bedrock access (for Claude Sonnet)

---

## Part 1: Backend Setup

### Step 1: Configure Settings

**🔒 SECURITY FIRST:** Never commit credentials to git! Use environment variables (recommended) or a local `settings.json` file (which is git-ignored).

There are **two ways** to configure the application:

---

#### **Option A: Environment Variables (Recommended - Most Secure)**

**1.1 Create environment file:**
```bash
cd backend
cp .env.example .env
```

**1.2 Edit `backend/.env` with your credentials:**

```bash
# Server Configuration
PORT=3000
ENV=development

# Database Configuration
DATABASE_URL=root:YOUR_DB_PASSWORD@tcp(localhost:3306)/code_quality_dev?charset=utf8mb4&parseTime=True&loc=Local

# AI Provider
AI_PROVIDER=bedrock

# AWS Bedrock Configuration
AWS_REGION=ap-south-1

# Authentication Method 1: Bearer Token (Recommended - Simpler)
AWS_BEARER_TOKEN_BEDROCK=YOUR_BEDROCK_API_KEY_HERE

# Authentication Method 2: IAM Credentials (Alternative)
# AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY
# AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_KEY

# Bedrock Model Configuration
BEDROCK_MODEL=global.anthropic.claude-sonnet-4-5-20250929-v1:0
BEDROCK_MODEL_ARN=arn:aws:bedrock:ap-south-1:YOUR_ACCOUNT_ID:inference-profile/global.anthropic.claude-sonnet-4-5-20250929-v1:0
BEDROCK_MAX_TOKENS=16192
BEDROCK_TEMPERATURE=0.3

# API Security
API_KEY=YOUR_SECURE_RANDOM_API_KEY

# Worker Configuration
POLLING_INTERVAL=30
```

**How to Get Bedrock API Key:**
1. Go to **AWS Console** → **Bedrock** → **API Keys**
2. Click **"Create API Key"**
3. Copy the bearer token (starts with `ABSK...`)
4. Paste it into `backend/.env` as `AWS_BEARER_TOKEN_BEDROCK`

---

#### **Option B: settings.json (Alternative)**

**1.1 Copy the example file:**
```bash
cp settings.example.json settings.json
```

**1.2 Edit `settings.json` with your credentials:**

```json
{
  "aws": {
    "region": "ap-south-1",
    "bearerToken": "YOUR_BEDROCK_API_KEY_HERE"
  },
  "bedrock": {
    "model": "global.anthropic.claude-sonnet-4-5-20250929-v1:0",
    "modelArn": "arn:aws:bedrock:ap-south-1:YOUR_ACCOUNT_ID:inference-profile/...",
    "maxTokens": 16192,
    "temperature": 0.3
  },
  "api": {
    "port": 3000,
    "apiKey": "YOUR_SECURE_RANDOM_API_KEY"
  },
  "database": {
    "url": "root:YOUR_DB_PASSWORD@tcp(localhost:3306)/code_quality_dev?..."
  },
  "worker": {
    "pollingIntervalSeconds": 30
  }
}
```

**⚠️ Note:** `settings.json` is git-ignored for security. Never commit credentials to version control!

---

#### **Priority Order**

Environment variables **override** settings.json values:

1. **Environment variables** (highest priority)
2. **settings.json** (fallback)

This allows you to use `settings.json` for defaults and environment variables for sensitive credentials.

---

#### **Important Security Notes**

✅ **DO:**
- Use environment variables for production
- Keep credentials in `.env` files (git-ignored)
- Use strong random API keys
- Rotate credentials regularly

❌ **DON'T:**
- Commit credentials to git
- Share credentials in documentation
- Use weak or default API keys
- Store credentials in code comments

---

#### **Available Models**

- **Claude Sonnet 4.5** (Latest, Recommended): `global.anthropic.claude-sonnet-4-5-20250929-v1:0`
  - Max tokens: 16192
  - Requires inference profile ARN
  - Best performance and accuracy

- **Claude 3.5 Sonnet** (Alternative): `anthropic.claude-3-5-sonnet-20241022-v2:0`
  - Max tokens: 4000
  - Direct model access
  - Still very capable

---

#### **Other AI Providers**

If not using Bedrock, you can use other providers by changing `AI_PROVIDER` in `.env`:

```bash
# For OpenAI
AI_PROVIDER=openai
AI_API_KEY=sk-...
AI_MODEL=gpt-4-turbo-preview

# For Anthropic API
AI_PROVIDER=anthropic
AI_API_KEY=sk-ant-...
AI_MODEL=claude-3-5-sonnet-20241022

# For Google Gemini
AI_PROVIDER=google
AI_API_KEY=...
AI_MODEL=gemini-pro
```

### Step 2: Setup Database

**2.1 Create MySQL database:**
```bash
mysql -u root -p
```

```sql
CREATE DATABASE code_quality_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
exit;
```

**2.2 Run schema migration:**
```bash
cd backend
mysql -u root -p code_quality_dev < scripts/schema.sql
```

### Step 3: Install Backend Dependencies

```bash
cd backend

# Install Go dependencies
go mod download
go mod tidy

# Add AWS SDK for Bedrock
go get github.com/aws/aws-sdk-go/aws
go get github.com/aws/aws-sdk-go/service/bedrockruntime
```

### Step 4: Build Backend

```bash
make build
```

This creates:
- `bin/api` - API server
- `bin/worker` - Background worker

### Step 5: Run Backend

Open **two terminal windows**:

**Terminal 1 - API Server:**
```bash
cd backend
make dev
# or
./bin/api
```

API will start at: `http://localhost:3000`

**Terminal 2 - Background Worker:**
```bash
cd backend
make run-worker
# or
./bin/worker
```

Worker will:
- Poll GitHub every 30 seconds
- Run AI code reviews
- Update metrics automatically

### Step 6: Verify Backend

```bash
# Check health
curl http://localhost:3000/health

# Should return: {"status":"ok","env":"development"}
```

---

## Part 2: Frontend Setup

### Step 1: Install Frontend Dependencies

```bash
cd frontend
npm install
```

### Step 2: Configure Frontend

**2.1 Create environment file:**
```bash
cp .env.example .env
```

**2.2 Edit `.env`:**
```bash
NEXT_PUBLIC_API_URL=http://localhost:3000
NEXT_PUBLIC_API_KEY=your-api-key-here
```

Use the same `apiKey` from your `settings.json` or backend `.env`.

### Step 3: Run Frontend

```bash
npm run dev
```

Frontend will start at: `http://localhost:3001`

### Step 4: Access Dashboard

Open browser: `http://localhost:3001`

---

## Part 3: Onboard Your First Organization

### Option 1: Using curl

```bash
curl -X POST http://localhost:3000/api/organizations \
  -H "Authorization: Bearer your-api-key-here" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "github_org_name": "my-github-org",
    "github_token": "ghp_xxxxxxxxxxxxx",
    "repos": ["repo1", "repo2"]
  }'
```

### Option 2: Using Postman/Insomnia

- **Method**: POST
- **URL**: `http://localhost:3000/api/organizations`
- **Headers**:
  - `Authorization`: `Bearer your-api-key-here`
  - `Content-Type`: `application/json`
- **Body**:
```json
{
  "name": "My Company",
  "github_org_name": "my-github-org",
  "github_token": "ghp_xxxxxxxxxxxxx",
  "repos": ["repo1", "repo2"]
}
```

### Get GitHub Token

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select scopes:
   - `repo` (Full control of private repositories)
   - `read:org` (Read org and team membership)
4. Copy the token

---

## Part 4: How It Works

### Automatic Flow

1. **Worker polls GitHub** (every 30 seconds)
   - Fetches new/updated PRs
   - Stores PR diffs in database

2. **AI reviews code**
   - AWS Bedrock Claude analyzes diffs
   - Detects issues: null checks, logic errors, scalability, security, etc.
   - Assigns code quality scores (0-100)

3. **Metrics calculated**
   - Developer KPIs updated
   - Organization stats aggregated
   - Trends tracked over time

4. **View in dashboard**
   - No clicking through UI needed
   - All metrics available via API or dashboard

### Dashboard Features

- **Organizations List**: See all tracked organizations
- **Organization Dashboard**:
  - Total developers, PRs, issues
  - Issues by type (null_check, logic_error, scalability, security, etc.)
  - Issues by severity (critical, high, medium, low)
- **Developer Leaderboard**: Top 10 developers by code quality
- **Top Issues**: Most common issues across the organization

---

## Part 5: API Endpoints

### Organizations

```bash
# List organizations
GET /api/organizations

# Get organization details
GET /api/organizations/:id

# Create organization
POST /api/organizations

# Update organization
PUT /api/organizations/:id

# Delete organization
DELETE /api/organizations/:id
```

### Metrics

```bash
# Developer metrics
GET /api/metrics/developer/:id?start_date=2024-01-01&end_date=2024-01-31

# Organization summary
GET /api/metrics/organization/:id/summary

# Developer leaderboard
GET /api/metrics/organization/:id/leaderboard

# Top issues
GET /api/metrics/organization/:id/top-issues?limit=10
```

All endpoints require `Authorization: Bearer your-api-key` header.

---

## Part 6: Troubleshooting

### Backend Issues

**Database connection error:**
```bash
# Check MySQL is running
mysql -u root -p -e "SELECT 1"

# Verify DATABASE_URL in settings.json
```

**AWS Bedrock error:**
```bash
# Verify AWS credentials
aws bedrock list-foundation-models --region us-east-1

# Check Bedrock is enabled in your AWS account
```

**Worker not processing PRs:**
```bash
# Check worker logs
tail -f logs/worker.log

# Verify GitHub token has correct permissions
# Verify repos exist and are accessible
```

### Frontend Issues

**Cannot connect to API:**
```bash
# Verify API is running
curl http://localhost:3000/health

# Check NEXT_PUBLIC_API_URL in frontend/.env
```

**401 Unauthorized:**
```bash
# Verify NEXT_PUBLIC_API_KEY matches backend API_KEY
```

---

## Part 7: Production Deployment

### Backend Deployment

**Using Docker:**
```bash
cd backend

# Build images
docker build -t code-quality-api -f Dockerfile.api .
docker build -t code-quality-worker -f Dockerfile.worker .

# Run with docker-compose
docker-compose up -d
```

**Using systemd (Linux):**

Create `/etc/systemd/system/code-quality-api.service`:
```ini
[Unit]
Description=Code Quality API
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/code-quality/backend
ExecStart=/opt/code-quality/backend/bin/api
Restart=always

[Install]
WantedBy=multi-user.target
```

Create `/etc/systemd/system/code-quality-worker.service`:
```ini
[Unit]
Description=Code Quality Worker
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/code-quality/backend
ExecStart=/opt/code-quality/backend/bin/worker
Restart=always

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl enable code-quality-api code-quality-worker
sudo systemctl start code-quality-api code-quality-worker
```

### Frontend Deployment

**Using Vercel** (Recommended):
```bash
cd frontend
npm install -g vercel
vercel
```

**Using Docker:**
```bash
cd frontend
docker build -t code-quality-frontend .
docker run -p 3001:3000 code-quality-frontend
```

---

## Part 8: Next Steps

1. **Monitor worker logs** to see PRs being processed
2. **Check dashboard** at `http://localhost:3001`
3. **Query metrics API** to build custom reports
4. **Add more organizations** as needed
5. **Set up notifications** (Slack/Email) for critical issues
6. **Implement GitHub webhooks** for real-time processing (optional)

---

## Support

For issues or questions:
- Check logs: `backend/logs/`
- Review API documentation: `backend/README.md`
- Check GitHub issues: https://github.com/coding-jr/planto-reviewer/issues
