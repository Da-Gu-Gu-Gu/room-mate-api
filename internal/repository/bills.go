package repository

import (
	"room-mate-api/internal/model"

	"gorm.io/gorm"
)

type billRepository struct {
	db *gorm.DB
}

type BillRepository interface {
	GetBillCategories() ([]model.BillCategory, error)
}

func NewBillRepository(db *gorm.DB) *billRepository {
	return &billRepository{db: db}
}

func (r *billRepository) GetBillCategories() ([]model.BillCategory, error) {
	var billCategories []model.BillCategory
	err := r.db.Find(&billCategories).Error
	if err != nil {
		return nil, err
	}
	return billCategories, nil
}
