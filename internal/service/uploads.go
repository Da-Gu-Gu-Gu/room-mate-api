package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
)

type ImageObjectStorage interface {
	Upload(ctx context.Context, key string, body io.Reader, contentType string) error
	PublicURL(key string) string
}

type UploadService struct {
	storage ImageObjectStorage
}

func NewUploadService(storage ImageObjectStorage) *UploadService {
	return &UploadService{storage: storage}
}

// UploadImage stores an image under a feature-independent prefix and returns
// its permanent public URL.
func (s *UploadService) UploadImage(image []byte, contentType, extension string) (string, error) {
	if s.storage == nil {
		return "", ErrRoomBillStorageUnavailable
	}

	key := fmt.Sprintf("uploads/images/%s.%s", uuid.New(), extension)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.storage.Upload(ctx, key, bytes.NewReader(image), contentType); err != nil {
		return "", err
	}
	return s.storage.PublicURL(key), nil
}
