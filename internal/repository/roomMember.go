package repository

import (
	"context"

	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoomMemberRepository interface {
	GetRoomMembers(id uuid.UUID) ([]*model.RoomMember, error)
	GetMyInvitations(userID uuid.UUID) ([]*model.Invitation, error)
	IsActiveMember(roomID, userID uuid.UUID) (bool, error)
	Invite(ctx context.Context, roomID, inviterID, inviteeID uuid.UUID) (*model.Invitation, error)
	RespondInvitation(ctx context.Context, invitationID, userID uuid.UUID, accept bool) error
	LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error
	TransferOwnership(ctx context.Context, roomID, ownerID, newOwnerID uuid.UUID) error
}

type roomMemberRepository struct {
	db *gorm.DB
}

func NewRoomMemberRepository(db *gorm.DB) *roomMemberRepository {
	return &roomMemberRepository{db: db}
}

func (r *roomMemberRepository) GetRoomMembers(id uuid.UUID) ([]*model.RoomMember, error) {
	var members []*model.RoomMember
	err := r.db.Preload("User").
		Where("room_id = ? AND leave_at IS NULL", id).
		Order("joined_at ASC").
		Find(&members).Error
	if err != nil {
		return nil, err
	}
	return members, nil
}

func (r *roomMemberRepository) GetMyInvitations(userID uuid.UUID) ([]*model.Invitation, error) {
	var invitations []*model.Invitation
	err := r.db.Preload("Room").
		Where("invitee_id = ? AND status = ?", userID, model.InvitationPending).
		Order("created_at DESC").
		Find(&invitations).Error
	if err != nil {
		return nil, err
	}
	return invitations, nil
}
