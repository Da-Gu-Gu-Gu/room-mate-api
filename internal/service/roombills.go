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
	PublicURL(key string) string
}

type RoomBillService struct {
	roomBillRepository repository.RoomBillRepository
	objectStorage      RoomBillObjectStorage
}

func NewRoomBillService(roomBillRepo repository.RoomBillRepository, objectStorage RoomBillObjectStorage) *RoomBillService {
	return &RoomBillService{roomBillRepository: roomBillRepo, objectStorage: objectStorage}
}

func (s *RoomBillService) CreateRoomBill(req *dto.RoomBillRequest, receiptImage []byte, contentType, extension string) (*model.RoomBill, error) {
	roomBill := &model.RoomBill{
		ID: uuid.New(), RoomId: req.RoomID, CategoryId: req.CategoryID,
		Name: req.Name, Description: req.Description, Total: req.Total,
		Status: "pending", DueDate: req.DueDate, CreatedBy: req.CreatedBy,
	}

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
		roomBill.ReceiptImageKey = &key
	}

	createdRoomBill, err := s.roomBillRepository.Create(roomBill)
	if err != nil {
		if roomBill.ReceiptImageKey != nil && s.objectStorage != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			_ = s.objectStorage.Delete(ctx, *roomBill.ReceiptImageKey)
		}
		return nil, err
	}
	s.addReceiptImageURL(createdRoomBill)
	return createdRoomBill, nil
}

func (s *RoomBillService) GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error) {
	bills, err := s.roomBillRepository.GetRoomBills(roomID, userID)
	if err != nil {
		return nil, err
	}
	for _, bill := range bills {
		s.addReceiptImageURL(bill)
	}
	return bills, nil
}

func (s *RoomBillService) GetRoomBillByID(roomID, billID, userID uuid.UUID) (*model.RoomBill, error) {
	bill, err := s.roomBillRepository.GetRoomBillById(roomID, userID, billID)
	if err != nil {
		return nil, err
	}
	s.addReceiptImageURL(bill)
	return bill, nil
}

func (s *RoomBillService) addReceiptImageURL(bill *model.RoomBill) {
	if bill.ReceiptImageKey == nil || *bill.ReceiptImageKey == "" || s.objectStorage == nil {
		return
	}
	bill.ReceiptImageURL = s.objectStorage.PublicURL(*bill.ReceiptImageKey)
}

func (s *RoomBillService) CreateRoomBillShares(roomID, billID, requesterID uuid.UUID, req *dto.CreateRoomBillSharesRequest) ([]*model.RoomBillShare, error) {
	shares := make([]model.RoomBillShare, 0, len(req.Members))
	for _, member := range req.Members {
		shares = append(shares, model.RoomBillShare{UserID: member.UserID, Amount: member.Amount})
	}
	return s.roomBillRepository.CreateShares(roomID, billID, requesterID, shares)
}
