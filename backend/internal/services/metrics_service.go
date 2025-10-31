package services

import (
	"fmt"
	"time"

	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"gorm.io/gorm"
)

type MetricsService struct {
	db *gorm.DB
}

func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{db: db}
}

// UpdateDeveloperMetrics calculates and updates metrics for a developer
func (s *MetricsService) UpdateDeveloperMetrics(developerID uint64) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Get developer info
	var dev models.Developer
	if err := s.db.First(&dev, developerID).Error; err != nil {
		return fmt.Errorf("failed to fetch developer: %w", err)
	}

	// Calculate metrics for today
	metrics, err := s.calculateDeveloperMetrics(developerID, today)
	if err != nil {
		return fmt.Errorf("failed to calculate metrics: %w", err)
	}

	// Check if metrics exist for today
	var existing models.DeveloperMetric
	err = s.db.Where("developer_id = ? AND DATE(metric_date) = DATE(?)", developerID, today).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new metrics
		if err := s.db.Create(metrics).Error; err != nil {
			return fmt.Errorf("failed to create metrics: %w", err)
		}
		fmt.Printf("✅ Created metrics for developer %d\n", developerID)
	} else if err != nil {
		return fmt.Errorf("failed to check existing metrics: %w", err)
	} else {
		// Update existing metrics
		if err := s.db.Model(&existing).Updates(metrics).Error; err != nil {
			return fmt.Errorf("failed to update metrics: %w", err)
		}
		fmt.Printf("✅ Updated metrics for developer %d\n", developerID)
	}

	return nil
}

