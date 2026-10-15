package model

import (
	"time"

	"github.com/google/uuid"
)

type Theme struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	PrimaryColor     string     `json:"primaryColor"`
	SecondaryColor   string     `json:"secondaryColor"`
	BackgroundColor  string     `json:"backgroundColor"`
	BackgroundImgUrl *string    `json:"backgroundImgUrl"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}
