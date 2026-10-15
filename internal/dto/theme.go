package dto

import "time"

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
	BackgroundImage *string   `json:"backgroundImage"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type UserThemeRequest struct {
	UserID  uint `json:"userId" binding:"required"`
	ThemeID uint `json:"themeId" binding:"required"`
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
