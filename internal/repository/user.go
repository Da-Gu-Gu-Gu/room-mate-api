package repository

import (
	"room-mate-api/internal/model"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepository interface {
	GET() ([]*model.User, error)
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GET() ([]*model.User, error) {
	var userList []*model.User
	if err := r.db.Find(&userList).Error; err != nil {
		return nil, err
	}
	return userList, nil

}
