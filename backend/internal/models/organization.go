package models

import "time"

type Organization struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name           string    `gorm:"size:255;not null" json:"name"`
	GithubOrgName  string    `gorm:"size:255;uniqueIndex;not null" json:"github_org_name"`
	GithubToken    string    `gorm:"size:512;not null" json:"-"` // Never expose in JSON
	Settings       string    `gorm:"type:json" json:"settings,omitempty"`
	IsActive       bool      `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations (not loaded by default)
	Repositories []Repository `gorm:"foreignKey:OrganizationID" json:"repositories,omitempty"`
	Developers   []Developer  `gorm:"foreignKey:OrganizationID" json:"developers,omitempty"`
}

func (Organization) TableName() string {
	return "organizations"
}

type OrganizationSettings struct {
	Repos           []string `json:"repos"`
	AutoReview      bool     `json:"auto_review"`
	ExcludePatterns []string `json:"exclude_patterns"`
}