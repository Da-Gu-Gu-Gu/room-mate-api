package repository

import (
	"errors"
	"fmt"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

type UserRepository interface {
	GET() ([]*model.User, error)
	GetUser(id uuid.UUID) (*model.User, error)
	Create(user *model.User) error
	Update(user *model.User) error
	Delete(id uuid.UUID) error
	CreateUserTheme(id uuid.UUID, req *dto.UserThemeRequest) (*model.UserTheme, error)
	GetUserTheme(id uuid.UUID) (*model.UserTheme, error)
	GetAboutMe(id uuid.UUID) (*dto.MeResponse, error)
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

// user
func (r *userRepository) GET() ([]*model.User, error) {
	var userList []*model.User
	if err := r.db.Find(&userList).Error; err != nil {
		return nil, err
	}
	return userList, nil

}

func (r *userRepository) GetUser(id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *model.User) error {
	err := r.db.Create(&user).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Update(user *model.User) error {
	if err := r.db.Save(user).Error; err != nil {
		return err
	}
	return nil
}

func (r *userRepository) Delete(id uuid.UUID) error {
	if err := r.db.Delete(&model.User{}, id).Error; err != nil {
		return err
	}
	return nil
}

// end-user

// user-with-theme
func (r *userRepository) CreateUserTheme(id uuid.UUID, req *dto.UserThemeRequest) (*model.UserTheme, error) {

	userTheme := &model.UserTheme{
		UserID:  id,
		ThemeID: req.ThemeID,
	}
	err := r.db.Create(userTheme).Error
	if err != nil {
		return nil, err
	}
	err = r.db.Preload("User").Preload("Theme").First(userTheme, "id = ? ", userTheme.ID).Error
	if err != nil {
		return nil, err
	}
	return userTheme, nil
}

func (r *userRepository) GetUserTheme(id uuid.UUID) (*model.UserTheme, error) {
	var userTheme *model.UserTheme
	err := r.db.Preload("User").Preload("Theme").First(&userTheme, "user_id= ? ", id).Error
	if err != nil {
		return nil, err
	}
	return userTheme, err
}

func (r *userRepository) GetAboutMe(id uuid.UUID) (*dto.MeResponse, error) {
	// 1. The user must exist.
	user, err := r.GetUser(id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	response := &dto.MeResponse{
		User: dto.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			Email:      user.Email,
			ProfileUrl: user.ProfileUrl,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		},
	}

	// 2. A custom theme is optional.
	userTheme, err := r.GetUserTheme(id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get user theme: %w", err)
	}

	if err == nil && userTheme != nil {
		response.Theme = &dto.ThemeResponse{
			PrimaryColor:    userTheme.Theme.PrimaryColor,
			SecondaryColor:  userTheme.Theme.SecondaryColor,
			BackgroundColor: userTheme.Theme.BackgroundColor,
			BackgroundImage: userTheme.Theme.BackgroundImgUrl,
			CreatedAt:       userTheme.Theme.CreatedAt,
			UpdatedAt:       userTheme.Theme.UpdatedAt,
		}
	}

	// 3. A room is optional. Find the user's active ,
	// not just rooms that the user created.
	var room model.Room

	err = r.db.
		Table("rooms").
		Select("rooms.*").
		Joins("JOIN room_members ON room_members.room_id = rooms.id").
		Where(
			"room_members.user_id = ? AND room_members.leave_at IS NULL",
			id,
		).
		Order("room_members.joined_at DESC").
		First(&room).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("get user's room: %w", err)
	}

	if err == nil {
		response.Room = &dto.RoomResponse{
			RoomName:  room.RoomName,
			RoomCode:  room.RoomCode,
			CreatedAt: room.CreatedAt,
			UpdatedAt: room.UpdatedAt,
		}
	}

	return response, nil
}
