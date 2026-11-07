package handler

import (
	"errors"
	"io"
	"net/http"

	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gabriel-vasile/mimetype"
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

	detected := mimetype.Detect(imageBytes)
	allowedTypes := map[string]string{
		"image/jpeg": "jpg",
		"image/png":  "png",
		"image/webp": "webp",
	}
	extension, ok := allowedTypes[detected.String()]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only JPEG, PNG, and WebP images are allowed"})
		return
	}

	imageURL, err := h.uploadService.UploadImage(imageBytes, detected.String(), extension)
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
