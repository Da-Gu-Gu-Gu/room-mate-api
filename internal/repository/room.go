package repository

import (
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
	CREATE(payload *model.Room) error
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

func (r *roomRepository) CREATE(payload *model.Room) error {
	if err := r.db.Create(&payload).Error; err != nil {
		return err
	}
	return nil
}

func (r *roomRepository) UPDATE(payload *model.Room) error {
	if err := r.db.Save(&payload).Error; err != nil {
		return err
	}

	return nil
}
