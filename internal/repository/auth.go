package repository

import (
	"room-mate-api/internal/model"

	"gorm.io/gorm"
)

type authRepo struct {
	db *gorm.DB
}

type AuthRepo interface {
	FindByProvider(provider string,
		providerUserID string,
	) (*model.AuthIdentity, error)
	Create(payload *model.AuthIdentity) error
}

func NewAuthRepo(db *gorm.DB) *authRepo {
	return &authRepo{
		db: db,
	}
}

func (r *authRepo) FindByProvider(provider string,
	providerUserID string,
) (*model.AuthIdentity, error) {
	var authIdentity *model.AuthIdentity
	err := r.db.Preload("User").Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&authIdentity).Error
	if err != nil {
		return nil, err
	}
	return authIdentity, nil
}

func (r *authRepo) Create(
	identity *model.AuthIdentity,
) error {
	return r.db.Create(identity).Error
}
