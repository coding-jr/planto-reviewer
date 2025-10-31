package models

import "time"

type Developer struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint64    `gorm:"not null;uniqueIndex:uk_org_username,priority:1" json:"organization_id"`
	GithubUsername string    `gorm:"size:255;not null;uniqueIndex:uk_org_username,priority:2;index" json:"github_username"`
	GithubID       *uint64   `json:"github_id,omitempty"`
	Email          *string   `gorm:"size:255" json:"email,omitempty"`
	Name           *string   `gorm:"size:255" json:"name,omitempty"`
	AvatarURL      *string   `gorm:"size:512" json:"avatar_url,omitempty"`
	TotalPRs       int       `gorm:"default:0" json:"total_prs"`
	TotalCommits   int       `gorm:"default:0" json:"total_commits"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

func (Developer) TableName() string {
	return "developers"
}