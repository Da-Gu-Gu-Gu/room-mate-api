package model

import (
	"time"

	"github.com/google/uuid"
)

type AuthIdentity struct {
	Id             uuid.UUID `json:"id" gorm:"type:uuid,primaryKey;default:gen_random_uuid()"`
	UserId         uuid.UUID `json:"userId" gorm:"type:uuid;not null"`
	Provider       string    `json:"provider" gorm:"type:varChar(20);not null"`
	ProviderUserId string    `json:"-" gorm:"type:varChar(20);not null;unique"`
	CreatedAt      time.Time `json:"createdAt" gorm:"autoCreatedTime"`

	User User `json:"-" gorm:"foreignKey:UserID;references:ID"`
}
