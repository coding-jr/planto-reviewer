# Quick Start Guide

Get your Code Quality Tracker running in 5 minutes with AWS Bedrock Claude Sonnet 4.5!

## ⚡ Prerequisites

- ✅ MySQL running on localhost
- ✅ Node.js 18+ installed
- ✅ Go 1.21+ installed
- ✅ AWS Bedrock API key (from your VSCode settings)

## 🚀 5-Minute Setup

### 1. Database Setup (1 min)

```bash
# Create database
mysql -u root -p -e "CREATE DATABASE code_quality_dev CHARACTER SET utf8mb4;"

# Run schema
cd backend
mysql -u root -p code_quality_dev < scripts/schema.sql
```

### 2. Configuration (Already Done! ✅)

Your `settings.json` is already configured with:
- ✅ AWS Region: `ap-south-1`
- ✅ Bedrock API Key (bearer token)
- ✅ Claude Sonnet 4.5 model
- ✅ Max tokens: 16192

Just verify your database password in `settings.json`:
```json
"database": {
  "url": "root:YOUR_PASSWORD@tcp(localhost:3306)/code_quality_dev?..."
}
```

### 3. Install & Run Backend (2 min)

```bash
cd backend

# Install dependencies
go mod download

# Build
make build

# Run API server (Terminal 1)
./bin/api

# Run worker (Terminal 2)
./bin/worker
```

You should see:
```
✅ Loaded configuration from settings.json
✅ Using AWS Bedrock with API Key authentication
ℹ️  Model: global.anthropic.claude-sonnet-4-5-20250929-v1:0
ℹ️  Region: ap-south-1
🚀 Background worker started
```

### 4. Install & Run Frontend (2 min)

```bash
cd frontend

# Install
npm install

# Configure (update API key)
cp .env.example .env
# Edit .env and set NEXT_PUBLIC_API_KEY to match settings.json

# Run
npm run dev
```

Frontend: **http://localhost:3001**

### 5. Onboard Organization

```bash
# Get your GitHub token from: https://github.com/settings/tokens
# Scopes needed: repo, read:org

curl -X POST http://localhost:3000/api/organizations \
  -H "Authorization: Bearer your-api-key" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Company",
    "github_org_name": "your-github-org",
    "github_token": "ghp_xxxxxxxxxxxxx",
    "repos": ["repo1", "repo2"]
  }'
```

## ✨ What You Get

### Dashboard Features
- 📊 Organization overview with key metrics
- 🏆 Developer leaderboard (top 10 by code quality)
- 🐛 Top issues tracking
- 📈 Issues breakdown by type and severity
- 🎯 Real-time KPI updates

### Issue Detection
The system automatically detects:
- ❌ `null_check` - Missing null/undefined checks
- 🔍 `logic_error` - Flawed logic, edge cases
- 📈 `scalability` - Performance issues, N+1 queries
- 🔒 `security` - SQL injection, XSS, exposed secrets
- ⚡ `performance` - Inefficient algorithms
- 🔄 `race_condition` - Concurrency issues
- 💾 `memory_leak` - Resource leaks
- ⚠️ `error_handling` - Missing try-catch

### Automatic Workflow
1. **Worker polls GitHub** (every 30s)
2. **Claude Sonnet 4.5 reviews** PRs
3. **Metrics updated** automatically
4. **View in dashboard** (no clicking needed!)

## 🔧 Configuration

Your configuration should be set via environment variables (recommended) or `settings.json` (not tracked in git):

**Option 1: Environment Variables (Recommended)**
```bash
# Copy the example file
cp backend/.env.example backend/.env

# Edit backend/.env and add your credentials:
AWS_BEARER_TOKEN_BEDROCK=your-bedrock-api-key-here
```

**Option 2: settings.json**
```bash
# Copy the example file
cp settings.example.json settings.json

# Edit settings.json and add your credentials
```

**Model Configuration:**
- Model: `global.anthropic.claude-sonnet-4-5-20250929-v1:0`
- Region: `ap-south-1`
- Max Tokens: `16192`
- Temperature: `0.3`

**Using:**
- ✅ AWS Bedrock Claude Sonnet 4.5 (latest model)
- ✅ Bearer token authentication (no IAM needed)
- ✅ Inference profile for optimal performance
- ✅ 16K max tokens for detailed reviews

## 📖 Next Steps

1. **View dashboard**: http://localhost:3001
2. **Check worker logs**: See PRs being processed in Terminal 2
3. **Query API directly**: Use the endpoints in SETUP.md
4. **Add more organizations**: Use the same curl command

## 🆘 Troubleshooting

**Worker not processing PRs?**
```bash
# Check GitHub token has correct scopes
# Verify repos exist in the organization
# Check worker logs for errors
```

**API returns 401?**
```bash
# Verify API_KEY in settings.json matches NEXT_PUBLIC_API_KEY in frontend/.env
```

**Database error?**
```bash
# Check MySQL is running: mysql -u root -p -e "SELECT 1"
# Verify DATABASE_URL in settings.json
```

## 📚 Full Documentation

- **[SETUP.md](SETUP.md)** - Complete 8-part setup guide
- **[frontend/README.md](frontend/README.md)** - Frontend documentation
- **[backend/README.md](backend/README.md)** - Backend API documentation

---

**Powered by AWS Bedrock Claude Sonnet 4.5** 🚀
