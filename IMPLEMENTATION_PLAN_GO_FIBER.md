# 🎯 Implementation Plan: Developer KPI Tracking with Go Fiber + MySQL

## Overview

**Goal**: Track developer code quality metrics, reduce bugs/issues, monitor performance across organizations - **without needing to click through a dashboard**. Data-first approach with API access.

**Tech Stack**:
- **Backend**: Go 1.21+ with Fiber v2
- **Database**: MySQL 8.0+
- **Architecture**: Multi-tenant, API-first
- **No Frontend** (initially - just data + APIs)

---

## 🎯 What We're Building

### Core Features
1. ✅ **Store all code review data** in MySQL
2. ✅ **Track developer KPIs** (bugs found, code quality scores)
3. ✅ **Multi-org support** (easy onboarding)
4. ✅ **API endpoints** to query metrics
5. ✅ **GitHub integration** (polling or webhooks)
6. ✅ **AI review engine** (detect bugs, null checks, logic issues)

### What We're NOT Building (Yet)
- ❌ Frontend dashboard UI
- ❌ Authentication UI
- ❌ Complex visualizations
- ❌ Real-time WebSocket updates
- ❌ Mobile apps

---

## 🗄️ MySQL Database Schema

### Multi-Tenant Design

```sql
-- Organizations (tenants)
CREATE TABLE organizations (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255) NOT NULL,
    github_org_name VARCHAR(255) UNIQUE NOT NULL,
    github_token VARCHAR(512) NOT NULL, -- encrypted
    settings JSON,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_github_org (github_org_name),
    INDEX idx_active (is_active)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Repositories
CREATE TABLE repositories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    github_id BIGINT UNSIGNED UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(512) NOT NULL,
    language VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    last_synced_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    INDEX idx_org_repo (organization_id, name),
    INDEX idx_github_id (github_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Developers (extracted from commits/PRs)
CREATE TABLE developers (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    github_username VARCHAR(255) NOT NULL,
    github_id BIGINT UNSIGNED,
    email VARCHAR(255),
    name VARCHAR(255),
    avatar_url VARCHAR(512),
    total_prs INT DEFAULT 0,
    total_commits INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    UNIQUE KEY uk_org_username (organization_id, github_username),
    INDEX idx_github_username (github_username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Pull Requests
CREATE TABLE pull_requests (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    repository_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    github_id BIGINT UNSIGNED UNIQUE NOT NULL,
    pr_number INT NOT NULL,
    title VARCHAR(512) NOT NULL,
    state VARCHAR(50) NOT NULL, -- open, closed, merged
    base_branch VARCHAR(255),
    head_branch VARCHAR(255),
    lines_added INT DEFAULT 0,
    lines_deleted INT DEFAULT 0,
    files_changed INT DEFAULT 0,
    commits_count INT DEFAULT 0,
    opened_at TIMESTAMP NOT NULL,
    closed_at TIMESTAMP NULL,
    merged_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (repository_id) REFERENCES repositories(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    INDEX idx_org_repo (organization_id, repository_id),
    INDEX idx_developer (developer_id),
    INDEX idx_state (state),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Reviews (AI-generated reviews)
CREATE TABLE reviews (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    pull_request_id BIGINT UNSIGNED NOT NULL,
    review_type VARCHAR(50) NOT NULL, -- summary, code_review, explanation
    model_used VARCHAR(100) NOT NULL, -- gpt-4, claude-3, etc
    tokens_used INT DEFAULT 0,
    cost_usd DECIMAL(10, 6) DEFAULT 0.000000,
    status VARCHAR(50) NOT NULL, -- pending, completed, failed
    summary_text TEXT,
    completed_at TIMESTAMP NULL,
    error_message TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    INDEX idx_org_pr (organization_id, pull_request_id),
    INDEX idx_status (status),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Issues Found (the KPIs we care about!)
CREATE TABLE issues_found (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    review_id BIGINT UNSIGNED NOT NULL,
    pull_request_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    file_path VARCHAR(1024) NOT NULL,
    line_number INT,
    issue_type VARCHAR(100) NOT NULL, -- null_check, logic_error, scalability, security, etc
    severity VARCHAR(50) NOT NULL, -- critical, high, medium, low
    title VARCHAR(512) NOT NULL,
    description TEXT NOT NULL,
    suggestion TEXT,
    code_snippet TEXT,
    is_resolved BOOLEAN DEFAULT FALSE,
    resolved_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    INDEX idx_org_developer (organization_id, developer_id),
    INDEX idx_issue_type (issue_type),
    INDEX idx_severity (severity),
    INDEX idx_created_at (created_at),
    INDEX idx_resolved (is_resolved)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Developer KPI Metrics (aggregated data for fast queries)
CREATE TABLE developer_metrics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    developer_id BIGINT UNSIGNED NOT NULL,
    metric_date DATE NOT NULL, -- daily aggregation
    total_prs INT DEFAULT 0,
    total_issues_found INT DEFAULT 0,
    critical_issues INT DEFAULT 0,
    high_issues INT DEFAULT 0,
    medium_issues INT DEFAULT 0,
    low_issues INT DEFAULT 0,
    null_check_issues INT DEFAULT 0,
    logic_error_issues INT DEFAULT 0,
    scalability_issues INT DEFAULT 0,
    security_issues INT DEFAULT 0,
    performance_issues INT DEFAULT 0,
    lines_added INT DEFAULT 0,
    lines_deleted INT DEFAULT 0,
    avg_pr_size DECIMAL(10, 2) DEFAULT 0.00,
    code_quality_score DECIMAL(5, 2) DEFAULT 0.00, -- 0-100
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    FOREIGN KEY (developer_id) REFERENCES developers(id) ON DELETE CASCADE,
    UNIQUE KEY uk_dev_date (developer_id, metric_date),
    INDEX idx_org_date (organization_id, metric_date),
    INDEX idx_code_quality (code_quality_score)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Organization-wide metrics (for comparison across orgs)
CREATE TABLE organization_metrics (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    organization_id BIGINT UNSIGNED NOT NULL,
    metric_date DATE NOT NULL,
    total_prs INT DEFAULT 0,
    total_developers INT DEFAULT 0,
    total_issues_found INT DEFAULT 0,
    avg_issues_per_pr DECIMAL(5, 2) DEFAULT 0.00,
    avg_code_quality_score DECIMAL(5, 2) DEFAULT 0.00,
    total_cost_usd DECIMAL(10, 2) DEFAULT 0.00,
    total_tokens_used BIGINT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (organization_id) REFERENCES organizations(id) ON DELETE CASCADE,
    UNIQUE KEY uk_org_date (organization_id, metric_date),
    INDEX idx_date (metric_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

---

## 🏗️ Go Fiber Project Structure

```
code-quality-tracker/
├── cmd/
│   ├── api/
│   │   └── main.go                 # API server entry point
│   ├── worker/
│   │   └── main.go                 # Background worker for reviews
│   └── migrate/
│       └── main.go                 # Database migrations
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration loading
│   ├── database/
│   │   ├── mysql.go                # MySQL connection
│   │   └── migrations/             # SQL migration files
│   ├── models/
│   │   ├── organization.go
│   │   ├── repository.go
│   │   ├── developer.go
│   │   ├── pull_request.go
│   │   ├── review.go
│   │   ├── issue.go
│   │   └── metrics.go
│   ├── handlers/
│   │   ├── organization_handler.go # Org CRUD APIs
│   │   ├── metrics_handler.go      # KPI APIs
│   │   ├── developer_handler.go    # Developer APIs
│   │   └── webhook_handler.go      # GitHub webhooks
│   ├── services/
│   │   ├── github_service.go       # GitHub API integration
│   │   ├── ai_service.go           # AI review service
│   │   ├── metrics_service.go      # KPI calculation
│   │   └── review_service.go       # Review orchestration
│   ├── repositories/
│   │   ├── organization_repo.go    # DB operations
│   │   ├── developer_repo.go
│   │   └── metrics_repo.go
│   └── middleware/
│       ├── auth.go                 # API key authentication
│       └── logger.go               # Request logging
├── pkg/
│   ├── ai/
│   │   ├── openai.go               # OpenAI client
│   │   ├── anthropic.go            # Anthropic client
│   │   └── prompts.go              # AI prompts
│   └── utils/
│       ├── crypto.go               # Encryption utils
│       └── response.go             # API response helpers
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
└── go.sum
```

---

## 📝 Implementation Steps (Senior Dev Approach)

### Phase 1: Foundation (Week 1)

#### Step 1.1: Project Setup (Day 1)
```bash
# Initialize Go module
mkdir code-quality-tracker && cd code-quality-tracker
go mod init github.com/yourusername/code-quality-tracker

