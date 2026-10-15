package service

import (
	"fmt"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
)

type ThemeService struct {
	themeRepository repository.ThemeRepository
}

func NewThemeService(themeRepository repository.ThemeRepository) *ThemeService {
	return &ThemeService{
		themeRepository: themeRepository,
	}
}

func (s *ThemeService) GetByThemeID(themeId uuid.UUID) (*model.Theme, error) {
	theme, err := s.themeRepository.FindById(themeId)
	if err != nil {
		return nil, fmt.Errorf("failed to get theme by id: %w", err)
	}
	return theme, nil

}

func (s *ThemeService) GetThemes() ([]*model.Theme, error) {
	themesList, err := s.themeRepository.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get themes: %w", err)
	}
	return themesList, nil
}

func (s *ThemeService) CreateTheme(req *dto.ThemeRequest) (*model.Theme, error) {
	theme := &model.Theme{
		PrimaryColor:    req.PrimaryColor,
		SecondaryColor:  req.SecondaryColor,
		BackgroundColor: req.BackgroundColor,
	}
	if req.BackgroundImgUrl != nil {
		theme.BackgroundImgUrl = req.BackgroundImgUrl
	}
	if err := s.themeRepository.Create(theme); err != nil {
		return nil, fmt.Errorf("failed to create theme: %w", err)
	}
	return theme, nil
}

func (s *ThemeService) DeleteTheme(themeId uuid.UUID) error {
	err := s.themeRepository.Delete(themeId)
	if err != nil {
		return fmt.Errorf("failed to delete theme: %w", err)
	}
	return nil
}

func (s *ThemeService) UpdateTheme(themeId uuid.UUID, req *dto.UpdateThemeRequest) (*model.Theme, error) {

	theme, err := s.themeRepository.FindById(themeId)

	if err != nil {
		return nil, err
	}
	if req.PrimaryColor != nil {
		theme.PrimaryColor = *req.PrimaryColor
	}
	if req.SecondaryColor != nil {
		theme.SecondaryColor = *req.SecondaryColor
	}
	if req.BackgroundColor != nil {
		theme.BackgroundColor = *req.BackgroundColor
	}
	if req.BackgroundImgUrl != nil {
		theme.BackgroundImgUrl = *&req.BackgroundImgUrl
	}

	err = s.themeRepository.Update(theme)
	if err != nil {
		return nil, err
	}
	return theme, nil
}
