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

### 4. Install Dependencies

```bash
make install
```

### 5. Run API Server

```bash
make dev
```

Server starts at: `http://localhost:3000`

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

## Next Steps

1. **Add Background Worker** - Run `make run-worker` to process PRs
2. **Add GitHub Integration** - Implement polling/webhooks
3. **Add AI Review Service** - Integrate with OpenAI/Anthropic
4. **Deploy** - Use Docker or deploy to VPS

## License

MIT
