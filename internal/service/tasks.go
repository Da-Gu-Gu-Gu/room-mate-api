package service

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
)

type TaskService struct{ Repository *repository.TaskRepository }

func NewTaskService(repo *repository.TaskRepository) *TaskService {
	return &TaskService{Repository: repo}
}

func taskMembers(taskID uuid.UUID, ids []uuid.UUID) ([]model.TaskMember, error) {
	if len(ids) == 0 || len(ids) > 100 {
		return nil, fmt.Errorf("%w: select 1 to 100 members", repository.ErrTaskInvalid)
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	members := make([]model.TaskMember, 0, len(ids))
	for index, id := range ids {
		if id == uuid.Nil || seen[id] {
			return nil, fmt.Errorf("%w: member IDs must be unique", repository.ErrTaskInvalid)
		}
		seen[id] = true
		members = append(members, model.TaskMember{ID: uuid.New(), TaskID: taskID, MemberID: id, OrderNo: index + 1})
	}
	return members, nil
}

func buildTask(roomID, userID uuid.UUID, req dto.CreateTaskRequest, now time.Time) (*model.Task, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len([]rune(req.Name)) > 200 || req.FirstDueAt.IsZero() || !req.FirstDueAt.After(now) {
		return nil, fmt.Errorf("%w: name and a future firstDueAt are required", repository.ErrTaskInvalid)
	}
	if req.Timezone == "" || req.Timezone == "Local" {
		return nil, fmt.Errorf("%w: timezone is required", repository.ErrTaskInvalid)
	}
	if _, err := time.LoadLocation(req.Timezone); err != nil {
		return nil, fmt.Errorf("%w: unknown timezone", repository.ErrTaskInvalid)
	}
	switch req.ScheduleType {
	case "one_time":
		if req.Frequency != nil || req.RepeatUntil != nil || len(req.MemberIDs) != 1 {
			return nil, fmt.Errorf("%w: one-time tasks require one member and no recurrence", repository.ErrTaskInvalid)
		}
	case "recurring":
		if req.Frequency == nil || (*req.Frequency != "daily" && *req.Frequency != "weekly") {
			return nil, fmt.Errorf("%w: frequency must be daily or weekly", repository.ErrTaskInvalid)
		}
	default:
		return nil, fmt.Errorf("%w: unknown scheduleType", repository.ErrTaskInvalid)
	}
	if req.RepeatUntil != nil && req.RepeatUntil.Before(req.FirstDueAt) {
		return nil, fmt.Errorf("%w: repeatUntil precedes firstDueAt", repository.ErrTaskInvalid)
	}
	grace := 1440
	if req.GracePeriodMinutes != nil {
		grace = *req.GracePeriodMinutes
	}
	if grace < 0 || grace > 10080 {
		return nil, fmt.Errorf("%w: gracePeriodMinutes must be between 0 and 10080", repository.ErrTaskInvalid)
	}
	task := &model.Task{ID: uuid.New(), RoomID: roomID, CreatedBy: userID, Name: req.Name,
		Description: req.Description, ScheduleType: req.ScheduleType, FirstDueAt: req.FirstDueAt,
		RepeatUntil: req.RepeatUntil, Frequency: req.Frequency, Timezone: req.Timezone,
		GracePeriodMinutes: grace, Status: "active"}
	var err error
	task.Members, err = taskMembers(task.ID, req.MemberIDs)
	return task, err
}

func (s *TaskService) Create(ctx context.Context, roomID, userID uuid.UUID, req dto.CreateTaskRequest) (*model.Task, error) {
	now := time.Now()
	task, err := buildTask(roomID, userID, req, now)
	if err != nil {
		return nil, err
	}
	err = s.Repository.Create(ctx, task)
	return task, err
}

func (s *TaskService) Complete(ctx context.Context, roomID, taskID, occurrenceID, userID uuid.UUID, req dto.CompleteTaskRequest) (*model.TaskOccurrence, error) {
	if req.ProofURL != nil {
		value := strings.TrimSpace(*req.ProofURL)
		parsed, err := url.Parse(value)
		if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil || len(value) > 2048 {
			return nil, fmt.Errorf("%w: proofUrl must be an HTTPS URL", repository.ErrTaskInvalid)
		}
		req.ProofURL = &value
	}
	return s.Repository.Complete(ctx, roomID, taskID, occurrenceID, userID, req.ProofURL, time.Now())
}

func (s *TaskService) ReplaceMembers(ctx context.Context, roomID, taskID, userID uuid.UUID, ids []uuid.UUID) error {
	members, err := taskMembers(taskID, ids)
	if err != nil {
		return err
	}
	return s.Repository.ReplaceMembers(ctx, roomID, taskID, userID, members)
}

func (s *TaskService) RunScheduler(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := s.Repository.Sync(ctx, time.Now()); err != nil && ctx.Err() == nil {
			log.Printf("task scheduler: %v", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
