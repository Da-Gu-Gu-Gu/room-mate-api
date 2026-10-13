package model

type UserTheme struct {
	ID      uint `json:"id"`
	UserID  uint `json:"userId"`
	ThemeID uint `json:"themeId"`
}
