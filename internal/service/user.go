package service

import (
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
)

type UserService struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (s *UserService) GetUsers() ([]*model.User, error) {
	userList, err := s.userRepository.GET()
	if err != nil {
		return nil, err
	}
	return userList, nil
}
