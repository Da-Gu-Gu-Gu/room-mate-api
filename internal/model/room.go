package model

import "time"

type Room struct {
	ID         uint      `json:"id"`
	RoomName   string    `json:"roomName"`
	RoomCode   string    `json:"roomCode"`
	RoomNumber string    `json:"roomNumber"`
	CreatedBy  uint      `json:"createdBy"`
	CreatedAt  time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
