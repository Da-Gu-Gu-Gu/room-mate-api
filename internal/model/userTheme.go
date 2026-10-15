package model

import "time"

type UserTheme struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"userId"`
	ThemeID   uint      `json:"themeId"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
