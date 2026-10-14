package repository

import (
	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type themeRepository struct {
	db *gorm.DB
}

type ThemeRepository interface {
	FindById(id uuid.UUID) (*model.Theme, error)
	Create(theme *model.Theme) error
	FindAll() ([]*model.Theme, error)
	Delete(themeId uuid.UUID) error
	Update(themeId uuid.UUID, theme *model.Theme) error
}

func NewThemeRepository(db *gorm.DB) *themeRepository {
	return &themeRepository{
		db: db,
	}
}

func (r *themeRepository) FindById(id uuid.UUID) (*model.Theme, error) {
	var theme model.Theme
	if err := r.db.First(&theme, id).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *themeRepository) Create(theme *model.Theme) error {
	if err := r.db.Create(theme).Error; err != nil {
		return err
	}
	return nil
}

func (r *themeRepository) FindAll() ([]*model.Theme, error) {
	var themes []*model.Theme
	if err := r.db.Find(&themes).Error; err != nil {
		return nil, err
	}
	return themes, nil
}

func (r *themeRepository) Delete(themeId uuid.UUID) error {
	if err := r.db.Delete(&model.Theme{}, themeId).Error; err != nil {
		return err
	}
	return nil
}

func (r *themeRepository) Update(themId uuid.UUID, theme *model.Theme) error {
	if err := r.db.Save(theme).Error; err != nil {
		return err
	}
	return nil
}
