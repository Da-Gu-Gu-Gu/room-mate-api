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
}

func NewRoomRepository(db *gorm.DB) *roomRepository {
	return &roomRepository{
		db: db,
	}
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
	if err := r.db.First(&room, id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) DELETE(id uuid.UUID) error {
	err := r.db.Delete(&model.Room{}, id).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) Create(
	ctx context.Context,
	payload *model.Room,
) error {
	return r.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 1. Create the room.
			if err := tx.Create(payload).Error; err != nil {
				return err
			}

			// 2. Register the creator as a room member.
			joinedAt := time.Now()

			roomMember := &model.RoomMember{
				RoomID:   payload.ID,
				UserID:   payload.CreatedBy,
				Role:     "owner",
				JoinedAt: &joinedAt,
				// LeaveAt:  &joinedAt,
			}

			if err := tx.Create(roomMember).Error; err != nil {
				fmt.Print(err)
				return err
			}

			return nil
		},
	)
}

func (r *roomRepository) UPDATE(payload *model.Room) error {
	if err := r.db.Save(&payload).Error; err != nil {
		return err
	}

	return nil
}
