package service

import (
	"context"
	"errors"
	"io"
	"strings"

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
		if err != nil || !strings.HasPrefix(key, "uploads/images/") {
			return nil, ErrInvalidReceiptImageURL
		}

		// Persist the complete public URL instead of the S3 object key.
		imageURL := strings.TrimSpace(*req.ReceiptImageURL)
		roomBill.ReceiptImageURL = &imageURL
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

// addReceiptImageURL keeps responses compatible with records created before
// receipt_image_url was introduced.
func (s *RoomBillService) addReceiptImageURL(bill *model.RoomBill) {
	if bill.ReceiptImageURL != nil && *bill.ReceiptImageURL != "" {
		return
	}
	if bill.ReceiptImageKey == nil || *bill.ReceiptImageKey == "" || s.objectStorage == nil {
		return
	}
	imageURL := s.objectStorage.PublicURL(*bill.ReceiptImageKey)
	bill.ReceiptImageURL = &imageURL
}

func (s *RoomBillService) CreateRoomBillShares(roomID, billID, requesterID uuid.UUID, req *dto.CreateRoomBillSharesRequest) ([]*model.RoomBillShare, error) {
	shares := make([]model.RoomBillShare, 0, len(req.Members))
	for _, member := range req.Members {
		shares = append(shares, model.RoomBillShare{UserID: member.UserID, Amount: member.Amount})
	}
	return s.roomBillRepository.CreateShares(roomID, billID, requesterID, shares)
}

func (s *RoomBillService) GetRoomBillShares(roomID, billID, userID uuid.UUID) ([]*model.RoomBillShare, error) {
	return s.roomBillRepository.GETShares(roomID, billID, userID)
}

func (s *RoomBillService) CreateRoomBillPayment(roomID, billShareID, userID uuid.UUID, req *dto.CreateRoomBillPaymentRequest) (*model.RoomBillPayment, error) {
	var payment = &model.RoomBillPayment{
		BillShareID: req.BillShareID, UserID: req.UserID, Status: "Pending", Note: req.Note, ProofURL: req.ProofURL, Amount: req.Amount,
	}
	return s.roomBillRepository.CreatePayment(roomID, billShareID, userID, payment)
}
