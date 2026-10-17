package model

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoomName   string    `json:"roomName" gorm:"not null"`
	RoomCode   string    `json:"roomCode" gorm:"not null"`
	RoomNumber string    `json:"roomNumber" gorm:"not null"`
	CreatedBy  uuid.UUID `json:"createdBy" gorm:"not null;type:uuid"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
