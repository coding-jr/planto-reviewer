package models

import "time"

type DeveloperMetric struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID       uint64    `gorm:"not null;index:idx_org_date,priority:1" json:"organization_id"`
	DeveloperID          uint64    `gorm:"not null;uniqueIndex:uk_dev_date,priority:1" json:"developer_id"`
	MetricDate           time.Time `gorm:"type:date;not null;uniqueIndex:uk_dev_date,priority:2;index:idx_org_date,priority:2" json:"metric_date"`
	TotalPRs             int       `gorm:"default:0" json:"total_prs"`
	TotalIssuesFound     int       `gorm:"default:0" json:"total_issues_found"`
	CriticalIssues       int       `gorm:"default:0" json:"critical_issues"`
	HighIssues           int       `gorm:"default:0" json:"high_issues"`
	MediumIssues         int       `gorm:"default:0" json:"medium_issues"`
	LowIssues            int       `gorm:"default:0" json:"low_issues"`
	NullCheckIssues      int       `gorm:"default:0" json:"null_check_issues"`
	LogicErrorIssues     int       `gorm:"default:0" json:"logic_error_issues"`
	ScalabilityIssues    int       `gorm:"default:0" json:"scalability_issues"`
	SecurityIssues       int       `gorm:"default:0" json:"security_issues"`
	PerformanceIssues    int       `gorm:"default:0" json:"performance_issues"`
	LinesAdded           int       `gorm:"default:0" json:"lines_added"`
	LinesDeleted         int       `gorm:"default:0" json:"lines_deleted"`
	AvgPRSize            float64   `gorm:"type:decimal(10,2);default:0" json:"avg_pr_size"`
	CodeQualityScore     float64   `gorm:"type:decimal(5,2);default:0;index" json:"code_quality_score"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Developer    *Developer    `gorm:"foreignKey:DeveloperID" json:"developer,omitempty"`
}

func (DeveloperMetric) TableName() string {
	return "developer_metrics"
}

type OrganizationMetric struct {
	ID                   uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID       uint64    `gorm:"not null;uniqueIndex:uk_org_date,priority:1" json:"organization_id"`
	MetricDate           time.Time `gorm:"type:date;not null;uniqueIndex:uk_org_date,priority:2;index" json:"metric_date"`
	TotalPRs             int       `gorm:"default:0" json:"total_prs"`
	TotalDevelopers      int       `gorm:"default:0" json:"total_developers"`
	TotalIssuesFound     int       `gorm:"default:0" json:"total_issues_found"`
	AvgIssuesPerPR       float64   `gorm:"type:decimal(5,2);default:0" json:"avg_issues_per_pr"`
	AvgCodeQualityScore  float64   `gorm:"type:decimal(5,2);default:0" json:"avg_code_quality_score"`
	TotalCostUSD         float64   `gorm:"type:decimal(10,2);default:0" json:"total_cost_usd"`
	TotalTokensUsed      int64     `gorm:"default:0" json:"total_tokens_used"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

func (OrganizationMetric) TableName() string {
	return "organization_metrics"
}
