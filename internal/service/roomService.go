package service

import (
	"context"
	"errors"

	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRoomMemberForbidden = repository.ErrRoomMemberForbidden
	ErrAlreadyInRoom = repository.ErrAlreadyInRoom
	ErrPendingInvitationExists = repository.ErrPendingInvitationExists
	ErrInvitationNotFound = repository.ErrInvitationNotFound
	ErrInvitationNotPending = repository.ErrInvitationNotPending
	ErrOwnerCannotLeave = repository.ErrOwnerCannotLeave
	ErrCannotTransferToSelf = repository.ErrCannotTransferToSelf
	ErrTargetNotActiveMember = repository.ErrTargetNotActiveMember
	ErrNotActiveRoomMember = repository.ErrNotActiveRoomMember
)

type RoomMemberService struct {
	roomMemberRepository repository.RoomMemberRepository
}

func NewRoomMemberService(roomMemberRepository repository.RoomMemberRepository) *RoomMemberService {
	return &RoomMemberService{roomMemberRepository: roomMemberRepository}
}

func (s *RoomMemberService) GetRoomDetail(actorID, roomID uuid.UUID) ([]*model.RoomMember, error) {
	isMember, err := s.roomMemberRepository.IsActiveMember(roomID, actorID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrRoomMemberForbidden
	}
	return s.roomMemberRepository.GetRoomMembers(roomID)
}

func (s *RoomMemberService) GetMyInvitations(userID uuid.UUID) ([]*model.Invitation, error) {
	return s.roomMemberRepository.GetMyInvitations(userID)
}

func (s *RoomMemberService) Invite(ctx context.Context, roomID, inviterID, inviteeID uuid.UUID) (*model.Invitation, error) {
	invitation, err := s.roomMemberRepository.Invite(ctx, roomID, inviterID, inviteeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (s *RoomMemberService) RespondInvitation(ctx context.Context, invitationID, userID uuid.UUID, accept bool) error {
	return s.roomMemberRepository.RespondInvitation(ctx, invitationID, userID, accept)
}

func (s *RoomMemberService) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	return s.roomMemberRepository.LeaveRoom(ctx, roomID, userID)
}

func (s *RoomMemberService) TransferOwnership(ctx context.Context, roomID, ownerID, newOwnerID uuid.UUID) error {
	return s.roomMemberRepository.TransferOwnership(ctx, roomID, ownerID, newOwnerID)
}
