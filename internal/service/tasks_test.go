package service

import (
	"github.com/google/uuid"
	"room-mate-api/internal/dto"
	"testing"
	"time"
)

func TestBuildTaskValidation(t *testing.T) {
	now := time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC)
	frequency := "weekly"
	valid := dto.CreateTaskRequest{Name: "Kitchen", ScheduleType: "recurring", FirstDueAt: now.Add(time.Hour), Frequency: &frequency, Timezone: "Asia/Bangkok", MemberIDs: []uuid.UUID{uuid.New(), uuid.New()}}
	task, err := buildTask(uuid.New(), uuid.New(), valid, now)
	if err != nil || task.GracePeriodMinutes != 1440 || task.Members[1].OrderNo != 2 {
		t.Fatalf("valid task: %+v %v", task, err)
	}
	for _, tc := range []struct {
		name   string
		change func(*dto.CreateTaskRequest)
	}{
		{"duplicate", func(r *dto.CreateTaskRequest) { r.MemberIDs = []uuid.UUID{r.MemberIDs[0], r.MemberIDs[0]} }},
		{"one time recurrence", func(r *dto.CreateTaskRequest) { r.ScheduleType = "one_time" }},
		{"missing frequency", func(r *dto.CreateTaskRequest) { r.Frequency = nil }},
		{"invalid timezone", func(r *dto.CreateTaskRequest) { r.Timezone = "invalid" }},
		{"past due", func(r *dto.CreateTaskRequest) { r.FirstDueAt = now.Add(-time.Hour) }},
		{"early cutoff", func(r *dto.CreateTaskRequest) { r.RepeatUntil = &now }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := valid
			tc.change(&req)
			if _, err := buildTask(uuid.New(), uuid.New(), req, now); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
