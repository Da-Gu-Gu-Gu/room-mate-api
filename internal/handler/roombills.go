package handler

import (
	"errors"
	"net/http"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomBillHandler struct {
	roomBillService *service.RoomBillService
}

func NewRoomBillHandler(roomBillService *service.RoomBillService) *RoomBillHandler {
	return &RoomBillHandler{roomBillService: roomBillService}
}

func (h *RoomBillHandler) CreateRoomBill(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var req dto.RoomBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}
	req.RoomID = roomID
	req.CreatedBy = userID

	createdRoomBill, err := h.roomBillService.CreateRoomBill(&req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrRoomBillStorageUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image storage is not configured"})
		case errors.Is(err, service.ErrInvalidReceiptImageURL):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			respondRoomError(c, err)
		}
		return
	}
	c.JSON(http.StatusCreated, createdRoomBill)
}

func (h *RoomBillHandler) GetRoomBills(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	bills, err := h.roomBillService.GetRoomBills(roomID, userID)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, bills)
}

func (h *RoomBillHandler) GetRoomBillByID(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	billID, err := uuid.Parse(c.Param("billId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bill ID"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	bill, err := h.roomBillService.GetRoomBillByID(roomID, billID, userID)
	if err != nil {
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusOK, bill)
}

func (h *RoomBillHandler) CreateRoomBillShares(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	billID, err := uuid.Parse(c.Param("billId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bill ID"})
		return
	}
	requesterID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var req dto.CreateRoomBillSharesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	shares, err := h.roomBillService.CreateRoomBillShares(roomID, billID, requesterID, &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRoomBillNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "room bill not found"})
		case errors.Is(err, repository.ErrRoomBillShareForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "you or one of the selected users is not an active member of this room"})
		case errors.Is(err, repository.ErrRoomBillShareTotal), errors.Is(err, repository.ErrInvalidRoomBillShares):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, repository.ErrRoomBillSharesAlreadyAssigned):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusCreated, gin.H{"members": shares})
}

func (h *RoomBillHandler) GetRoomBillShares(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	billID, err := uuid.Parse(c.Param("billId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bill ID"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	shares, err := h.roomBillService.GetRoomBillShares(roomID, billID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRoomBillNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "room bill not found"})
		case errors.Is(err, repository.ErrRoomBillShareForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "you are not an active member of this room"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"members": shares})
}

func (h *RoomBillHandler) CreateRoomBillPayment(c *gin.Context) {
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return
	}
	billID, err := uuid.Parse(c.Param("billId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bill ID"})
		return
	}
	userID, ok := currentUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	var req dto.CreateRoomBillPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	payment, err := h.roomBillService.CreateRoomBillPayment(roomID, billID, userID, &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrRoomBillNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "room bill not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusCreated, payment)
}
