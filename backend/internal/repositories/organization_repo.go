package repositories

import (
	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"gorm.io/gorm"
)

type OrganizationRepository struct {
	db *gorm.DB
}

func NewOrganizationRepository(db *gorm.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(org *models.Organization) error {
	return r.db.Create(org).Error
}

func (r *OrganizationRepository) FindAll() ([]models.Organization, error) {
	var orgs []models.Organization
	err := r.db.Find(&orgs).Error
	return orgs, err
}

func (r *OrganizationRepository) FindAllActive() ([]models.Organization, error) {
	var orgs []models.Organization
	err := r.db.Where("is_active = ?", true).Find(&orgs).Error
	return orgs, err
}

func (r *OrganizationRepository) FindByID(id uint64) (*models.Organization, error) {
	var org models.Organization
	err := r.db.First(&org, id).Error
	return &org, err
}

func (r *OrganizationRepository) FindByGithubOrgName(name string) (*models.Organization, error) {
	var org models.Organization
	err := r.db.Where("github_org_name = ?", name).First(&org).Error
	return &org, err
}

func (r *OrganizationRepository) Update(org *models.Organization) error {
	return r.db.Save(org).Error
}

func (r *OrganizationRepository) Delete(id uint64) error {
	return r.db.Delete(&models.Organization{}, id).Error
}