# Install core dependencies
go get github.com/gofiber/fiber/v2
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/joho/godotenv
go get github.com/google/go-github/v57/github
```

#### Step 1.2: Database Setup (Day 1-2)
```bash
# Create MySQL database
mysql -u root -p

CREATE DATABASE code_quality_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'cqt_user'@'localhost' IDENTIFIED BY 'secure_password';
GRANT ALL PRIVILEGES ON code_quality_dev.* TO 'cqt_user'@'localhost';
FLUSH PRIVILEGES;

# Run migrations (copy schema from above)
mysql -u cqt_user -p code_quality_dev < schema.sql
```

#### Step 1.3: Config + Database Connection (Day 2)

**`internal/config/config.go`**:
```go
package config

import (
    "github.com/joho/godotenv"
    "log"
    "os"
)

type Config struct {
    Port           string
    DatabaseURL    string
    AIProvider     string
    AIAPIKey       string
    AIModel        string
    EncryptionKey  string
}

func Load() *Config {
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    return &Config{
        Port:          getEnv("PORT", "3000"),
        DatabaseURL:   getEnv("DATABASE_URL", ""),
        AIProvider:    getEnv("AI_PROVIDER", "openai"),
        AIAPIKey:      getEnv("AI_API_KEY", ""),
        AIModel:       getEnv("AI_MODEL", "gpt-4"),
        EncryptionKey: getEnv("ENCRYPTION_KEY", ""),
    }
}

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

