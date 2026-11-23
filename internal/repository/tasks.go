package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"room-mate-api/internal/model"
)

var (
	ErrTaskInvalid  = errors.New("invalid task request")
	ErrTaskConflict = errors.New("task occurrence is no longer pending")
)

type TaskRepository struct{ db *gorm.DB }

func NewTaskRepository(db *gorm.DB) *TaskRepository { return &TaskRepository{db: db} }

func taskActor(tx *gorm.DB, roomID, userID uuid.UUID) (*model.RoomMember, error) {
	var member model.RoomMember
	err := tx.Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, userID).First(&member).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRoomMemberForbidden
	}
	return &member, err
}

func validateTaskMembers(tx *gorm.DB, roomID uuid.UUID, members []model.TaskMember) error {
	ids := make([]uuid.UUID, len(members))
	for i, member := range members {
		ids[i] = member.MemberID
	}
	var count int64
	if err := tx.Model(&model.RoomMember{}).Where("room_id = ? AND id IN ? AND leave_at IS NULL", roomID, ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return ErrTaskInvalid
	}
	return nil
}

func (r *TaskRepository) Create(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := taskActor(tx, task.RoomID, task.CreatedBy); err != nil {
			return err
		}
		if err := validateTaskMembers(tx, task.RoomID, task.Members); err != nil {
			return err
		}
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		// Always create the first occurrence, including deadlines beyond the lookahead window.
		if err := createTaskOccurrence(tx, task, task.Members[0], task.FirstDueAt); err != nil {
			return err
		}
		return tx.Omit("Members").Save(task).Error
	})
}

// NextTaskDue advances calendar days in the task's time zone, preserving local wall time.
func NextTaskDue(due time.Time, frequency string, location *time.Location) time.Time {
	days := 1
	if frequency == "weekly" {
		days = 7
	}
	return due.In(location).AddDate(0, 0, days)
}

func createTaskOccurrence(tx *gorm.DB, task *model.Task, member model.TaskMember, due time.Time) error {
	occurrence := model.TaskOccurrence{
		ID: uuid.New(), TaskID: task.ID, DueAt: due,
		ExpiresAt:        due.Add(time.Duration(task.GracePeriodMinutes) * time.Minute),
		AssignedMemberID: member.MemberID, Status: "pending",
	}
	if err := tx.Create(&occurrence).Error; err != nil {
		return err
	}
	task.LastOrderNo = member.OrderNo
	task.NextDueAt = nil
	if task.ScheduleType == "recurring" {
		location, err := time.LoadLocation(task.Timezone)
		if err != nil {
			return err
		}
		next := NextTaskDue(due, *task.Frequency, location)
		if task.RepeatUntil == nil || !next.After(*task.RepeatUntil) {
			task.NextDueAt = &next
		}
	}
	return nil
}

// Sync creates up to seven days of upcoming turns. Row locks serialize workers and
// cursor advancement with inserts. Bounded batches catch up after downtime.
func (r *TaskRepository) Sync(ctx context.Context, now time.Time) error {
	if err := r.db.WithContext(ctx).Model(&model.TaskOccurrence{}).
		Where("status = ? AND expires_at <= ?", "pending", now).
		Updates(map[string]interface{}{"status": "missed", "updated_at": now}).Error; err != nil {
		return err
	}
	horizon := now.AddDate(0, 0, 7)
	var ids []uuid.UUID
	if err := r.db.WithContext(ctx).Model(&model.Task{}).
		Where("status = ? AND next_due_at <= ?", "active", horizon).
		Order("next_due_at, id").Limit(100).Pluck("id", &ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			var task model.Task
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&task, "id = ?", id).Error; err != nil {
				return err
			}
			if task.Status != "active" {
				return nil
			}
			var members []model.TaskMember
			if err := tx.Table("task_members AS tm").Select("tm.*").
				Joins("JOIN room_members rm ON rm.id = tm.member_id").
				Where("tm.task_id = ? AND rm.room_id = ? AND rm.leave_at IS NULL", id, task.RoomID).
				Order("tm.order_no").Scan(&members).Error; err != nil {
				return err
			}
			if len(members) == 0 {
				task.Status = "paused"
				return tx.Omit("Members").Save(&task).Error
			}
			for generated := 0; generated < 100 && task.NextDueAt != nil && !task.NextDueAt.After(horizon); generated++ {
				nextMember := members[0]
				for _, member := range members {
					if member.OrderNo > task.LastOrderNo {
						nextMember = member
						break
					}
				}
				if err := createTaskOccurrence(tx, &task, nextMember, *task.NextDueAt); err != nil {
					return err
				}
			}
			return tx.Omit("Members").Save(&task).Error
		}); err != nil {
			return err
		}
	}
	// Newly generated historical turns may already have expired.
	return r.db.WithContext(ctx).Model(&model.TaskOccurrence{}).
		Where("status = ? AND expires_at <= ?", "pending", now).
		Updates(map[string]interface{}{"status": "missed", "updated_at": now}).Error
}

