package model

import (
	"time"

	"github.com/google/uuid"
)

type RoomBill struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoomId           uuid.UUID  `json:"roomId" gorm:"type:uuid;not null"`
	CategoryId       uuid.UUID  `json:"categoryId" gorm:"type:uuid;not null"`
	Name             string     `json:"name" gorm:"not null"`
	Description      *string    `json:"description"`
	Total            float64    `json:"total" gorm:"type:numeric(12,2);not null;check:total >= 0"`
	Status           string     `json:"status" gorm:"type:text;not null;default:pending"`
	DueDate          *time.Time `json:"dueDate" gorm:"type:timestamptz"`
	CreatedBy        uuid.UUID  `json:"createdBy" gorm:"type:uuid;not null"`
	ReceiptImageKey  *string    `json:"-" gorm:"column:receipt_image_key"`
	ReceiptImageURL  string     `json:"receiptImageUrl,omitempty" gorm:"-"`
	CreatedAt        time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`

	Room         Room         `json:"room" gorm:"foreignKey:RoomId;references:ID"`
	BillCategory BillCategory `json:"billCategory" gorm:"foreignKey:CategoryId;references:ID"`
	Creator      User         `json:"creator" gorm:"foreignKey:CreatedBy;references:ID"`
}
