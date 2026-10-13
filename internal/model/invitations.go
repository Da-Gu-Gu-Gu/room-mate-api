package model

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
	Declined Status = "declined"
)

type Invitation struct {
	ID        uint   `json:"id"`
	RoomID    uint   `json:"roomId"`
	InviterID uint   `json:"inviterId"`
	InviteeID uint   `json:"inviteeId"`
	Status    Status `json:"status"`
	CreatedAt string `json:"createdAt"`
}
