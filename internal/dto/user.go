package dto

import (
	"time"

	"github.com/google/uuid"
)

// UserRequest is the payload for creating a user.
// Server-managed timestamps are intentionally not accepted from clients.
type UserRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required,email"`
	ProfileUrl *string `json:"profileUrl"`
}

type UserResponse struct {
	ID         uuid.UUID `json:"id,omitempty"`
	Name       string    `json:"name"`
	ProfileUrl string    `json:"profileUrl,omitempty"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

type UserUpdateRequest struct {
	Name       *string `json:"name"`
	Email      *string `json:"email" binding:"omitempty,email"`
	ProfileUrl *string `json:"profileUrl"`
}

type MeResponse struct {
	User  UserResponse   `json:"user"`
	Room  *RoomResponse  `json:"room,omitempty"`
	Theme *ThemeResponse `json:"theme,omitempty"`
}
