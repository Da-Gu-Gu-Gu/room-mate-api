package repository

import (
	"errors"
	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomBillRepository interface {
	Create(roomBill *model.RoomBill) (*model.RoomBill, error)
	GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error)
	GetRoomBillById(roomID, userID, billID uuid.UUID) (*model.RoomBill, error)
	// Define methods for RoomBillRepository here
}

type roomBillRepository struct {
	db *gorm.DB
}

func NewRoomBillRepository(db *gorm.DB) *roomBillRepository {
	return &roomBillRepository{db: db}
}

func (r *roomBillRepository) isBelongToRoom(roomID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *roomBillRepository) Create(roomBill *model.RoomBill) (*model.RoomBill, error) {
	ok, err := r.isBelongToRoom(roomBill.RoomId, roomBill.CreatedBy)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not belong to this room, cannot create room bill")
	}
	if err := r.db.Create(roomBill).Error; err != nil {
		return nil, err
	}
	return roomBill, nil
}

func (r *roomBillRepository) GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error) {
	ok, err := r.isBelongToRoom(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not belong to this room, cannot create room bill")
	}
	var bills []*model.RoomBill
	if err := r.db.Where("room_id = ?", roomID).Find(&bills).Error; err != nil {
		return nil, err
	}
	return bills, nil
}

func (r *roomBillRepository) GetRoomBillById(roomID, userID, billID uuid.UUID) (*model.RoomBill, error) {
	ok, err := r.isBelongToRoom(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not belong to this room, cannot create room bill")
	}
	var bill model.RoomBill
	if err := r.db.Where("room_id = ? AND id = ?", roomID, billID).First(&bill).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("room bill not found")
		}
		return nil, err
	}
	return &bill, nil
}
