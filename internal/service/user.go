package service

import (
	"fmt"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"

	"github.com/google/uuid"
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

func (s *UserService) GetUserById(id uuid.UUID) (*model.User, error) {
	user, err := s.userRepository.GetUser(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) CreateUser(req *dto.UserRequest) (*model.User, error) {

	userModel := &model.User{
		Name:  req.Name,
		Email: req.Email,
	}
	if req.ProfileUrl != nil {
		userModel.ProfileUrl = *req.ProfileUrl
	}
	err := s.userRepository.Create(userModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return userModel, nil
}

func (s *UserService) UpdateUser(userId uuid.UUID, payload *dto.UserUpdateRequest) (*model.User, error) {
	user, err := s.userRepository.GetUser(userId)
	if err != nil {
		return nil, err
	}
	if payload.Name != nil {
		user.Name = *payload.Name
	}
	if payload.Email != nil {
		user.Email = *payload.Email
	}
	if payload.ProfileUrl != nil {
		user.ProfileUrl = *payload.ProfileUrl
	}
	if err := s.userRepository.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) DeleteUser(id uuid.UUID) error {
	if err := s.userRepository.Delete(id); err != nil {
		return err
	}
	return nil

}
