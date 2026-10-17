package service

import (
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepository repository.RoomRepository
}

func NewRoomService(roomRepo repository.RoomRepository) *RoomService {
	return &RoomService{
		roomRepository: roomRepo,
	}
}

func (s *RoomService) CreateRoom(req *dto.RoomRequest) (*model.Room, error) {

	roomModel := &model.Room{
		RoomName:   req.RoomName,
		RoomNumber: req.RoomNumber,
		CreatedBy:  req.CreatedBy,
		RoomCode:   "123123", //generated random number 6
	}
	err := s.roomRepository.CREATE(roomModel)
	if err != nil {
		return nil, err
	}
	return roomModel, nil
}

func (s *RoomService) UpdateRoom(id uuid.UUID, req *dto.UpdateRoomRequest) (*model.Room, error) {
	room, err := s.roomRepository.GETBYID(id)
	if err != nil {
		return nil, err
	}

	if req.RoomName != nil {
		room.RoomName = *req.RoomName
	}
	if req.RoomNumber != nil {
		room.RoomName = *req.RoomNumber
	}
	if err := s.roomRepository.UPDATE(room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetRoom() ([]*model.Room, error) {
	rooms, err := s.roomRepository.GET()
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (s *RoomService) GetRoomById(id uuid.UUID) (*model.Room, error) {
	room, err := s.roomRepository.GETBYID(id)
	if err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) DeleteRoom(id uuid.UUID) error {
	err := s.roomRepository.DELETE(id)
	if err != nil {
		return err
	}
	return nil
}
