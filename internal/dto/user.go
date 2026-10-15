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
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ProfileUrl string    `json:"profileUrl"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type UserUpdateRequest struct {
	Name       *string `json:"name"`
	Email      *string `json:"email" binding:"omitempty,email"`
	ProfileUrl *string `json:"profileUrl"`
}