func (r *TaskRepository) List(ctx context.Context, roomID, userID uuid.UUID, limit, offset int) ([]model.Task, error) {
	db := r.db.WithContext(ctx)
	if _, err := taskActor(db, roomID, userID); err != nil {
		return nil, err
	}
	tasks := []model.Task{}
	err := db.Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Order("order_no") }).
		Where("room_id = ?", roomID).Order("created_at DESC, id").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) Get(ctx context.Context, roomID, taskID, userID uuid.UUID) (*model.Task, error) {
	db := r.db.WithContext(ctx)
	if _, err := taskActor(db, roomID, userID); err != nil {
		return nil, err
	}
	var task model.Task
	err := db.Preload("Members", func(db *gorm.DB) *gorm.DB { return db.Order("order_no") }).
		Where("room_id = ? AND id = ?", roomID, taskID).First(&task).Error
	return &task, err
}

func (r *TaskRepository) Occurrences(ctx context.Context, roomID, taskID, userID uuid.UUID, limit, offset int) ([]model.TaskOccurrence, error) {
	if _, err := r.Get(ctx, roomID, taskID, userID); err != nil {
		return nil, err
	}
	occurrences := []model.TaskOccurrence{}
	err := r.db.WithContext(ctx).Where("task_id = ?", taskID).Order("due_at DESC, id").Limit(limit).Offset(offset).Find(&occurrences).Error
	return occurrences, err
}

func (r *TaskRepository) Complete(ctx context.Context, roomID, taskID, occurrenceID, userID uuid.UUID, proof *string, now time.Time) (*model.TaskOccurrence, error) {
	var occurrence model.TaskOccurrence
	expired := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		actor, err := taskActor(tx, roomID, userID)
		if err != nil {
			return err
		}
		var task model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("room_id = ? AND id = ?", roomID, taskID).First(&task).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND task_id = ?", occurrenceID, taskID).First(&occurrence).Error; err != nil {
			return err
		}
		// A retry from the same completing member returns the original completion.
		if occurrence.Status == "completed" && occurrence.CompletedByMemberID != nil && *occurrence.CompletedByMemberID == actor.ID {
			return nil
		}
		if occurrence.Status != "pending" || task.Status == "archived" {
			return ErrTaskConflict
		}
		if !now.Before(occurrence.ExpiresAt) {
			expired = true
			occurrence.Status = "missed"
		} else {
			occurrence.Status = "completed"
			occurrence.CompletedByMemberID = &actor.ID
			occurrence.CompletedAt = &now
			occurrence.ProofURL = proof
		}
		return tx.Save(&occurrence).Error
	})
	if err == nil && expired {
		err = ErrTaskConflict
	}
	return &occurrence, err
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, roomID, taskID, userID uuid.UUID, status string, now time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		actor, err := taskActor(tx, roomID, userID)
		if err != nil {
			return err
		}
		var task model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("room_id = ? AND id = ?", roomID, taskID).First(&task).Error; err != nil {
			return err
		}
		if task.CreatedBy != userID && actor.Role != model.RoleAdmin {
			return ErrRoomMemberForbidden
		}
		if task.Status == "archived" && status != "archived" {
			return ErrTaskConflict
		}
		if task.Status == "paused" && status == "active" && task.NextDueAt != nil {
			location, err := time.LoadLocation(task.Timezone)
			if err != nil {
				return err
			}
			// Paused dates are skipped without consuming rotation positions.
			for task.NextDueAt != nil && !task.NextDueAt.After(now) {
				next := NextTaskDue(*task.NextDueAt, *task.Frequency, location)
				task.NextDueAt = &next
				if task.RepeatUntil != nil && next.After(*task.RepeatUntil) {
					task.NextDueAt = nil
				}
			}
		}
		task.Status = status
		if status == "archived" {
			if err := tx.Model(&model.TaskOccurrence{}).Where("task_id = ? AND status = ?", taskID, "pending").
				Updates(map[string]interface{}{"status": "cancelled", "updated_at": now}).Error; err != nil {
				return err
			}
		}
		return tx.Omit("Members").Save(&task).Error
	})
}

func (r *TaskRepository) ReplaceMembers(ctx context.Context, roomID, taskID, userID uuid.UUID, members []model.TaskMember) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		actor, err := taskActor(tx, roomID, userID)
		if err != nil {
			return err
		}
		var task model.Task
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("room_id = ? AND id = ?", roomID, taskID).First(&task).Error; err != nil {
			return err
		}
		if task.CreatedBy != userID && actor.Role != model.RoleAdmin {
			return ErrRoomMemberForbidden
		}
		if task.Status == "archived" {
			return ErrTaskConflict
		}
		if task.ScheduleType == "one_time" {
			return ErrTaskInvalid
		}
		if err := validateTaskMembers(tx, roomID, members); err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&model.TaskMember{}).Error; err != nil {
			return err
		}
		if err := tx.Create(&members).Error; err != nil {
			return err
		}
		return tx.Model(&task).Update("last_order_no", 0).Error
	})
}
