package model

import "time"

type RoleEnum = string

const (
	RoleAdmin  RoleEnum = "admin"
	RoleMember RoleEnum = "member"
)

type RoomMember struct {
	ID       uint       `json:"id"`
	RoomID   uint       `json:"roomId"`
	UserID   uint       `json:"userId"`
	Role     RoleEnum   `json:"role"`
	joinedAt *time.Time `json:"-"`
	leaveAt  *time.Time `json:"-"`
}
