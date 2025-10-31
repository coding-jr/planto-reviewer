package handlers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type MetricsHandler struct {
	db *gorm.DB
}

func NewMetricsHandler(db *gorm.DB) *MetricsHandler {
	return &MetricsHandler{db: db}
}

// GET /api/metrics/developer/:id
func (h *MetricsHandler) GetDeveloperMetrics(c *fiber.Ctx) error {
	developerID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid developer ID",
		})
	}

	startDate := c.Query("start_date") // YYYY-MM-DD
	endDate := c.Query("end_date")

	// Query developer metrics
	query := h.db.Table("developer_metrics").
		Select("*").
		Where("developer_id = ?", developerID)

	if startDate != "" {
		query = query.Where("metric_date >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("metric_date <= ?", endDate)
	}

	var metrics []map[string]interface{}
	if err := query.Find(&metrics).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch metrics",
		})
	}

	// Calculate aggregates
	var totalPRs, totalIssues int
	var codeQualitySum float64
	for _, m := range metrics {
		if prs, ok := m["total_prs"].(int64); ok {
			totalPRs += int(prs)
		}
		if issues, ok := m["total_issues_found"].(int64); ok {
			totalIssues += int(issues)
		}
		if quality, ok := m["code_quality_score"].(float64); ok {
			codeQualitySum += quality
		}
	}

	avgQuality := 0.0
	if len(metrics) > 0 {
		avgQuality = codeQualitySum / float64(len(metrics))
	}

	return c.JSON(fiber.Map{
		"developer_id": developerID,
		"period": fiber.Map{
			"start": startDate,
			"end":   endDate,
		},
		"summary": fiber.Map{
			"total_prs":         totalPRs,
			"total_issues":      totalIssues,
			"avg_quality_score": avgQuality,
		},
		"daily_metrics": metrics,
	})
}

// GET /api/metrics/organization/:id/summary
func (h *MetricsHandler) GetOrgSummary(c *fiber.Ctx) error {
	orgID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	// Get total issues by type
	var issuesByType []struct {
		IssueType string `json:"issue_type"`
		Count     int64  `json:"count"`
	}

	h.db.Table("issues_found").
		Select("issue_type, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("issue_type").
		Scan(&issuesByType)

	// Get total issues by severity
	var issuesBySeverity []struct {
		Severity string `json:"severity"`
		Count    int64  `json:"count"`
	}

	h.db.Table("issues_found").
		Select("severity, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("severity").
		Scan(&issuesBySeverity)

	// Get developer count
	var developerCount int64
	h.db.Table("developers").
		Where("organization_id = ?", orgID).
		Count(&developerCount)

	// Get PR count
	var prCount int64
	h.db.Table("pull_requests").
		Where("organization_id = ?", orgID).
		Count(&prCount)

	return c.JSON(fiber.Map{
		"organization_id": orgID,
		"summary": fiber.Map{
			"total_developers": developerCount,
			"total_prs":        prCount,
			"issues_by_type":   issuesByType,
			"issues_by_severity": issuesBySeverity,
		},
	})
}

// GET /api/metrics/organization/:id/leaderboard
func (h *MetricsHandler) GetLeaderboard(c *fiber.Ctx) error {
	orgID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	var leaderboard []struct {
		DeveloperID      uint64  `json:"developer_id"`
		GithubUsername   string  `json:"github_username"`
		TotalPRs         int     `json:"total_prs"`
		TotalIssues      int     `json:"total_issues"`
		CodeQualityScore float64 `json:"code_quality_score"`
	}

	h.db.Table("developers d").
		Select("d.id as developer_id, d.github_username, d.total_prs, "+
			"COALESCE(SUM(dm.total_issues_found), 0) as total_issues, "+
			"COALESCE(AVG(dm.code_quality_score), 0) as code_quality_score").
		Joins("LEFT JOIN developer_metrics dm ON d.id = dm.developer_id").
		Where("d.organization_id = ?", orgID).
		Group("d.id, d.github_username, d.total_prs").
		Order("code_quality_score DESC").
		Limit(10).
		Scan(&leaderboard)

	return c.JSON(fiber.Map{
		"organization_id": orgID,
		"leaderboard":     leaderboard,
	})
}

// GET /api/metrics/organization/:id/top-issues
func (h *MetricsHandler) GetTopIssues(c *fiber.Ctx) error {
	orgID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	limit := c.QueryInt("limit", 10)

	var topIssues []struct {
		IssueType   string `json:"issue_type"`
		Severity    string `json:"severity"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Count       int64  `json:"count"`
	}

	h.db.Table("issues_found").
		Select("issue_type, severity, title, description, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("issue_type, severity, title, description").
		Order("count DESC").
		Limit(limit).
		Scan(&topIssues)

	return c.JSON(fiber.Map{
		"organization_id": orgID,
		"top_issues":      topIssues,
	})
}
