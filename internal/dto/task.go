package dto

import (
	"github.com/google/uuid"
	"time"
)

type CreateTaskRequest struct {
	Name               string      `json:"name"`
	Description        *string     `json:"description"`
	ScheduleType       string      `json:"scheduleType"`
	FirstDueAt         time.Time   `json:"firstDueAt"`
	RepeatUntil        *time.Time  `json:"repeatUntil"`
	Frequency          *string     `json:"frequency"`
	Timezone           string      `json:"timezone"`
	GracePeriodMinutes *int        `json:"gracePeriodMinutes"`
	MemberIDs          []uuid.UUID `json:"memberIds"`
}

type CompleteTaskRequest struct {
	ProofURL *string `json:"proofUrl"`
}
