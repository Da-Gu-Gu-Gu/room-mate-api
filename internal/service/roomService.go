package service

import (
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

type RoomMemberService struct {
	roomMemberRepository repository.RoomMemberRepository
}

func NewRoomMemberService(roomMemberRepository repository.RoomMemberRepository) *RoomMemberService {
	return &RoomMemberService{
		roomMemberRepository: roomMemberRepository,
	}
}

func (r *RoomMemberService) GetRoomDetail(id uuid.UUID) ([]*model.RoomMember, error) {
	roomMembers, err := r.roomMemberRepository.GetRoomMembers(id)
	if err != nil {
		return nil, err
	}
	return roomMembers, nil
}
