package repository

import (
	"context"
	"fmt"
	"room-mate-api/internal/model"
	"time"

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
	return r.db.Delete(&model.Room{}, "id = ?", id).Error
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
		if err := tx.Create(payload).Error; err != nil {
			return err
		}

		joinedAt := time.Now()
		roomMember := &model.RoomMember{
			RoomID: payload.ID,
			UserID: payload.CreatedBy,
			Role: "owner",
			JoinedAt: &joinedAt,
		}
		if err := tx.Create(roomMember).Error; err != nil {
			fmt.Print(err)
			return err
		}
		return nil
	})
}

func (r *roomRepository) UPDATE(payload *model.Room) error {
	return r.db.Save(payload).Error
}
