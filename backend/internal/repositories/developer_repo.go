package repositories

import (
	"github.com/coding-jr/planto-reviewer/backend/internal/models"
	"gorm.io/gorm"
)

type DeveloperRepository struct {
	db *gorm.DB
}

func NewDeveloperRepository(db *gorm.DB) *DeveloperRepository {
	return &DeveloperRepository{db: db}
}

func (r *DeveloperRepository) Create(dev *models.Developer) error {
	return r.db.Create(dev).Error
}

func (r *DeveloperRepository) FindOrCreate(dev *models.Developer) error {
	return r.db.Where("organization_id = ? AND github_username = ?",
		dev.OrganizationID, dev.GithubUsername).
		FirstOrCreate(dev).Error
}

func (r *DeveloperRepository) FindByOrgID(orgID uint64) ([]models.Developer, error) {
	var devs []models.Developer
	err := r.db.Where("organization_id = ?", orgID).Find(&devs).Error
	return devs, err
}

func (r *DeveloperRepository) FindByID(id uint64) (*models.Developer, error) {
	var dev models.Developer
	err := r.db.First(&dev, id).Error
	return &dev, err
}

func (r *DeveloperRepository) Update(dev *models.Developer) error {
	return r.db.Save(dev).Error
}

func (r *DeveloperRepository) IncrementPRCount(devID uint64) error {
	return r.db.Model(&models.Developer{}).
		Where("id = ?", devID).
		UpdateColumn("total_prs", gorm.Expr("total_prs + ?", 1)).
		Error
}
