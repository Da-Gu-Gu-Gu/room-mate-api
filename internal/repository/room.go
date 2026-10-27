package repository

import (
	"context"
	"time"

	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type roomRepository struct {
	db *gorm.DB
}

type RoomRepository interface {
	GET() ([]*model.Room, error)
	GETBYID(id uuid.UUID) (*model.Room, error)
	DELETE(id uuid.UUID) error
	Create(ctx context.Context, payload *model.Room) error
	UPDATE(payload *model.Room) error
	IsActiveMember(roomID, userID uuid.UUID) (bool, error)
}

func NewRoomRepository(db *gorm.DB) *roomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) GET() ([]*model.Room, error) {
	var rooms []*model.Room
	if err := r.db.Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) GETBYID(id uuid.UUID) (*model.Room, error) {
	var room model.Room
	if err := r.db.First(&room, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) DELETE(id uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := tx.Model(&model.RoomMember{}).
			Where("room_id = ? AND leave_at IS NULL", id).
			Updates(map[string]interface{}{"leave_at": now, "updated_at": now}).Error; err != nil {
			return err
		}
		if err := tx.Where("room_id = ?", id).Delete(&model.Invitation{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Room{}, "id = ?", id).Error
	})
}

func (r *roomRepository) IsActiveMember(roomID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *roomRepository) Create(ctx context.Context, payload *model.Room) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Enforce one current room per user. The partial unique index in migration
		// 000011 also protects against concurrent requests.
		var activeCount int64
		if err := tx.Model(&model.RoomMember{}).
			Where("user_id = ? AND leave_at IS NULL", payload.CreatedBy).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrAlreadyInRoom
		}

		if err := tx.Create(payload).Error; err != nil {
			return err
		}

		joinedAt := time.Now().UTC()
		member := &model.RoomMember{
			RoomID: payload.ID,
			UserID: payload.CreatedBy,
			Role: model.RoleAdmin,
			JoinedAt: &joinedAt,
		}
		return tx.Create(member).Error
	})
}

func (r *roomRepository) UPDATE(payload *model.Room) error {
	return r.db.Save(payload).Error
}
