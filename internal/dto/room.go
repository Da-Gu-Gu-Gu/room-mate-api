package dto

import "github.com/google/uuid"

type RoomRequest struct {
	RoomName  string    `json:"roomName" binding:"required"`
	RoomCode  string    `json:"roomCode" binding:"required"`
	CreatedBy uuid.UUID `json:"createdBy" binding:"required"`
}

type RoomResponse struct {
	RoomName   string               `json:"roomName"`
	RoomCode   string               `json:"roomCode"`
	Theme      ThemeResponse        `json:"theme"`
	RoomMember []RoomMemberResponse `json:"roomMember"`
}

type RoomMemberResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ProfileUrl string    `json:"profileUrl"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	JoinedAt   string    `json:"joinedAt"`
	LeaveAt    string    `json:"-"`
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
	ID        uuid.UUID `json:"id"`
	RoomID    uuid.UUID `json:"roomId"`
	InviterID uuid.UUID `json:"inviterId"`
	InviteeID uuid.UUID `json:"inviteeId"`
	Status    Status    `json:"status"`
	CreatedAt string    `json:"createdAt"`
}
