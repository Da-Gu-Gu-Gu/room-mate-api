package dto

import "github.com/google/uuid"

//need to set up for register and login and jwt

type UserRequest struct {
	Name       string  `json:"name" binding:"required"`
	Email      string  `json:"email" binding:"required"`
	ProfileUrl *string `json:"profileUrl" `
}

type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ProfileUrl string    `json:"profileUrl"`
	Email      string    `json:"email"`
	CreatedAt  string    `json:"createdAt"`
}

type UserUpdateRequest struct {
	Name       *string `json:"name" binding:"required"`
	Email      *string `json:"email" binding:"required"`
	ProfileUrl *string `json:"profileUrl" `
}
