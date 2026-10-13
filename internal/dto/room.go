package dto

type RoomRequest struct {
	RoomName  string `json:"roomName" binding:"required"`
	RoomCode  string `json:"roomCode" binding:"required"`
	CreatedBy uint   `json:"createdBy" binding:"required"`
}

type RoomResponse struct {
	RoomName   string               `json:"roomName"`
	RoomCode   string               `json:"roomCode"`
	Theme      ThemeResponse        `json:"theme"`
	RoomMember []RoomMemberResponse `json:"roomMember"`
}

type RoomMemberResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	ProfileUrl string `json:"profileUrl"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	JoinedAt   string `json:"joinedAt"`
	LeaveAt    string `json:"-"`
}

type RoomMemberRequest struct {
	RoomID uint `json:"roomId" binding:"required"`
	UserID uint `json:"userId" binding:"required"`
}
type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
	Declined Status = "declined"
)

type RoomInvitationRequest struct {
	RoomID    uint   `json:"roomId" binding:"required"`
	InviteeID uint   `json:"inviteeId" binding:"required"`
	Status    Status `json:"status" binding:"required"`
	InviterID uint   `json:"inviterId" binding:"required"`
}

type RoomInvitationResponse struct {
	ID        uint   `json:"id"`
	RoomID    uint   `json:"roomId"`
	InviterID uint   `json:"inviterId"`
	InviteeID uint   `json:"inviteeId"`
	Status    Status `json:"status"`
	CreatedAt string `json:"createdAt"`
}