**`internal/database/mysql.go`**:
```go
package database

import (
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "log"
)

func Connect(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })

    if err != nil {
        return nil, err
    }

    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)

    log.Println("✅ Database connected successfully")
    return db, nil
}
```

#### Step 1.4: Core Models (Day 2-3)

**`internal/models/organization.go`**:
```go
package models

import "time"

type Organization struct {
    ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    Name           string    `gorm:"size:255;not null" json:"name"`
    GithubOrgName  string    `gorm:"size:255;uniqueIndex;not null" json:"github_org_name"`
    GithubToken    string    `gorm:"size:512;not null" json:"-"` // encrypted, never expose
    Settings       string    `gorm:"type:json" json:"settings"`
    IsActive       bool      `gorm:"default:true" json:"is_active"`
    CreatedAt      time.Time `json:"created_at"`
    UpdatedAt      time.Time `json:"updated_at"`

    // Relations
    Repositories []Repository `gorm:"foreignKey:OrganizationID" json:"repositories,omitempty"`
    Developers   []Developer  `gorm:"foreignKey:OrganizationID" json:"developers,omitempty"`
}

type Developer struct {
    ID              uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
    OrganizationID  uint64    `gorm:"not null;index:idx_org_username,priority:1" json:"organization_id"`
    GithubUsername  string    `gorm:"size:255;not null;index:idx_org_username,priority:2" json:"github_username"`
    GithubID        *uint64   `json:"github_id"`
    Email           *string   `gorm:"size:255" json:"email"`
    Name            *string   `gorm:"size:255" json:"name"`
    AvatarURL       *string   `gorm:"size:512" json:"avatar_url"`
    TotalPRs        int       `gorm:"default:0" json:"total_prs"`
    TotalCommits    int       `gorm:"default:0" json:"total_commits"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`

    // Relations
    Organization Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

// Add other models similarly...
```

---

### Phase 2: API Endpoints (Week 2)

#### Step 2.1: Organization APIs (Day 1)

**`internal/handlers/organization_handler.go`**:
```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "code-quality-tracker/internal/models"
    "code-quality-tracker/internal/repositories"
)

type OrganizationHandler struct {
    repo *repositories.OrganizationRepository
}

func NewOrganizationHandler(repo *repositories.OrganizationRepository) *OrganizationHandler {
    return &OrganizationHandler{repo: repo}
}

// POST /api/organizations
func (h *OrganizationHandler) Create(c *fiber.Ctx) error {
    var req struct {
        Name          string `json:"name" validate:"required"`
        GithubOrgName string `json:"github_org_name" validate:"required"`
        GithubToken   string `json:"github_token" validate:"required"`
    }

    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
    }

    org := &models.Organization{
        Name:          req.Name,
        GithubOrgName: req.GithubOrgName,
        GithubToken:   req.GithubToken, // TODO: Encrypt before saving
        IsActive:      true,
    }

    if err := h.repo.Create(org); err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.Status(201).JSON(org)
}

// GET /api/organizations
func (h *OrganizationHandler) List(c *fiber.Ctx) error {
    orgs, err := h.repo.FindAll()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(orgs)
}

// GET /api/organizations/:id
func (h *OrganizationHandler) Get(c *fiber.Ctx) error {
    id, err := c.ParamsInt("id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
    }

    org, err := h.repo.FindByID(uint64(id))
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Organization not found"})
    }

    return c.JSON(org)
}
```

