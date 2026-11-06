package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxRoomBillReceiptSize = 5 << 20 // 5 MB

type RoomBillHandler struct {
	roomBillService *service.RoomBillService
}

func NewRoomBillHandler(roomBillService *service.RoomBillService) *RoomBillHandler {
	return &RoomBillHandler{roomBillService: roomBillService}
}

// UploadRoomBillReceipt uploads a receipt independently. The returned URL can
// then be included as receiptImageUrl when creating the room bill.
func (h *RoomBillHandler) UploadRoomBillReceipt(c *gin.Context) {
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

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRoomBillReceiptSize+(1<<20))
	fileHeader, err := c.FormFile("receiptImage")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "receiptImage file is required"})
		return
	}
	if fileHeader.Size > maxRoomBillReceiptSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "receiptImage must be 5 MB or smaller"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read receiptImage"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(io.LimitReader(file, maxRoomBillReceiptSize+1))
	if err != nil || len(imageBytes) > maxRoomBillReceiptSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "receiptImage must be 5 MB or smaller"})
		return
	}
	contentType, extension, err := detectReceiptImage(imageBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageURL, err := h.roomBillService.UploadRoomBillReceipt(roomID, userID, imageBytes, contentType, extension)
	if err != nil {
		if errors.Is(err, service.ErrRoomBillStorageUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image storage is not configured"})
			return
		}
		respondRoomError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"receiptImageUrl": imageURL})
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

func detectReceiptImage(data []byte) (contentType, extension string, err error) {
	switch {
	case len(data) >= 3 && bytes.Equal(data[:3], []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg", "jpg", nil
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}):
		return "image/png", "png", nil
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp", "webp", nil
	default:
		return "", "", errors.New("receiptImage must be a valid JPEG, PNG, or WebP image")
	}
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
