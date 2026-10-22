package dto

import (
	"time"

	"github.com/google/uuid"
)

type ThemeRequest struct {
	PrimaryColor     string  `json:"primaryColor" binding:"required"`
	SecondaryColor   string  `json:"secondaryColor" binding:"required"`
	BackgroundColor  string  `json:"backgroundColor" binding:"required"`
	BackgroundImgUrl *string `json:"backgroundImgUrl"`
}

type ThemeResponse struct {
	PrimaryColor    string    `json:"primaryColor"`
	SecondaryColor  string    `json:"secondaryColor"`
	BackgroundColor string    `json:"backgroundColor"`
	BackgroundImage *string   `json:"backgroundImage,omitempty"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt,omitempty"`
}

type UserThemeRequest struct {
	ThemeID uuid.UUID `json:"themeId" binding:"required"`
}

type UserThemeResponse struct {
	Theme     ThemeResponse `json:"theme"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

type UpdateThemeRequest struct {
	PrimaryColor     *string `json:"primaryColor"`
	SecondaryColor   *string `json:"secondaryColor"`
	BackgroundColor  *string `json:"backgroundColor"`
	BackgroundImgUrl *string `json:"backgroundImgUrl"`
}
