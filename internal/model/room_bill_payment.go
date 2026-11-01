package model

import "github.com/google/uuid"

type RoomBillPayment struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;PrimaryKey;default;gen_random_uuid()"`
	BillShareId uuid.UUID  `json:"billShareId" gorm:"type:uuid;not null"`
	UserId      uuid.UUID  `json:"userId" gorm:"type:uuid;not null"`
	Status      string     `json:"status" gorm:"type:text;not null;default:'pending'"`
	Note        *string    `json:"note" gorm:"type:text;nullable"`
	ProofUrl    *string    `json:"proofUrl" gorm:"type:text;not null;"`
	Amount      float64    `json:"amount" gorm:"type:numeric(12,2);not null;default:0"`
	ReviewedBy  *uuid.UUID `json:"reviewedBy" gorm:"type:uuid;nullable"`
	ReviewedAt  *string    `json:"reviewedAt" gorm:"type:timestamptz;nullable"`
	CreatedAt   string     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   string     `json:"updatedAt" gorm:"autoUpdateTime"`

	BillShare RoomBillShare `json:"billShare" gorm:"foreignKey:BillShareId;references:ID"`
	User      User          `json:"user" gorm:"foreignKey:UserId;references:ID"`
	Reviewer  *User         `json:"reviewer" gorm:"foreignKey:ReviewedBy;references:ID"`
}
