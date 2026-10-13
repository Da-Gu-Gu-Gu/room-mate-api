package model

type Room struct {
	ID         uint   `json:"id"`
	RoomName   string `json:"roomName"`
	RoomCode   string `json:"roomCode"`
	RoomNumber string `json:"roomNumber"`
	CreatedBy  uint   `json:"createdBy"`
	CreatedAt  string `json:"createdAt"`
}
