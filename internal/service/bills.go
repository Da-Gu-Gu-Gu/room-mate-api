package service

import (
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
)

type BillService struct {
	billRepository repository.BillRepository
}

func NewBillService(billRepo repository.BillRepository) *BillService {
	return &BillService{billRepository: billRepo}
}

func (s *BillService) GetBillCategories() ([]model.BillCategory, error) {
	categories, err := s.billRepository.GetBillCategories()
	if err != nil {
		return nil, err
	}
	return categories, nil
}
