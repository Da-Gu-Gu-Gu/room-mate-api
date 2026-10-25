package repository

import (
	"fmt"
	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomMemberRepository interface {
	GetRoomMembers(id uuid.UUID) ([]*model.RoomMember, error)
}

type roomMemberRepository struct {
	db *gorm.DB
}

func NewRoomMemberRepository(db *gorm.DB) *roomMemberRepository {
	return &roomMemberRepository{
		db: db,
	}
}

func (r *roomMemberRepository) GetRoomMembers(id uuid.UUID) ([]*model.RoomMember, error) {
	var roomMemberList []*model.RoomMember
	err := r.db.Preload("User").Preload("Room").Where("room_id = ? AND leave_at IS NULL ", id).Order("joined_at ASC").Find(&roomMemberList).Error
	if err != nil {
		return nil, err
	}
	fmt.Print(&roomMemberList)
	return roomMemberList, nil
}
