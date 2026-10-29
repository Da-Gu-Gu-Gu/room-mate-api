package model

import (
	"time"

	"github.com/google/uuid"
)

type RoleEnum = string

const (
	RoleAdmin  RoleEnum = "owner"
	RoleMember RoleEnum = "member"
)

type RoomMember struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	RoomID    uuid.UUID  `json:"roomId" gorm:"type:uuid;not null"`
	UserID    uuid.UUID  `json:"userId" gorm:"type:uuid;not null"`
	Role      RoleEnum   `json:"role" gorm:"type:role;not null"`
	JoinedAt  *time.Time `json:"joinedAt"`
	LeaveAt   *time.Time `json:"leaveAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`

	User User `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Room Room `json:"-" gorm:"foreignKey:RoomID;references:ID"`
}
