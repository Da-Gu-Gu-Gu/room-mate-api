package model

import "github.com/google/uuid"

type RoomBill struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoomId      uuid.UUID `json:"roomId" gorm:"type:uuid;not null"`
	CategoryId  uuid.UUID `json:"categoryId" gorm:"type:uuid;not null"`
	Name        string    `json:"name" gorm:"not null"`
	Description *string   `json:"description" gorm:"nullable"`
	Total       float64   `json:"total" gorm:"type:numeric(12,2);not null;default:0"`
	Status      string    `json:"status" gorm:"type:text;not null;default:'pending'"`
	DueDate     *string   `json:"dueDate" gorm:"type:timestamptz;nullable"`
	CreatedBy   uuid.UUID `json:"createdBy" gorm:"type:uuid;not null"`
	CreatedAt   string    `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   string    `json:"updatedAt" gorm:"autoUpdateTime"`

	Room         Room         `json:"room" gorm:"foreignKey:RoomId;references:ID"`
	BillCategory BillCategory `json:"billCategory" gorm:"foreignKey:CategoryId;references:ID"`
	Creator      User         `json:"creator" gorm:"foreignKey:CreatedBy;references:ID"`
}
