package service

import (
	"context"
	"errors"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound = errors.New("resource not found")
)

type RoomService struct {
	roomRepository repository.RoomRepository
}

func NewRoomService(roomRepo repository.RoomRepository) *RoomService {
	return &RoomService{roomRepository: roomRepo}
}

func (s *RoomService) CreateRoom(ctx context.Context, req *dto.RoomRequest) (*model.Room, error) {
	room := &model.Room{
		RoomName: req.RoomName,
		RoomNumber: req.RoomNumber,
		CreatedBy: req.CreatedBy,
		RoomCode: uuid.NewString()[:8],
	}
	if err := s.roomRepository.Create(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) UpdateRoom(actorID, roomID uuid.UUID, req *dto.UpdateRoomRequest) (*model.Room, error) {
	room, err := s.roomRepository.GETBYID(roomID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if room.CreatedBy != actorID {
		return nil, ErrForbidden
	}

	if req.RoomName != nil {
		room.RoomName = *req.RoomName
	}
	if req.RoomNumber != nil {
		room.RoomNumber = *req.RoomNumber
	}
	if err := s.roomRepository.UPDATE(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetRoom() ([]*model.Room, error) {
	return s.roomRepository.GET()
}

func (s *RoomService) GetRoomById(actorID, roomID uuid.UUID) (*model.Room, error) {
	room, err := s.roomRepository.GETBYID(roomID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	isMember, err := s.roomRepository.IsActiveMember(roomID, actorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}
	return room, nil
}

func (s *RoomService) DeleteRoom(actorID, roomID uuid.UUID) error {
	room, err := s.roomRepository.GETBYID(roomID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if room.CreatedBy != actorID {
		return ErrForbidden
	}
	return s.roomRepository.DELETE(roomID)
}