#### Step 2.2: Metrics APIs (Day 2-3) - **THE IMPORTANT PART**

**`internal/handlers/metrics_handler.go`**:
```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "code-quality-tracker/internal/services"
)

type MetricsHandler struct {
    metricsService *services.MetricsService
}

func NewMetricsHandler(service *services.MetricsService) *MetricsHandler {
    return &MetricsHandler{metricsService: service}
}

// GET /api/metrics/developer/:developer_id
// Query params: start_date, end_date
func (h *MetricsHandler) GetDeveloperMetrics(c *fiber.Ctx) error {
    developerID, err := c.ParamsInt("developer_id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid developer ID"})
    }

    startDate := c.Query("start_date") // YYYY-MM-DD
    endDate := c.Query("end_date")

    metrics, err := h.metricsService.GetDeveloperMetrics(
        uint64(developerID),
        startDate,
        endDate,
    )

    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "developer_id": developerID,
        "period": fiber.Map{
            "start": startDate,
            "end": endDate,
        },
        "metrics": metrics,
    })
}

// GET /api/metrics/organization/:org_id/summary
func (h *MetricsHandler) GetOrgSummary(c *fiber.Ctx) error {
    orgID, err := c.ParamsInt("org_id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid org ID"})
    }

    summary, err := h.metricsService.GetOrgSummary(uint64(orgID))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(summary)
}

// GET /api/metrics/organization/:org_id/top-issues
func (h *MetricsHandler) GetTopIssues(c *fiber.Ctx) error {
    orgID, err := c.ParamsInt("org_id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid org ID"})
    }

    limit := c.QueryInt("limit", 10)

    issues, err := h.metricsService.GetTopIssues(uint64(orgID), limit)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(fiber.Map{
        "organization_id": orgID,
        "top_issues": issues,
    })
}

// GET /api/metrics/organization/:org_id/leaderboard
func (h *MetricsHandler) GetLeaderboard(c *fiber.Ctx) error {
    orgID, err := c.ParamsInt("org_id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid org ID"})
    }

    leaderboard, err := h.metricsService.GetCodeQualityLeaderboard(uint64(orgID))
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    return c.JSON(leaderboard)
}
```

---

### Phase 3: Background Worker (Week 3)

#### Step 3.1: Review Service (Day 1-2)

**`internal/services/review_service.go`**:
```go
package services

import (
    "code-quality-tracker/internal/models"
    "code-quality-tracker/pkg/ai"
    "encoding/json"
)

type ReviewService struct {
    aiClient *ai.Client
}

type ReviewResult struct {
    Summary string  `json:"summary"`
    Issues  []Issue `json:"issues"`
}

type Issue struct {
    FilePath    string `json:"file_path"`
    LineNumber  int    `json:"line_number"`
    IssueType   string `json:"issue_type"` // null_check, logic_error, scalability, etc
    Severity    string `json:"severity"`   // critical, high, medium, low
    Title       string `json:"title"`
    Description string `json:"description"`
    Suggestion  string `json:"suggestion"`
    CodeSnippet string `json:"code_snippet"`
}

func (s *ReviewService) ReviewPullRequest(pr *models.PullRequest, files []PullRequestFile) (*ReviewResult, error) {
    // Build AI prompt focusing on KPIs we care about
    prompt := buildReviewPrompt(pr, files)

    // Call AI
    response, err := s.aiClient.Complete(prompt)
    if err != nil {
        return nil, err
    }

    // Parse response
    var result ReviewResult
    if err := json.Unmarshal([]byte(response), &result); err != nil {
        return nil, err
    }

    return &result, nil
}

func buildReviewPrompt(pr *models.PullRequest, files []PullRequestFile) string {
    return `You are a senior software engineer reviewing code for quality issues.

Focus on finding these specific issues:
1. NULL_CHECK: Missing null/undefined checks
2. LOGIC_ERROR: Flawed logic, edge cases not handled
3. SCALABILITY: Performance issues, N+1 queries, inefficient algorithms
4. SECURITY: SQL injection, XSS, exposed secrets
5. RACE_CONDITION: Concurrency issues
6. MEMORY_LEAK: Resource leaks, unclosed connections
7. ERROR_HANDLING: Missing try-catch, unhandled errors

