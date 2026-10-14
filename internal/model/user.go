package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name       string    `json:"name"`
	ProfileUrl string    `json:"profileUrl"`
	Email      string    `json:"email"`
	CreatedAt  time.Time `json:"createdAt"`
}
