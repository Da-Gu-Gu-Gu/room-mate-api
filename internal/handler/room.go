package handler

import (
	"fmt"
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomHandler struct {
	roomService *service.RoomService
}

func NewRoomHandler(roomService *service.RoomService) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var payload dto.RoomRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		fmt.Print(&payload)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})

		return
	}
	createdRoom, err := h.roomService.CreateRoom(&payload)
	if err != nil {
		c.JSON(500, "Something Wrong!")
		return
	}
	c.JSON(200, createdRoom)
}

func (h *RoomHandler) GetRooms(c *gin.Context) {
	rooms, err := h.roomService.GetRoom()
	if err != nil {
		c.JSON(500, "Something Wrong!")
		return
	}
	c.JSON(200, rooms)
}

func (h *RoomHandler) GetRoomById(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, "Something Wrong!")
		return
	}
	room, err := h.roomService.GetRoomById(roomID)
	if err != nil {
		c.JSON(500, "Something Wrong!")
		return
	}
	c.JSON(200, room)
}

func (h *RoomHandler) UpdateRoom(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, "Something Wrong!")
		return
	}
	var payload dto.UpdateRoomRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, "Something Wrong!")
		return
	}
	roomUpdated, err := h.roomService.UpdateRoom(roomID, &payload)
	if err != nil {
		c.JSON(500, "Something Wrong!")
	}
	c.JSON(200, roomUpdated)
}

func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, "Something Wrong!")
		return
	}
	if err := h.roomService.DeleteRoom(roomID); err != nil {
		c.JSON(500, "Something Wrong!")
	}

	c.JSON(200, "Deleted successfully.")
}
