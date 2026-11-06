package handler

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
)

const maxImageUploadSize = 5 << 20 // 5 MB

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{uploadService: uploadService}
}

// UploadImage is a generic authenticated image upload endpoint.
// Multipart form field: image.
func (h *UploadHandler) UploadImage(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxImageUploadSize+(1<<20))
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}
	if fileHeader.Size > maxImageUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image must be 5 MB or smaller"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to read image"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(io.LimitReader(file, maxImageUploadSize+1))
	if err != nil || len(imageBytes) > maxImageUploadSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "image must be 5 MB or smaller"})
		return
	}

	contentType, extension, err := detectUploadedImage(imageBytes)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageURL, err := h.uploadService.UploadImage(imageBytes, contentType, extension)
	if err != nil {
		if errors.Is(err, service.ErrRoomBillStorageUnavailable) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "image storage is not configured"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to upload image"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"imageUrl": imageURL})
}

func detectUploadedImage(data []byte) (contentType, extension string, err error) {
	switch {
	case len(data) >= 3 && bytes.Equal(data[:3], []byte{0xff, 0xd8, 0xff}):
		return "image/jpeg", "jpg", nil
	case len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a}):
		return "image/png", "png", nil
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return "image/webp", "webp", nil
	default:
		return "", "", errors.New("image must be a valid JPEG, PNG, or WebP image")
	}
}
