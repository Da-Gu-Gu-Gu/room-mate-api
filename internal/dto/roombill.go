package dto

import (
	"time"

	"github.com/google/uuid"
)

type RoomBillRequest struct {
	RoomID      uuid.UUID  `json:"roomId" binding:"required"`
	CategoryID  uuid.UUID  `json:"categoryId" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	Total       float64    `json:"total" binding:"required,gt=0"`
	Status      *string    `json:"status"`
	DueDate     *time.Time `json:"dueDate"`
	CreatedBy   uuid.UUID  `json:"createdBy" binding:"required"`
}
