package repository

import (
	"errors"
	"fmt"
	"math"

	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRoomBillNotFound       = errors.New("room bill not found")
	ErrRoomBillShareForbidden = errors.New("user is not an active member of this room")
	ErrRoomBillShareTotal     = errors.New("assigned share amounts must equal the bill total")
	ErrInvalidRoomBillShares  = errors.New("room bill shares contain duplicate users or invalid amounts")
	ErrRoomBillSharesAlreadyAssigned = errors.New("shares have already been assigned for this bill")
)

type RoomBillRepository interface {
	CanCreate(roomID, userID uuid.UUID) error
	Create(roomBill *model.RoomBill) (*model.RoomBill, error)
	GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error)
	GetRoomBillById(roomID, userID, billID uuid.UUID) (*model.RoomBill, error)
	CreateShares(roomID, billID, requesterID uuid.UUID, shares []model.RoomBillShare) ([]*model.RoomBillShare, error)
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

func (r *roomBillRepository) CanCreate(roomID, userID uuid.UUID) error {
	ok, err := r.isBelongToRoom(roomID, userID)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("user not belong to this room, cannot create room bill")
	}
	return nil
}

func (r *roomBillRepository) Create(roomBill *model.RoomBill) (*model.RoomBill, error) {
	if err := r.CanCreate(roomBill.RoomId, roomBill.CreatedBy); err != nil {
		return nil, err
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

func (r *roomBillRepository) CreateShares(roomID, billID, requesterID uuid.UUID, shares []model.RoomBillShare) ([]*model.RoomBillShare, error) {
	createdShares := make([]*model.RoomBillShare, 0, len(shares))

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Lock the bill to prevent two requests from assigning shares simultaneously.
		var bill model.RoomBill
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND room_id = ?", billID, roomID).
			First(&bill).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoomBillNotFound
			}
			return err
		}

		var requesterCount int64
		if err := tx.Model(&model.RoomMember{}).
			Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, requesterID).
			Count(&requesterCount).Error; err != nil {
			return err
		}
		if requesterCount == 0 {
			return ErrRoomBillShareForbidden
		}

		// A bill must be split among unique, currently active members of this room.
		seenUsers := make(map[uuid.UUID]struct{}, len(shares))
		var assignedCents int64
		for i := range shares {
			share := &shares[i]
			if share.UserID == uuid.Nil || share.Amount <= 0 || math.IsNaN(share.Amount) || math.IsInf(share.Amount, 0) {
				return ErrInvalidRoomBillShares
			}
			if _, exists := seenUsers[share.UserID]; exists {
				return ErrInvalidRoomBillShares
			}
			seenUsers[share.UserID] = struct{}{}
			assignedCents += int64(math.Round(share.Amount * 100))

			var memberCount int64
			if err := tx.Model(&model.RoomMember{}).
				Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, share.UserID).
				Count(&memberCount).Error; err != nil {
				return err
			}
			if memberCount == 0 {
				return fmt.Errorf("%w: %s", ErrRoomBillShareForbidden, share.UserID)
			}
		}

		billCents := int64(math.Round(bill.Total * 100))
		if assignedCents != billCents {
			return ErrRoomBillShareTotal
		}

		// Do not allow shares to be assigned twice for the same bill.
		var existingCount int64
		if err := tx.Model(&model.RoomBillShare{}).
			Where("bill_id = ?", billID).
			Count(&existingCount).Error; err != nil {
			return err
		}
		if existingCount > 0 {
			return ErrRoomBillSharesAlreadyAssigned
		}

		// Save each member's share individually. The surrounding transaction
		// guarantees all rows are inserted or none are inserted.
		for i := range shares {
			shares[i].BillID = billID
			share := shares[i]
			if err := tx.Create(&share).Error; err != nil {
				return err
			}
			createdShares = append(createdShares, &share)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return createdShares, nil
}
