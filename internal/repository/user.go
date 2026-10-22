package repository

import (
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
	CreateUserTheme(id uuid.UUID, req *dto.UserThemeRequest) *model.UserTheme
	GetUserTheme(id uuid.UUID) (*model.UserTheme, error)
	GetAboutMe(id uuid.UUID) (*dto.MeResponse, error)
}

func NewUserRepository(db *gorm.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

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

func (r *userRepository) CreateUserTheme(id uuid.UUID, req *dto.UserThemeRequest) *model.UserTheme {

	userTheme := &model.UserTheme{
		UserID:  id,
		ThemeID: req.ThemeID,
	}
	err := r.db.Create(userTheme).Error
	if err != nil {
		return nil
	}
	err = r.db.Preload("User").Preload("Theme").First(userTheme, "id = ? ", userTheme.ID).Error
	if err != nil {
		return nil
	}
	return userTheme
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
	userTheme, err := r.GetUserTheme(id)
	if err != nil {
		return nil, err
	}
	var userRoom model.Room
	err = r.db.First(&userRoom, "created_by= ? ", id).Error
	if err != nil {
		return nil, err
	}

	return &dto.MeResponse{
		User: dto.UserResponse{
			Name:       userTheme.User.Name,
			Email:      userTheme.User.Email,
			ProfileUrl: userTheme.User.ProfileUrl,
		},
		Room: dto.RoomResponse{
			RoomName: userRoom.RoomName,
			RoomCode: userRoom.RoomCode,
		},
		Theme: dto.ThemeResponse{
			PrimaryColor:    userTheme.Theme.PrimaryColor,
			SecondaryColor:  userTheme.Theme.SecondaryColor,
			BackgroundColor: userTheme.Theme.BackgroundColor,
		},
	}, nil
}
