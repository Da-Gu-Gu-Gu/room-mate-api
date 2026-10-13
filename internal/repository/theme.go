package repository

import (
	"room-mate-api/internal/model"

	"gorm.io/gorm"
)

type themeRepository struct {
	db *gorm.DB
}

type ThemeRepository interface {
	FindById(id string) (*model.Theme, error)
}

func NewThemeRepository(db *gorm.DB) *themeRepository {
	return &themeRepository{
		db: db,
	}
}

func (r *themeRepository) FindById(id string) (*model.Theme, error) {
	var theme model.Theme
	if err := r.db.First(&theme, id).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}
