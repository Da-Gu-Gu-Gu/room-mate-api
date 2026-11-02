package model

import (
	"time"

	"github.com/google/uuid"
)

type RoomBillShare struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	BillID    uuid.UUID `json:"billId" gorm:"type:uuid;not null;uniqueIndex:uq_room_bill_shares_bill_user"`
	UserID    uuid.UUID `json:"userId" gorm:"type:uuid;not null;uniqueIndex:uq_room_bill_shares_bill_user"`
	Amount    float64   `json:"amount" gorm:"type:numeric(12,2);not null;check:amount >= 0"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	Bill RoomBill `json:"-" gorm:"foreignKey:BillID;references:ID"`
	User User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
}
