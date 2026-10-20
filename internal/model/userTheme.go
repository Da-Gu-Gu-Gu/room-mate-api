package model

import (
	"time"

	"github.com/google/uuid"
)

type UserTheme struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null;unique"`
	ThemeID   uuid.UUID `json:"themeId" gorm:"type:uuid;not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	User  User  `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Theme Theme `json:"theme" gorm:"foreignKey:ThemeID;references:ID"`
}

func (UserTheme) TableName() string {
	return "users-themes"
}
