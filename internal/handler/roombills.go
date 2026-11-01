package handler

import (
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomBillHandler struct {
	roomBillService *service.RoomBillService
}

func NewRoomBillHandler(roomBillService *service.RoomBillService) *RoomBillHandler {
	return &RoomBillHandler{
		roomBillService: roomBillService,
	}
}

func (h *RoomBillHandler) CreateRoomBill(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("id"))
	id, ok := currentUserID(c)
	if !ok || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user or room ID"})
		return
	}
	var req dto.RoomBillRequest
	req.CreatedBy = id
	req.RoomID = roomId
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	createdRoomBill, err := h.roomBillService.CreateRoomBill(&req)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, createdRoomBill)
}

func (h *RoomBillHandler) GetRoomBills(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	userId, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	bills, err := h.roomBillService.GetRoomBills(roomId, userId)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, bills)
}

func (h *RoomBillHandler) GetRoomBillByID(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	billId, err := uuid.Parse(c.Param("billId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bill ID"})
		return
	}
	userId, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	bill, err := h.roomBillService.GetRoomBillByID(roomId, billId, userId)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, bill)

}
