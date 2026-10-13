package dto

type ThemeRequest struct {
	PrimaryColor    string  `json:"primaryColor" binding:"required"`
	SecondaryColor  string  `json:"secondaryColor" binding:"required"`
	BackgroundColor string  `json:"backgroundColor" binding:"required"`
	BackgroundImage *string `json:"backgroundImage"`
}

type ThemeResponse struct {
	PrimaryColor    string  `json:"primaryColor"`
	SecondaryColor  string  `json:"secondaryColor"`
	BackgroundColor string  `json:"backgroundColor"`
	BackgroundImage *string `json:"backgroundImage"`
}

type UserThemeRequest struct {
	UserID  uint `json:"userId" binding:"required"`
	ThemeID uint `json:"themeId" binding:"required"`
}

type UserThemeResponse struct {
	Theme ThemeResponse `json:"theme"`
}
