package handler

import (
	"errors"
	"net/http"

	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomMember struct {
	roomMemberService *service.RoomMemberService
}

func NewRoomMemberHandler(roomMemberService *service.RoomMemberService) *RoomMember {
	return &RoomMember{roomMemberService: roomMemberService}
}

func (h *RoomMember) GetMyInvitations(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	invitations, err := h.roomMemberService.GetMyInvitations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load invitations"})
		return
	}
	c.JSON(http.StatusOK, invitations)
}

func (h *RoomMember) GetRoomDetail(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	members, err := h.roomMemberService.GetRoomDetail(userID, roomID)
	if err != nil {
		respondRoomMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *RoomMember) Invite(c *gin.Context) {
	inviterID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	var req struct {
		InviteeID uuid.UUID `json:"inviteeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "inviteeId must be a valid UUID"})
		return
	}
	invitation, err := h.roomMemberService.Invite(c.Request.Context(), roomID, inviterID, req.InviteeID)
	if err != nil {
		respondRoomMemberError(c, err)
		return
	}
	c.JSON(http.StatusCreated, invitation)
}

func (h *RoomMember) AcceptInvitation(c *gin.Context) {
	h.respondToInvitation(c, true)
}

func (h *RoomMember) DeclineInvitation(c *gin.Context) {
	h.respondToInvitation(c, false)
}

func (h *RoomMember) respondToInvitation(c *gin.Context, accept bool) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	invitationID, err := uuid.Parse(c.Param("invitationId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}
	if err := h.roomMemberService.RespondInvitation(c.Request.Context(), invitationID, userID, accept); err != nil {
		respondRoomMemberError(c, err)
		return
	}
	message := "invitation declined"
	if accept {
		message = "invitation accepted; you joined the room"
	}
	c.JSON(http.StatusOK, gin.H{"message": message})
}

func (h *RoomMember) TransferOwnership(c *gin.Context) {
	ownerID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	var req struct {
		NewOwnerID uuid.UUID `json:"newOwnerId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "newOwnerId must be a valid UUID"})
		return
	}
	if err := h.roomMemberService.TransferOwnership(c.Request.Context(), roomID, ownerID, req.NewOwnerID); err != nil {
		respondRoomMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "room ownership transferred successfully"})
}

func (h *RoomMember) LeaveRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	if err := h.roomMemberService.LeaveRoom(c.Request.Context(), roomID, userID); err != nil {
		respondRoomMemberError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "you have left the room"})
}

func respondRoomMemberError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrRoomMemberForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to perform this action"})
	case errors.Is(err, service.ErrAlreadyInRoom), errors.Is(err, service.ErrPendingInvitationExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvitationNotFound), errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "room, user, or invitation not found"})
	case errors.Is(err, service.ErrInvitationNotPending), errors.Is(err, service.ErrOwnerCannotLeave), errors.Is(err, service.ErrCannotTransferToSelf), errors.Is(err, service.ErrTargetNotActiveMember):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNotActiveRoomMember):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
