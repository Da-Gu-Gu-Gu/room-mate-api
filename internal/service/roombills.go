package service

import (
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

type RoomBillService struct {
	roomBillRepository repository.RoomBillRepository
}

func NewRoomBillService(roomBillRepo repository.RoomBillRepository) *RoomBillService {
	return &RoomBillService{roomBillRepository: roomBillRepo}
}

func (s *RoomBillService) CreateRoomBill(req *dto.RoomBillRequest) (*model.RoomBill, error) {
	roomBill := &model.RoomBill{
		RoomId:      req.RoomID,
		CategoryId:  req.CategoryID,
		Name:        req.Name,
		Description: req.Description,
		Total:       req.Total,
		Status:      "pending", // Default status
		DueDate:     req.DueDate,
		CreatedBy:   req.CreatedBy,
	}
	createdRoomBill, err := s.roomBillRepository.Create(roomBill)
	if err != nil {
		return nil, err
	}
	return createdRoomBill, nil
}

func (s *RoomBillService) GetRoomBills(roomID, userID uuid.UUID) ([]*model.RoomBill, error) {
	bills, err := s.roomBillRepository.GetRoomBills(roomID, userID)
	if err != nil {
		return nil, err
	}
	return bills, nil
}

func (s *RoomBillService) GetRoomBillByID(roomID, billID, userID uuid.UUID) (*model.RoomBill, error) {
	bills, err := s.roomBillRepository.GetRoomBillById(roomID, userID, billID)
	if err != nil {
		return nil, err
	}
	return bills, nil
}
