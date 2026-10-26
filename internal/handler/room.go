package handler

import (
	"errors"
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/middleware"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return uuid.Nil, false
	}
	return id, true
}

func respondRoomError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "you do not have permission to access this room"})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var payload dto.RoomRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// The creator is derived from the verified token, never from client input.
	payload.CreatedBy = userID

	createdRoom, err := h.roomService.CreateRoom(c.Request.Context(), &payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create room"})
		return
	}
	c.JSON(http.StatusCreated, createdRoom)
}

func (h *RoomHandler) GetRoomById(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	room, err := h.roomService.GetRoomById(userID, roomID)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	var payload dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedRoom, err := h.roomService.UpdateRoom(userID, roomID, &payload)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, updatedRoom)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}
	if err := h.roomService.DeleteRoom(userID, roomID); err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "room deleted successfully"})
}
