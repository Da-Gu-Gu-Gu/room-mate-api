package model

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	InvitationPending  Status = "pending"
	InvitationAccepted Status = "accepted"
	InvitationDeclined Status = "rejected"
)

type Invitation struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoomID    uuid.UUID `json:"roomId" gorm:"type:uuid;not null"`
	InviterID uuid.UUID `json:"inviterId" gorm:"type:uuid;not null"`
	InviteeID uuid.UUID `json:"inviteeId" gorm:"type:uuid;not null"`
	Status    Status    `json:"status" gorm:"type:status;not null;default:pending"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Invitation) TableName() string { return "room_invitations" }
