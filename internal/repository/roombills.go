package repository

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRoomBillNotFound              = errors.New("room bill not found")
	ErrRoomBillShareForbidden        = errors.New("user is not an active member of this room")
	ErrRoomBillShareTotal            = errors.New("assigned share amounts must equal the bill total")
	ErrInvalidRoomBillShares         = errors.New("room bill shares contain duplicate users or invalid amounts")
	ErrRoomBillSharesAlreadyAssigned = errors.New("shares have already been assigned for this bill")
	ErrRoomBillPaymentNotFound       = errors.New("room bill payment not found")
	ErrRoomBillPaymentNotPending     = errors.New("room bill payment is no longer pending")
	ErrRoomBillReviewForbidden       = errors.New("only the active room owner can review payments")
)

type RoomBillRepository interface {
	CanCreate(roomID, userID uuid.UUID) error
	Create(roomBill *model.RoomBill) (*model.RoomBill, error)
	GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error)
	GetRoomBillById(roomID, userID, billID uuid.UUID) (*model.RoomBill, error)
	CreateShares(roomID, billID, requesterID uuid.UUID, shares []model.RoomBillShare) ([]*model.RoomBillShare, error)
	GETShares(roomID, billID, userID uuid.UUID) ([]*model.RoomBillShare, error)
	CreatePayment(roomID, billID, userID uuid.UUID, payment *model.RoomBillPayment) (*model.RoomBillPayment, error)
	ReviewPayment(roomID, billID, paymentID, reviewerID uuid.UUID, status string) (*model.RoomBillPayment, error)
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

func (r *roomBillRepository) CanCreatePayment(billID, shareId, userID uuid.UUID) (*model.RoomBillShare, error) {
	var share *model.RoomBillShare
	err := r.db.Model(&model.RoomBillShare{}).Where("bill_id = ? AND id = ? AND user_id = ?", billID, shareId, userID).First(&share).Error
	if share.ID == uuid.Nil {
		return nil, err
	}
	return share, nil
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

			//check if the user is still an active member of the room
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

func (r *roomBillRepository) GETShares(roomID, billID, userID uuid.UUID) ([]*model.RoomBillShare, error) {
	ok, err := r.isBelongToRoom(roomID, userID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.New("user not belong to this room, cannot get room bill shares")
	}

	var shares []*model.RoomBillShare
	if err := r.db.Where("bill_id = ?", billID).Find(&shares).Error; err != nil {
		return nil, err
	}
	return shares, nil
}

func (r *roomBillRepository) CreatePayment(roomID, billID, userID uuid.UUID, payment *model.RoomBillPayment) (*model.RoomBillPayment, error) {
	if err := r.CanCreate(roomID, userID); err != nil {
		return nil, err
	}
	share, err := r.CanCreatePayment(billID, payment.BillShareID, userID)
	if err != nil {
		return nil, err
	}

	var totalPaidAmount float64

	if err := r.db.Model(&model.RoomBillPayment{}).
		Select("COALESCE(SUM(amount), 0)").
		Where(
			"bill_share_id = ? AND LOWER(status) IN ?",
			payment.BillShareID,
			[]string{"pending", "approved"},
		).
		Scan(&totalPaidAmount).Error; err != nil {
		return nil, err
	}

	totalCents := int64(math.Round(share.Amount * 100))
	totalPaidCents := int64(math.Round(totalPaidAmount * 100))
	paymentCents := int64(math.Round(payment.Amount * 100)) //cuurent pay amount in cent

	remainingCents := totalCents - totalPaidCents

	if paymentCents <= 0 || paymentCents > remainingCents {
		return nil, fmt.Errorf("invalid payment amount: must be greater than 0 and less than or equal to remaining amount of %.2f", float64(remainingCents)/100)
	}

	if err := r.db.Create(payment).Error; err != nil {
		return nil, err
	}
	return payment, nil
}

// ReviewPayment allows only the active room owner to approve or reject a
// pending payment. Locking the bill and payment rows serializes reviews.
func (r *roomBillRepository) ReviewPayment(roomID, billID, paymentID, reviewerID uuid.UUID, status string) (*model.RoomBillPayment, error) {
	var reviewedPayment model.RoomBillPayment

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Confirm the bill belongs to this room and serialize reviews for it.
		var bill model.RoomBill
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND room_id = ?", billID, roomID).
			First(&bill).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoomBillNotFound
			}
			return err
		}

		// Review permission is intentionally restricted to the bill creator.
		if reviewerID != bill.CreatedBy {
			return ErrRoomBillReviewForbidden
		}

		// Join through the share to ensure the payment belongs to this bill.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Joins("JOIN room_bill_shares ON room_bill_shares.id = room_bill_payments.bill_share_id").
			Where("room_bill_shares.bill_id = ? AND room_bill_payments.id = ?", billID, paymentID).
			First(&reviewedPayment).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoomBillPaymentNotFound
			}
			return err
		}

		if !strings.EqualFold(reviewedPayment.Status, "pending") {
			return ErrRoomBillPaymentNotPending
		}

		now := time.Now().UTC()
		if err := tx.Model(&reviewedPayment).Updates(map[string]interface{}{
			"status":      status,
			"reviewed_by": reviewerID,
			"reviewed_at": now,
			"updated_at":  now,
		}).Error; err != nil {
			return err
		}

		reviewedPayment.Status = status
		reviewedPayment.ReviewedBy = &reviewerID
		reviewedPayment.ReviewedAt = &now
		return tx.Preload("User").Preload("Reviewer").Preload("BillShare").
			First(&reviewedPayment, "id = ?", paymentID).Error
	})
	if err != nil {
		return nil, err
	}
	return &reviewedPayment, nil
}
