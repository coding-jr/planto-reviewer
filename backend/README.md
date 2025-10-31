# Code Quality Tracker - Go Fiber Backend

Developer KPI tracking system with AI-powered code reviews.

## Features

- ✅ Multi-tenant organization support
- ✅ Developer performance metrics
- ✅ Issue tracking (null checks, logic errors, scalability, etc.)
- ✅ Code quality scoring
- ✅ REST API (no frontend needed)
- ✅ MySQL database with proper indexing
- ✅ Easy org onboarding (one API call)

## Tech Stack

- **Backend**: Go 1.21+ with Fiber v2
- **Database**: MySQL 8.0+
- **AI**: OpenAI/Anthropic/Google (configurable)

## Quick Start

### 1. Prerequisites

```bash
# Install Go 1.21+
# Install MySQL 8.0+
```

### 2. Setup Database

```bash
# Create database
mysql -u root -p < scripts/schema.sql

# Or manually:
mysql -u root -p
CREATE DATABASE code_quality_dev CHARACTER SET utf8mb4;
```

### 3. Configure Environment

```bash
cp .env.example .env
# Edit .env with your settings
```

**Important configuration:**
- `DATABASE_URL` - MySQL connection string
- `AI_PROVIDER` - Choose: `openai`, `anthropic`, or `google`
- `AI_API_KEY` - Your AI provider API key
- `API_KEY` - API key for authenticating requests
- `POLLING_INTERVAL` - How often to check for new PRs (seconds, default: 30)

### 4. Install Dependencies

```bash
make install
```

### 5. Run API Server

```bash
make dev
```

Server starts at: `http://localhost:3000`

### 6. Run Background Worker

The worker fetches PRs from GitHub and runs AI code reviews:

```bash
# In a separate terminal
make run-worker
```

**What the worker does:**
- Polls GitHub for new/updated PRs every 30 seconds (configurable)
- Runs AI reviews on PRs that need analysis
- Detects issues: null checks, logic errors, scalability, security, etc.
- Updates developer and organization metrics automatically
- Runs continuously in the background

## API Endpoints

### Organization Management

```bash
# Create organization (easy onboarding)
POST /api/organizations
{
  "name": "Acme Corp",
  "github_org_name": "acme-corp",
  "github_token": "ghp_xxxxx",
  "repos": ["repo1", "repo2"]
}

# List all organizations
GET /api/organizations

# Get organization details
GET /api/organizations/:id

# Update organization
PUT /api/organizations/:id

# Delete organization
DELETE /api/organizations/:id
```

### Developer Metrics (KPIs)

```bash
# Get developer metrics
GET /api/metrics/developer/:id?start_date=2024-01-01&end_date=2024-01-31

Response:
{
  "developer_id": 5,
  "period": {
    "start": "2024-01-01",
    "end": "2024-01-31"
  },
  "summary": {
    "total_prs": 12,
    "total_issues": 24,
    "avg_quality_score": 72.5
  },
  "daily_metrics": [...]
}
```

### Organization Metrics

```bash
# Get organization summary
GET /api/metrics/organization/:id/summary

# Get code quality leaderboard
GET /api/metrics/organization/:id/leaderboard

# Get top issue types
GET /api/metrics/organization/:id/top-issues?limit=10
```

## Authentication

All API endpoints (except `/health`) require an API key:

```bash
curl -H "Authorization: Bearer your-api-key" \
  http://localhost:3000/api/organizations
```

Set `API_KEY` in `.env` file.

## Database Schema

The system tracks:

- **Organizations** - Multi-tenant support
- **Repositories** - GitHub repos
- **Developers** - Developer profiles
- **Pull Requests** - All PRs with stats
- **Reviews** - AI review results
- **Issues Found** - Bugs, null checks, logic errors, etc.
- **Developer Metrics** - Daily aggregated KPIs
- **Organization Metrics** - Org-wide stats

## Development

```bash
# Run in development mode
make dev

# Build binaries
make build

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

## Project Structure

```
backend/
├── cmd/
│   ├── api/          # API server
│   ├── worker/       # Background worker
│   └── migrate/      # Database migrations
├── internal/
│   ├── config/       # Configuration
│   ├── database/     # Database connection
│   ├── models/       # GORM models
│   ├── handlers/     # HTTP handlers
│   ├── services/     # Business logic
│   ├── repositories/ # Database operations
│   └── middleware/   # HTTP middleware
├── pkg/
│   ├── ai/           # AI client
│   └── utils/        # Utilities
├── scripts/
│   └── schema.sql    # Database schema
└── .env.example      # Environment variables
```

## How It Works

1. **Onboard Organization** - Call `POST /api/organizations` with GitHub org details
2. **Worker Polls GitHub** - Every 30 seconds, fetches new/updated PRs
3. **AI Reviews Code** - Analyzes diffs and detects issues
4. **Metrics Calculated** - Developer and org KPIs updated automatically
5. **Query Metrics** - Use API endpoints to track performance without clicking through UI

## Deployment

### Docker (Recommended)

```bash
# Build images
docker build -t code-quality-api -f Dockerfile.api .
docker build -t code-quality-worker -f Dockerfile.worker .

# Run with docker-compose
docker-compose up -d
```

### VPS Deployment

```bash
# Build binaries
make build

# Run API server (with systemd or supervisor)
./bin/api

# Run worker (with systemd or supervisor)
./bin/worker
```

## Monitoring

**Worker Logs:**
```bash
# Shows PR fetching and review activity
tail -f logs/worker.log
```

**Check System Health:**
```bash
curl http://localhost:3000/health
```

## Next Steps

1. **Add GitHub Webhooks** - Real-time PR notifications instead of polling
2. **Add Notifications** - Slack/Email when critical issues found
3. **Add Historical Trends** - Track metrics over weeks/months
4. **Add Custom Rules** - Organization-specific coding standards

## License

MIT
