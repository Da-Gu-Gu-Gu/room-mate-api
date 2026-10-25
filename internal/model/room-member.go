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
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;PrimaryKey;Default:gen_random_uuid()"`
	RoomID    uuid.UUID  `json:"roomId" gorm:"type:uuid;not null"`
	UserID    uuid.UUID  `json:"userId" gorm:"type:uuid;not null"`
	Role      RoleEnum   `json:"role"`
	JoinedAt  *time.Time `json:"joinedAt"`
	LeaveAt   *time.Time `json:"leaveAt"`
	CreatedAt time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`

	User User `json:"-" gorm:"foreignKey:UserID;references:ID;"`
	Room Room `json:"-" gorm:"foreignKey:RoomID;references:ID;"`
}
