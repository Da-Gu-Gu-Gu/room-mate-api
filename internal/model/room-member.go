package model

import "time"

type RoleEnum = string

const (
	RoleAdmin  RoleEnum = "admin"
	RoleMember RoleEnum = "member"
)

type RoomMember struct {
	ID        uint       `json:"id"`
	RoomID    uint       `json:"roomId"`
	UserID    uint       `json:"userId"`
	Role      RoleEnum   `json:"role"`
	JoinedAt  *time.Time `json:"joinedAt"`
	LeaveAt   *time.Time `json:"leaveAt"`
	CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}
