package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

var ErrRoomBillStorageUnavailable = errors.New("room bill image storage is not configured")

type RoomBillObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
	PresignGet(ctx context.Context, key string) (string, error)
}

type RoomBillService struct {
	roomBillRepository repository.RoomBillRepository
	objectStorage      RoomBillObjectStorage
}

func NewRoomBillService(roomBillRepo repository.RoomBillRepository, objectStorage RoomBillObjectStorage) *RoomBillService {
	return &RoomBillService{
		roomBillRepository: roomBillRepo,
		objectStorage:      objectStorage,
	}
}

func (s *RoomBillService) CreateRoomBill(req *dto.RoomBillRequest, receiptImage []byte, contentType, extension string) (*model.RoomBill, error) {
	roomBill := &model.RoomBill{
		ID:          uuid.New(),
		RoomId:      req.RoomID,
		CategoryId:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Total:       req.Total,
		Status:      "pending",
		DueDate:     req.DueDate,
		CreatedBy:   req.CreatedBy,
	}

	// Check room access before uploading anything to S3. Create() checks again
	// immediately before inserting the row.
	if err := s.roomBillRepository.CanCreate(req.RoomID, req.CreatedBy); err != nil {
		return nil, err
	}

	if len(receiptImage) > 0 {
		if s.objectStorage == nil {
			return nil, ErrRoomBillStorageUnavailable
		}

		key := fmt.Sprintf("room-bills/%s/%s/receipt.%s", req.RoomID, roomBill.ID, extension)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.objectStorage.Upload(ctx, key, bytes.NewReader(receiptImage), contentType); err != nil {
			return nil, err
		}
		roomBill.ReceiptImageKey = key
	}

	createdRoomBill, err := s.roomBillRepository.Create(roomBill)
	if err != nil {
		if roomBill.ReceiptImageKey != "" && s.objectStorage != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.objectStorage.Delete(ctx, roomBill.ReceiptImageKey)
		}
		return nil, err
	}

	// The bill is already persisted. If signing the preview URL fails, do not
	// report the creation as failed; the next GET can generate a fresh URL.
	_ = s.addReceiptImageURL(createdRoomBill)
	return createdRoomBill, nil
}

func (s *RoomBillService) GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error) {
	bills, err := s.roomBillRepository.GetRoomBills(roomID, userID)
	if err != nil {
		return nil, err
	}
	for _, bill := range bills {
		if err := s.addReceiptImageURL(bill); err != nil {
			return nil, err
		}
	}
	return bills, nil
}

func (s *RoomBillService) GetRoomBillByID(roomID, billID, userID uuid.UUID) (*model.RoomBill, error) {
	bill, err := s.roomBillRepository.GetRoomBillById(roomID, userID, billID)
	if err != nil {
		return nil, err
	}
	if err := s.addReceiptImageURL(bill); err != nil {
		return nil, err
	}
	return bill, nil
}

func (s *RoomBillService) addReceiptImageURL(bill *model.RoomBill) error {
	if bill.ReceiptImageKey == "" {
		return nil
	}
	if s.objectStorage == nil {
		return ErrRoomBillStorageUnavailable
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url, err := s.objectStorage.PresignGet(ctx, bill.ReceiptImageKey)
	if err != nil {
		return err
	}
	bill.ReceiptImageURL = url
	return nil
}

func (s *RoomBillService) CreateRoomBillShares(roomID, billID, requesterID uuid.UUID, req *dto.CreateRoomBillSharesRequest) ([]*model.RoomBillShare, error) {
	shares := make([]model.RoomBillShare, 0, len(req.Members))
	for _, member := range req.Members {
		shares = append(shares, model.RoomBillShare{
			UserID: member.UserID,
			Amount: member.Amount,
		})
	}
	return s.roomBillRepository.CreateShares(roomID, billID, requesterID, shares)
}
