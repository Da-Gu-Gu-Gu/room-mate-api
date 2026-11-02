package handler

import (
	"bytes"
	"errors"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"room-mate-api/internal/dto"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const maxRoomBillReceiptSize = 5 << 20 // 5 MB

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
	var receiptImage []byte
	var contentType, extension string

	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(1 << 20); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart payload"})
			return
		}

		categoryID, err := uuid.Parse(strings.TrimSpace(c.PostForm("categoryId")))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid categoryId"})
			return
		}
		total, err := strconv.ParseFloat(strings.TrimSpace(c.PostForm("total")), 64)
		if err != nil || total <= 0 || math.IsInf(total, 0) || math.IsNaN(total) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "total must be greater than zero"})
			return
		}
		name := strings.TrimSpace(c.PostForm("name"))
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		req = dto.RoomBillRequest{
			RoomID:     roomID,
			CategoryID: categoryID,
			Name:       name,
			Total:      total,
			CreatedBy:  userID,
		}
		if description := strings.TrimSpace(c.PostForm("description")); description != "" {
			req.Description = &description
		}
		if dueDateValue := strings.TrimSpace(c.PostForm("dueDate")); dueDateValue != "" {
			dueDate, parseErr := time.Parse(time.RFC3339, dueDateValue)
			if parseErr != nil {
				dueDate, parseErr = time.Parse("2006-01-02", dueDateValue)
			}
			if parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "dueDate must use YYYY-MM-DD or RFC3339 format"})
				return
			}
			req.DueDate = &dueDate
		}

		fileHeader, fileErr := c.FormFile("receiptImage")
		if fileErr != nil && !errors.Is(fileErr, http.ErrMissingFile) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receiptImage"})
			return
		}
		if fileHeader != nil {
			if fileHeader.Size > maxRoomBillReceiptSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "receiptImage must be 5 MB or smaller"})
				return
			}
			file, openErr := fileHeader.Open()
			if openErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read receiptImage"})
				return
			}
			defer file.Close()

			receiptImage, err = io.ReadAll(io.LimitReader(file, maxRoomBillReceiptSize+1))
			if err != nil || len(receiptImage) > maxRoomBillReceiptSize {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "receiptImage must be 5 MB or smaller"})
				return
			}
			contentType, extension, err = detectReceiptImage(receiptImage)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
			return
		}
		req.RoomID = roomID
		req.CreatedBy = userID
	}

	createdRoomBill, err := h.roomBillService.CreateRoomBill(&req, receiptImage, contentType, extension)
	if err != nil {
		if errors.Is(err, service.ErrRoomBillStorageUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image storage is not configured"})
			return
		}
		respondRoomError(c, err)
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

type RoomBillHandler struct {
	roomBillService *service.RoomBillService
}

func NewRoomBillHandler(roomBillService *service.RoomBillService) *RoomBillHandler {
	return &RoomBillHandler{
		roomBillService: roomBillService,
	}
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
