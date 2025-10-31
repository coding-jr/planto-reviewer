package models

import "time"

type Review struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint64     `gorm:"not null;index:idx_org_pr,priority:1" json:"organization_id"`
	PullRequestID  uint64     `gorm:"not null;index:idx_org_pr,priority:2" json:"pull_request_id"`
	ReviewSummary    string     `gorm:"type:text" json:"review_summary"`
	CodeQualityScore float64    `gorm:"type:decimal(5,2);default:0" json:"code_quality_score"`
	IssuesFoundCount int        `gorm:"default:0" json:"issues_found_count"`
	ReviewedAt       time.Time  `gorm:"not null;index" json:"reviewed_at"`
	ReviewType       string     `gorm:"size:50;not null" json:"review_type"` // summary, code_review, explanation
	ModelUsed        *string    `gorm:"size:100" json:"model_used,omitempty"`
	TokensUsed       int        `gorm:"default:0" json:"tokens_used"`
	CostUSD          float64    `gorm:"type:decimal(10,6);default:0" json:"cost_usd"`
	Status           string     `gorm:"size:50;not null;index;default:'completed'" json:"status"` // pending, completed, failed
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
	ErrorMessage     *string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	PullRequest  *PullRequest  `gorm:"foreignKey:PullRequestID" json:"pull_request,omitempty"`
	Issues       []Issue       `gorm:"foreignKey:ReviewID" json:"issues,omitempty"`
}

func (Review) TableName() string {
	return "reviews"
}