For each issue found, provide:
- file_path
- line_number
- issue_type (from list above)
- severity (critical/high/medium/low)
- title (brief)
- description (what's wrong)
- suggestion (how to fix)

Return JSON only.`
}
```

#### Step 3.2: Background Worker (Day 3-5)

**`cmd/worker/main.go`**:
```go
package main

import (
    "log"
    "time"
    "code-quality-tracker/internal/config"
    "code-quality-tracker/internal/database"
    "code-quality-tracker/internal/services"
)

func main() {
    cfg := config.Load()
    db, err := database.Connect(cfg.DatabaseURL)
    if err != nil {
        log.Fatal(err)
    }

    githubService := services.NewGithubService(cfg)
    reviewService := services.NewReviewService(cfg)
    metricsService := services.NewMetricsService(db)

    log.Println("🚀 Worker started - polling for new PRs...")

    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        processOrganizations(db, githubService, reviewService, metricsService)
    }
}

func processOrganizations(/* ... */) {
    // 1. Get all active organizations
    // 2. For each org, get repositories
    // 3. For each repo, check for new/updated PRs
    // 4. For new PRs, run AI review
    // 5. Save issues to database
    // 6. Update metrics
}
```

---

### Phase 4: Multi-Tenant Onboarding (Week 4)

#### Easy Org Onboarding Flow

**1. API Call to Create Org**:
```bash
curl -X POST http://localhost:3000/api/organizations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Acme Corp",
    "github_org_name": "acme-corp",
    "github_token": "ghp_xxxxxxxxxxxx",
    "repos": ["repo1", "repo2", "repo3"]
  }'
```

**2. System Auto-Syncs**:
- Worker picks up new org
- Fetches all repos
- Starts monitoring PRs
- Begins generating metrics

**3. Query Metrics Immediately**:
```bash
curl http://localhost:3000/api/metrics/organization/1/summary
```

**Done! No UI clicking needed.**

---

## 🔍 Key API Endpoints (No UI Needed)

### Organization Management
```
POST   /api/organizations                 # Onboard new org
GET    /api/organizations                 # List all orgs
GET    /api/organizations/:id             # Get org details
PUT    /api/organizations/:id             # Update org
DELETE /api/organizations/:id             # Remove org
```

### Developer KPIs (The Important Ones)
```
GET /api/metrics/developer/:id                           # Get developer metrics
GET /api/metrics/developer/:id/issues                    # Issues by developer
GET /api/metrics/developer/:id/trends                    # Quality trends over time
GET /api/metrics/organization/:id/developers             # All developers ranked
GET /api/metrics/organization/:id/leaderboard            # Code quality leaderboard
```

### Issue Tracking
```
GET /api/metrics/organization/:id/issues                 # All issues
GET /api/metrics/organization/:id/issues/by-type         # Group by type
GET /api/metrics/organization/:id/issues/by-severity     # Group by severity
GET /api/metrics/organization/:id/issues/unresolved      # Open issues
```

### Aggregate Metrics
```
GET /api/metrics/organization/:id/summary                # Overall stats
GET /api/metrics/organization/:id/trends                 # Trends over time
GET /api/metrics/organization/:id/top-issues             # Most common issues
GET /api/metrics/organization/:id/cost                   # AI API costs
```

---

## 📊 Example API Responses

### GET /api/metrics/developer/5

```json
{
  "developer_id": 5,
  "github_username": "john_doe",
  "period": {
    "start": "2024-01-01",
    "end": "2024-01-31"
  },
  "metrics": {
    "total_prs": 12,
    "total_issues_found": 24,
    "issues_by_type": {
      "null_check": 8,
      "logic_error": 6,
      "scalability": 4,
      "security": 3,
      "performance": 3
    },
    "issues_by_severity": {
      "critical": 2,
      "high": 7,
      "medium": 10,
      "low": 5
    },
    "code_quality_score": 72.5,
    "avg_issues_per_pr": 2.0,
    "lines_added": 1245,
    "lines_deleted": 342,
    "avg_pr_size": 132
  }
}
```

### GET /api/metrics/organization/1/leaderboard

```json
{
  "organization_id": 1,
  "leaderboard": [
    {
      "rank": 1,
      "developer_id": 3,
      "github_username": "alice_smith",
      "code_quality_score": 92.3,
      "total_prs": 15,
      "total_issues": 8,
      "avg_issues_per_pr": 0.53
    },
    {
      "rank": 2,
      "developer_id": 7,
      "github_username": "bob_jones",
      "code_quality_score": 85.1,
      "total_prs": 22,
      "total_issues": 18,
      "avg_issues_per_pr": 0.82
    }
  ]
}
```

---

## 🚀 Deployment

### Docker Compose Setup

**`docker-compose.yml`**:
```yaml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpass
      MYSQL_DATABASE: code_quality_prod
      MYSQL_USER: cqt_user
      MYSQL_PASSWORD: securepass
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  api:
    build: .
    command: /app/api
    ports:
      - "3000:3000"
    environment:
      DATABASE_URL: "cqt_user:securepass@tcp(mysql:3306)/code_quality_prod?charset=utf8mb4&parseTime=True"
      AI_PROVIDER: openai
      AI_API_KEY: ${OPENAI_API_KEY}
      AI_MODEL: gpt-4
    depends_on:
      - mysql

  worker:
    build: .
    command: /app/worker
    environment:
      DATABASE_URL: "cqt_user:securepass@tcp(mysql:3306)/code_quality_prod?charset=utf8mb4&parseTime=True"
      AI_PROVIDER: openai
      AI_API_KEY: ${OPENAI_API_KEY}
    depends_on:
      - mysql

volumes:
  mysql_data:
```

---

## ⏱️ Realistic Timeline

### Week 1: Foundation
- Day 1-2: Project setup, database schema, config
- Day 3-4: Core models (GORM structs)
- Day 5: Basic API server setup

### Week 2: APIs
- Day 1-2: Organization CRUD APIs
- Day 3-4: Metrics APIs (the important ones)
- Day 5: Testing APIs with Postman

### Week 3: Background Processing
- Day 1-2: GitHub service (fetch PRs)
- Day 3-4: AI review service
- Day 5: Background worker (polling)

### Week 4: Multi-Tenant & Polish
- Day 1-2: Onboarding flow
- Day 3-4: Metrics calculation service
- Day 5: Testing with multiple orgs

**Total: 4 weeks (1 month) for working system**

---

## 💡 Why Go Fiber + MySQL?

### ✅ Your Choice is Good Because:

1. **Go Fiber**:
   - ✅ 10x faster than Node.js for APIs
   - ✅ Lower memory usage (~10MB vs ~50MB)
   - ✅ Built-in concurrency (goroutines)
   - ✅ Simple, Express-like API
   - ✅ Production-ready

2. **MySQL**:
   - ✅ You already know it
   - ✅ Excellent for structured data
   - ✅ ACID transactions
   - ✅ Mature ecosystem
   - ✅ Good performance for this scale

### 🤔 Alternative: PostgreSQL?

**Pros of switching to PostgreSQL**:
- ✅ Better JSON support (JSONB)
- ✅ More advanced features (CTEs, window functions)
- ✅ Better for analytics queries

**Cons**:
- ❌ You'd need to learn it
- ❌ MySQL is fine for this use case

**Verdict**: **Stick with MySQL** unless you need advanced analytics features.

---

## 🎯 Summary: What You Get

### Data You Can Query (No UI Clicks Needed)

1. **Developer Performance**:
   - Issues per PR (avg)
   - Code quality score (0-100)
   - Most common issues by developer
   - Trends over time

2. **Issue Tracking**:
   - All issues found
   - Grouped by type (null checks, logic errors, etc.)
   - Grouped by severity
   - Per file, per developer

3. **Organization Overview**:
   - Top issue types across org
   - Developer leaderboard
   - Total cost (AI API usage)
   - Trends over time

4. **Easy Onboarding**:
   - One API call to add new org
   - Auto-sync starts immediately
   - Metrics available within minutes

### What You DON'T Get (Good!)

- ❌ No frontend to maintain
- ❌ No authentication UI
- ❌ No complex visualizations
- ❌ No real-time WebSockets

**Just clean APIs returning JSON data.**

---

## 📝 Next Steps

1. **Review this plan** - Make sure it fits your needs
2. **I'll create starter code** - Full Go Fiber boilerplate if you want
3. **Deploy to VPS** - $5-10/month DigitalOcean droplet
4. **Add orgs via API** - Start tracking immediately

**This is production-ready architecture for your use case.**

Want me to generate the actual Go code to get you started?