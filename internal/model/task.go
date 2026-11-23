package model

import (
	"github.com/google/uuid"
	"time"
)

type Task struct {
	ID                 uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	RoomID             uuid.UUID    `json:"roomId"`
	Name               string       `json:"name"`
	Description        *string      `json:"description"`
	ScheduleType       string       `json:"scheduleType"`
	CreatedBy          uuid.UUID    `json:"createdBy"`
	FirstDueAt         time.Time    `json:"firstDueAt"`
	RepeatUntil        *time.Time   `json:"repeatUntil"`
	Frequency          *string      `json:"frequency"`
	Timezone           string       `json:"timezone"`
	GracePeriodMinutes int          `json:"gracePeriodMinutes"`
	Status             string       `json:"status"`
	NextDueAt          *time.Time   `json:"-"`
	LastOrderNo        int          `json:"-"`
	CreatedAt          time.Time    `json:"createdAt"`
	UpdatedAt          time.Time    `json:"updatedAt"`
	Members            []TaskMember `json:"members" gorm:"foreignKey:TaskID"`
}

type TaskMember struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	TaskID    uuid.UUID `json:"taskId"`
	MemberID  uuid.UUID `json:"memberId"`
	OrderNo   int       `json:"orderNo"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type TaskOccurrence struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey"`
	TaskID              uuid.UUID  `json:"taskId"`
	DueAt               time.Time  `json:"dueAt"`
	ExpiresAt           time.Time  `json:"expiresAt"`
	AssignedMemberID    uuid.UUID  `json:"assignedMemberId"`
	Status              string     `json:"status"`
	CompletedByMemberID *uuid.UUID `json:"completedByMemberId"`
	CompletedAt         *time.Time `json:"completedAt"`
	ProofURL            *string    `json:"proofUrl"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}
