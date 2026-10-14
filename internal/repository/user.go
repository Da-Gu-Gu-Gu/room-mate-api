package repository

import (
	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepository interface {
	GET() ([]*model.User, error)
	GetUser(id uuid.UUID) (*model.User, error)
	Create(user *model.User) error
	Update(id uuid.UUID, user *model.User) error
	Delete(id uuid.UUID) error
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

func (r *userRepository) GetUser(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *model.User) error {
	err := r.db.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Update(id uuid.UUID, user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}
