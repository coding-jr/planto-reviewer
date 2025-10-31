package models

import "time"

type Repository struct {
	ID             uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	OrganizationID uint64     `gorm:"not null;index:idx_org_repo,priority:1" json:"organization_id"`
	GithubID       uint64     `gorm:"uniqueIndex;not null" json:"github_id"`
	Name           string     `gorm:"size:255;not null;index:idx_org_repo,priority:2" json:"name"`
	FullName       string     `gorm:"size:512;not null" json:"full_name"`
	Language       *string    `gorm:"size:100" json:"language,omitempty"`
	IsActive       bool       `gorm:"default:true;index" json:"is_active"`
	LastSyncedAt   *time.Time `json:"last_synced_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Organization *Organization `gorm:"foreignKey:OrganizationID" json:"organization,omitempty"`
}

func (Repository) TableName() string {
	return "repositories"
}