package models

import "time"

type Issue struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint64     `gorm:"not null;index:idx_org_developer,priority:1" json:"organization_id"`
	ReviewID       uint64     `gorm:"not null" json:"review_id"`
	PullRequestID  uint64     `gorm:"not null" json:"pull_request_id"`
	DeveloperID    uint64     `gorm:"not null;index:idx_org_developer,priority:2" json:"developer_id"`
	FilePath       string     `gorm:"size:1024;not null" json:"file_path"`
	LineNumber     *int       `json:"line_number,omitempty"`
	IssueType      string     `gorm:"size:100;not null;index" json:"issue_type"` // null_check, logic_error, scalability, security, performance
	Severity       string     `gorm:"size:50;not null;index" json:"severity"`    // critical, high, medium, low
	Title          string     `gorm:"size:512;not null" json:"title"`
	Description    string     `gorm:"type:text;not null" json:"description"`
	Suggestion     *string    `gorm:"type:text" json:"suggestion,omitempty"`
	CodeSnippet    *string    `gorm:"type:text" json:"code_snippet,omitempty"`
	IsResolved     bool       `gorm:"default:false;index" json:"is_resolved"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Review       *Review       `gorm:"foreignKey:ReviewID" json:"review,omitempty"`
	PullRequest  *PullRequest  `gorm:"foreignKey:PullRequestID" json:"pull_request,omitempty"`
	Developer    *Developer    `gorm:"foreignKey:DeveloperID" json:"developer,omitempty"`
}

func (Issue) TableName() string {
	return "issues_found"
}

// Issue type constants
const (
	IssueTypeNullCheck     = "null_check"
	IssueTypeLogicError    = "logic_error"
	IssueTypeScalability   = "scalability"
	IssueTypeSecurity      = "security"
	IssueTypePerformance   = "performance"
	IssueTypeRaceCondition = "race_condition"
	IssueTypeMemoryLeak    = "memory_leak"
	IssueTypeErrorHandling = "error_handling"
)

// Severity constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
)
