package model

import (
	"time"

	"github.com/google/uuid"
)

type RoomBillPayment struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BillShareID uuid.UUID  `json:"billShareId" gorm:"type:uuid;not null"`
	UserID      uuid.UUID  `json:"userId" gorm:"type:uuid;not null"`
	Status      string     `json:"status" gorm:"type:text;not null;default:pending"`
	Note        *string    `json:"note"`
	ProofURL    string     `json:"proofUrl" gorm:"type:varchar(255);not null"`
	Amount      float64    `json:"amount" gorm:"type:numeric(12,2);not null;check:amount > 0"`
	ReviewedBy  *uuid.UUID `json:"reviewedBy" gorm:"type:uuid"`
	ReviewedAt  *time.Time `json:"reviewedAt" gorm:"type:timestamptz"`
	CreatedAt   time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`

	BillShare RoomBillShare `json:"billShare" gorm:"foreignKey:BillShareID;references:ID"`
	User      User          `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Reviewer  *User         `json:"reviewer" gorm:"foreignKey:ReviewedBy;references:ID"`
}
