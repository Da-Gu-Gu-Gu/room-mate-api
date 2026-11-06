package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrRoomBillStorageUnavailable = errors.New("room bill image storage is not configured")
	ErrInvalidReceiptImageURL     = errors.New("receipt image URL is invalid for this room")
)

type RoomBillObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
	ObjectKeyFromPublicURL(publicURL string) (string, error)
}

type RoomBillService struct {
	roomBillRepository repository.RoomBillRepository
	objectStorage      RoomBillObjectStorage
}

func NewRoomBillService(roomBillRepo repository.RoomBillRepository, objectStorage RoomBillObjectStorage) *RoomBillService {
	return &RoomBillService{roomBillRepository: roomBillRepo, objectStorage: objectStorage}
}

// UploadRoomBillReceipt uploads a receipt independently from bill creation.
// The returned URL is stable and can be included in CreateRoomBillRequest.
func (s *RoomBillService) UploadRoomBillReceipt(roomID, userID uuid.UUID, image []byte, contentType, extension string) (string, error) {
	if s.objectStorage == nil {
		return "", ErrRoomBillStorageUnavailable
	}
	if err := s.roomBillRepository.CanCreate(roomID, userID); err != nil {
		return "", err
	}

	key := fmt.Sprintf("room-bills/%s/uploads/%s.%s", roomID, uuid.New(), extension)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.objectStorage.Upload(ctx, key, bytes.NewReader(image), contentType); err != nil {
		return "", err
	}
	return s.objectStorage.PublicURL(key), nil
}

func (s *RoomBillService) CreateRoomBill(req *dto.RoomBillRequest) (*model.RoomBill, error) {
	roomBill := &model.RoomBill{
		ID: uuid.New(), RoomId: req.RoomID, CategoryId: req.CategoryID,
		Name: req.Name, Description: req.Description, Total: req.Total,
		Status: "pending", DueDate: req.DueDate, CreatedBy: req.CreatedBy,
	}

	if err := s.roomBillRepository.CanCreate(req.RoomID, req.CreatedBy); err != nil {
		return nil, err
	}

	if req.ReceiptImageURL != nil && strings.TrimSpace(*req.ReceiptImageURL) != "" {
		if s.objectStorage == nil {
			return nil, ErrRoomBillStorageUnavailable
		}
		key, err := s.objectStorage.ObjectKeyFromPublicURL(*req.ReceiptImageURL)
		if err != nil || !strings.HasPrefix(key, fmt.Sprintf("room-bills/%s/uploads/", req.RoomID)) {
			return nil, ErrInvalidReceiptImageURL
		}
		roomBill.ReceiptImageKey = &key
	}

	createdRoomBill, err := s.roomBillRepository.Create(roomBill)
	if err != nil {
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
