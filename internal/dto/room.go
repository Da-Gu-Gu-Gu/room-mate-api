package dto

import (
	"time"

	"github.com/google/uuid"
)

type RoomRequest struct {
	RoomName   string    `json:"roomName" binding:"required"`
	RoomNumber string    `json:"roomNumber" binding:"required"`
	CreatedBy  uuid.UUID `json:"createdBy" binding:"required"`
}

type RoomResponse struct {
	RoomName   string               `json:"roomName"`
	RoomCode   string               `json:"roomCode"`
	RoomMember []RoomMemberResponse `json:"roomMember,omitempty"`
	CreatedAt  time.Time            `json:"createdAt,omitempty"`
	UpdatedAt  time.Time            `json:"updatedAt,omitempty"`
}

type UpdateRoomRequest struct {
	RoomName   *string `json:"roomName"`
	RoomNumber *string `json:"roomNumber"`
}

type RoomMemberResponse struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	ProfileUrl string     `json:"profileUrl"`
	Email      string     `json:"email"`
	Role       string     `json:"role"`
	JoinedAt   time.Time  `json:"joinedAt"`
	LeaveAt    *time.Time `json:"leaveAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type RoomMemberRequest struct {
	RoomID uuid.UUID `json:"roomId" binding:"required"`
	UserID uuid.UUID `json:"userId" binding:"required"`
}

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
	Declined Status = "rejected"
)

type RoomInvitationRequest struct {
	RoomID    uuid.UUID `json:"roomId" binding:"required"`
	InviteeID uuid.UUID `json:"inviteeId" binding:"required"`
	Status    Status    `json:"status" binding:"required"`
	InviterID uuid.UUID `json:"inviterId" binding:"required"`
}

type RoomInvitationResponse struct {
	ID        uint      `json:"id"`
	RoomID    uint      `json:"roomId"`
	InviterID uint      `json:"inviterId"`
	InviteeID uint      `json:"inviteeId"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
