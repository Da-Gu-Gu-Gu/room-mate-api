package model

import "time"

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
	Declined Status = "declined"
)

type Invitation struct {
	ID        uint      `json:"id"`
	RoomID    uint      `json:"roomId"`
	InviterID uint      `json:"inviterId"`
	InviteeID uint      `json:"inviteeId"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}
