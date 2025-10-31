package models

import "time"

type PullRequest struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint64     `gorm:"not null;index:idx_org_repo,priority:1" json:"organization_id"`
	RepositoryID   uint64     `gorm:"not null;index:idx_org_repo,priority:2" json:"repository_id"`
	DeveloperID    uint64     `gorm:"not null;index" json:"developer_id"`
	GithubID           uint64     `gorm:"uniqueIndex;not null" json:"github_id"`
	PRNumber           uint       `gorm:"not null" json:"pr_number"`
	Title              string     `gorm:"size:512;not null" json:"title"`
	State              string     `gorm:"size:50;not null;index" json:"state"` // open, closed, merged
	BaseBranch         *string    `gorm:"size:255" json:"base_branch,omitempty"`
	HeadBranch         *string    `gorm:"size:255" json:"head_branch,omitempty"`
	LinesAdded         int        `gorm:"default:0" json:"lines_added"`
	LinesDeleted       int        `gorm:"default:0" json:"lines_deleted"`
	FilesChanged       int        `gorm:"default:0" json:"files_changed"`
	CommitsCount       int        `gorm:"default:0" json:"commits_count"`
	PRDiff             *string    `gorm:"type:longtext" json:"pr_diff,omitempty"`
	NeedsReview        bool       `gorm:"default:true;index" json:"needs_review"`
	CodeQualityScore   *float64   `gorm:"type:decimal(5,2)" json:"code_quality_score,omitempty"`
	IssuesFoundCount   *int       `gorm:"default:0" json:"issues_found_count,omitempty"`
	LastReviewID       *uint64    `json:"last_review_id,omitempty"`
	OpenedAt           time.Time  `gorm:"not null" json:"opened_at"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
	MergedAt           *time.Time `json:"merged_at,omitempty"`
	GithubCreatedAt    time.Time  `gorm:"not null;index" json:"github_created_at"`
	GithubUpdatedAt    time.Time  `gorm:"not null;index" json:"github_updated_at"`
	GithubMergedAt     *time.Time `json:"github_merged_at,omitempty"`
	CreatedAt          time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
	Repository   *Repository   `gorm:"foreignKey:RepositoryID" json:"repository,omitempty"`
	Developer    *Developer    `gorm:"foreignKey:DeveloperID" json:"developer,omitempty"`
	Reviews      []Review      `gorm:"foreignKey:PullRequestID" json:"reviews,omitempty"`
}

func (PullRequest) TableName() string {
	return "pull_requests"
}
