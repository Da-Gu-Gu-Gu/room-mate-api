package repository

import (
	"context"
	"errors"
	"time"

	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrRoomMemberForbidden = errors.New("room member action forbidden")
	ErrAlreadyInRoom = errors.New("user already belongs to an active room")
	ErrPendingInvitationExists = errors.New("a pending invitation already exists")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrInvitationNotPending = errors.New("invitation is no longer pending")
	ErrOwnerCannotLeave = errors.New("room owner cannot leave; delete the room or transfer ownership first")
	ErrNotActiveRoomMember = errors.New("user is not an active member of this room")
)

func (r *roomMemberRepository) IsActiveMember(roomID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Model(&model.RoomMember{}).
		Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *roomMemberRepository) Invite(ctx context.Context, roomID, inviterID, inviteeID uuid.UUID) (*model.Invitation, error) {
	var invitation model.Invitation
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inviter model.RoomMember
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, inviterID).
			First(&inviter).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrRoomMemberForbidden
			}
			return err
		}
		if inviter.Role != model.RoleAdmin {
			return ErrRoomMemberForbidden
		}

		var invitee model.User
		if err := tx.First(&invitee, "id = ?", inviteeID).Error; err != nil {
			return err
		}
		if inviteeID == inviterID {
			return ErrRoomMemberForbidden
		}

		var activeCount int64
		if err := tx.Model(&model.RoomMember{}).
			Where("user_id = ? AND leave_at IS NULL", inviteeID).
			Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrAlreadyInRoom
		}

		var pendingCount int64
		if err := tx.Model(&model.Invitation{}).
			Where("room_id = ? AND invitee_id = ? AND status = ?", roomID, inviteeID, model.InvitationPending).
			Count(&pendingCount).Error; err != nil {
			return err
		}
		if pendingCount > 0 {
			return ErrPendingInvitationExists
		}

		invitation = model.Invitation{
			RoomID: roomID, InviterID: inviterID, InviteeID: inviteeID,
			Status: model.InvitationPending,
		}
		return tx.Create(&invitation).Error
	})
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *roomMemberRepository) RespondInvitation(ctx context.Context, invitationID, userID uuid.UUID, accept bool) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invitation model.Invitation
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&invitation, "id = ?", invitationID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInvitationNotFound
		}
		if err != nil {
			return err
		}
		if invitation.InviteeID != userID {
			return ErrRoomMemberForbidden
		}
		if invitation.Status != model.InvitationPending {
			return ErrInvitationNotPending
		}

		if accept {
			var activeCount int64
			if err := tx.Model(&model.RoomMember{}).
				Where("user_id = ? AND leave_at IS NULL", userID).
				Count(&activeCount).Error; err != nil {
				return err
			}
			if activeCount > 0 {
				return ErrAlreadyInRoom
			}

			joinedAt := time.Now().UTC()
			member := model.RoomMember{
				RoomID: invitation.RoomID,
				UserID: userID,
				Role: model.RoleMember,
				JoinedAt: &joinedAt,
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
			invitation.Status = model.InvitationAccepted
		} else {
			invitation.Status = model.InvitationDeclined
		}
		return tx.Model(&invitation).Updates(map[string]interface{}{
			"status": invitation.Status,
			"updated_at": time.Now().UTC(),
		}).Error
	})
}

func (r *roomMemberRepository) LeaveRoom(ctx context.Context, roomID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var member model.RoomMember
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("room_id = ? AND user_id = ? AND leave_at IS NULL", roomID, userID).
			First(&member).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotActiveRoomMember
		}
		if err != nil {
			return err
		}
		if member.Role == model.RoleAdmin {
			return ErrOwnerCannotLeave
		}
		now := time.Now().UTC()
		return tx.Model(&member).Updates(map[string]interface{}{
			"leave_at": now,
			"updated_at": now,
		}).Error
	})
}
