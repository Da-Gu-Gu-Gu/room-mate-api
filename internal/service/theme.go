package service

import (
	"fmt"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
)

type ThemeService struct {
	themeRepository repository.ThemeRepository
}

func NewThemeService(themeRepository repository.ThemeRepository) *ThemeService {
	return &ThemeService{
		themeRepository: themeRepository,
	}
}

func (s *ThemeService) GetByThemeID(themeId string) (*model.Theme, error) {
	theme, err := s.themeRepository.FindById(themeId)
	if err != nil {
		return nil, fmt.Errorf("failed to get theme by id: %w", err)
	}
	return theme, nil

}
