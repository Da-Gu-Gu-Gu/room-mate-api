package model

import "time"

type RoomBillShare struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	BillID    string    `json:"billId" gorm:"type:uuid;not null"`
	UserID    string    `json:"userId" gorm:"type:uuid;not null"`
	Amount    float64   `json:"amount" gorm:"type:numeric(12,2);not null;default:0"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`

	Bill RoomBill `json:"-" gorm:"foreignKey:BillID;references:ID"`
	User User     `json:"user" gorm:"foreignKey:UserID;references:ID"`
}
