package dto

import (
	"time"

	"github.com/google/uuid"
)

type RoomBillRequest struct {
	RoomID          uuid.UUID  `json:"roomId"`
	CategoryID      uuid.UUID  `json:"categoryId" binding:"required"`
	Name            string     `json:"name" binding:"required"`
	Description     *string    `json:"description"`
	Total           float64    `json:"total" binding:"required,gt=0"`
	Status          *string    `json:"status"`
	DueDate         *time.Time `json:"dueDate"`
	ReceiptImageURL *string    `json:"receiptImageUrl"`
	CreatedBy       uuid.UUID  `json:"createdBy"`
}

type CreateRoomBillSharesRequest struct {
	Members []RoomBillShareItem `json:"members" binding:"required,min=1,dive"`
}

type RoomBillShareItem struct {
	UserID uuid.UUID `json:"userId" binding:"required"`
	Amount float64   `json:"amount" binding:"required,gt=0"`
}

type CreateRoomBillPaymentRequest struct {
	BillShareID uuid.UUID `json:"billShareId" binding:"required"`
	Note        *string   `json:"note"`
	ProofURL    string    `json:"proofUrl" binding:"required"`
	Amount      float64   `json:"amount" binding:"required,gt=0"`
}