// calculateDeveloperMetrics calculates metrics for a developer for a specific date
func (s *MetricsService) calculateDeveloperMetrics(developerID uint64, date time.Time) (*models.DeveloperMetric, error) {
	var dev models.Developer
	if err := s.db.First(&dev, developerID).Error; err != nil {
		return nil, err
	}

	dateStr := date.Format("2006-01-02")

	// Count PRs for the date
	var totalPRs int64
	s.db.Model(&models.PullRequest{}).
		Where("developer_id = ? AND DATE(github_created_at) = ?", developerID, dateStr).
		Count(&totalPRs)

	// Count total issues found
	var totalIssues int64
	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ?", developerID, dateStr).
		Count(&totalIssues)

	// Count issues by type
	var nullCheckIssues, logicErrorIssues, scalabilityIssues, securityIssues, performanceIssues int64

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND issue_type = ?", developerID, dateStr, "null_check").
		Count(&nullCheckIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND issue_type = ?", developerID, dateStr, "logic_error").
		Count(&logicErrorIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND issue_type = ?", developerID, dateStr, "scalability").
		Count(&scalabilityIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND issue_type = ?", developerID, dateStr, "security").
		Count(&securityIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND issue_type = ?", developerID, dateStr, "performance").
		Count(&performanceIssues)

	// Count issues by severity
	var criticalIssues, highIssues, mediumIssues, lowIssues int64

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND severity = ?", developerID, dateStr, "critical").
		Count(&criticalIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND severity = ?", developerID, dateStr, "high").
		Count(&highIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND severity = ?", developerID, dateStr, "medium").
		Count(&mediumIssues)

	s.db.Model(&models.Issue{}).
		Where("developer_id = ? AND DATE(created_at) = ? AND severity = ?", developerID, dateStr, "low").
		Count(&lowIssues)

	// Calculate average code quality score
	var avgQuality struct {
		Score float64
	}
	s.db.Model(&models.PullRequest{}).
		Select("COALESCE(AVG(code_quality_score), 0) as score").
		Where("developer_id = ? AND DATE(github_created_at) = ? AND code_quality_score IS NOT NULL", developerID, dateStr).
		Scan(&avgQuality)

	// Calculate total lines of code
	var linesOfCode struct {
		Added   int64
		Deleted int64
	}
	s.db.Model(&models.PullRequest{}).
		Select("COALESCE(SUM(lines_added), 0) as added, COALESCE(SUM(lines_deleted), 0) as deleted").
		Where("developer_id = ? AND DATE(github_created_at) = ?", developerID, dateStr).
		Scan(&linesOfCode)

	// Calculate average PR size
	avgPRSize := 0.0
	if totalPRs > 0 {
		avgPRSize = float64(linesOfCode.Added+linesOfCode.Deleted) / float64(totalPRs)
	}

	metrics := &models.DeveloperMetric{
		DeveloperID:       developerID,
		OrganizationID:    dev.OrganizationID,
		MetricDate:        date,
		TotalPRs:          int(totalPRs),
		TotalIssuesFound:  int(totalIssues),
		CodeQualityScore:  avgQuality.Score,
		NullCheckIssues:   int(nullCheckIssues),
		LogicErrorIssues:  int(logicErrorIssues),
		ScalabilityIssues: int(scalabilityIssues),
		SecurityIssues:    int(securityIssues),
		PerformanceIssues: int(performanceIssues),
		CriticalIssues:    int(criticalIssues),
		HighIssues:        int(highIssues),
		MediumIssues:      int(mediumIssues),
		LowIssues:         int(lowIssues),
		LinesAdded:        int(linesOfCode.Added),
		LinesDeleted:      int(linesOfCode.Deleted),
		AvgPRSize:         avgPRSize,
	}

	return metrics, nil
}

// UpdateOrganizationMetrics calculates and updates metrics for an organization
func (s *MetricsService) UpdateOrganizationMetrics(orgID uint64) error {
	today := time.Now().Truncate(24 * time.Hour)

	// Calculate metrics for today
	metrics, err := s.calculateOrganizationMetrics(orgID, today)
	if err != nil {
		return fmt.Errorf("failed to calculate org metrics: %w", err)
	}

	// Check if metrics exist for today
	var existing models.OrganizationMetric
	err = s.db.Where("organization_id = ? AND DATE(metric_date) = DATE(?)", orgID, today).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new metrics
		if err := s.db.Create(metrics).Error; err != nil {
			return fmt.Errorf("failed to create org metrics: %w", err)
		}
		fmt.Printf("✅ Created metrics for organization %d\n", orgID)
	} else if err != nil {
		return fmt.Errorf("failed to check existing org metrics: %w", err)
	} else {
		// Update existing metrics
		if err := s.db.Model(&existing).Updates(metrics).Error; err != nil {
			return fmt.Errorf("failed to update org metrics: %w", err)
		}
		fmt.Printf("✅ Updated metrics for organization %d\n", orgID)
	}

	return nil
}

// calculateOrganizationMetrics calculates metrics for an organization for a specific date
func (s *MetricsService) calculateOrganizationMetrics(orgID uint64, date time.Time) (*models.OrganizationMetric, error) {
	dateStr := date.Format("2006-01-02")

	// Count total developers
	var totalDevelopers int64
	s.db.Model(&models.Developer{}).
		Where("organization_id = ?", orgID).
		Count(&totalDevelopers)

	// Count PRs for the date
	var totalPRs int64
	s.db.Model(&models.PullRequest{}).
		Where("organization_id = ? AND DATE(github_created_at) = ?", orgID, dateStr).
		Count(&totalPRs)

	// Count total issues
	var totalIssues int64
	s.db.Model(&models.Issue{}).
		Where("organization_id = ? AND DATE(created_at) = ?", orgID, dateStr).
		Count(&totalIssues)

	// Calculate average code quality score
	var avgQuality struct {
		Score float64
	}
	s.db.Model(&models.PullRequest{}).
		Select("COALESCE(AVG(code_quality_score), 0) as score").
		Where("organization_id = ? AND DATE(github_created_at) = ? AND code_quality_score IS NOT NULL", orgID, dateStr).
		Scan(&avgQuality)

	// Calculate average issues per PR
	avgIssuesPerPR := 0.0
	if totalPRs > 0 {
		avgIssuesPerPR = float64(totalIssues) / float64(totalPRs)
	}

	metrics := &models.OrganizationMetric{
		OrganizationID:      orgID,
		MetricDate:          date,
		TotalDevelopers:     int(totalDevelopers),
		TotalPRs:            int(totalPRs),
		TotalIssuesFound:    int(totalIssues),
		AvgIssuesPerPR:      avgIssuesPerPR,
		AvgCodeQualityScore: avgQuality.Score,
	}

	return metrics, nil
}

// RecalculateAllMetrics recalculates metrics for all developers and organizations
func (s *MetricsService) RecalculateAllMetrics() error {
	// Get all developers
	var developers []models.Developer
	if err := s.db.Find(&developers).Error; err != nil {
		return fmt.Errorf("failed to fetch developers: %w", err)
	}

	fmt.Printf("📊 Recalculating metrics for %d developers\n", len(developers))

	for _, dev := range developers {
		if err := s.UpdateDeveloperMetrics(dev.ID); err != nil {
			fmt.Printf("❌ Failed to update metrics for developer %d: %v\n", dev.ID, err)
		}
	}

	// Get all organizations
	var orgs []models.Organization
	if err := s.db.Find(&orgs).Error; err != nil {
		return fmt.Errorf("failed to fetch organizations: %w", err)
	}

	fmt.Printf("📊 Recalculating metrics for %d organizations\n", len(orgs))

	for _, org := range orgs {
		if err := s.UpdateOrganizationMetrics(org.ID); err != nil {
			fmt.Printf("❌ Failed to update metrics for org %d: %v\n", org.ID, err)
		}
	}

	return nil
}
