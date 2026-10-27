package service

import (
	"context"
	"errors"
	"fmt"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/model"
	"room-mate-api/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type AuthService struct {
	googleAuth *GoogleAuthService
	appleAuth  *AppleAuthService
	authRepo   repository.AuthRepo
	jwtSecret  string
	db         *gorm.DB
}

func NewAuthService(googleAuth *GoogleAuthService, appleAuth *AppleAuthService, authRepo repository.AuthRepo, jwtSecret string, db *gorm.DB) *AuthService {
	return &AuthService{
		googleAuth: googleAuth,
		appleAuth:  appleAuth,
		authRepo:   authRepo,
		jwtSecret:  jwtSecret,
		db:         db,
	}
}

func (s *AuthService) LoginWithGoogle(ctx context.Context, rawToken string) (*model.User, string, error) {
	googleUser, err := s.googleAuth.VerifyGoogleUser(ctx, rawToken)
	if err != nil {
		return nil, "", err
	}
	if googleUser.Email == "" || !googleUser.EmailVerified {
		return nil, "", errors.New("Google account must have a verified email")
	}
	identityUser, err := s.authRepo.FindByProvider("google", googleUser.ProviderUserID)
	var user *model.User
	if err == nil {
		// Returning user.
		user = &identityUser.User
	} else if !errors.Is(err, gorm.ErrRecordNotFound) { //check other err , except for not found for new user
		return nil, "", fmt.Errorf(
			"find auth identity: %w", err,
		)
	} else {
		//register process
		user = &model.User{
			Name:       googleUser.Name,
			Email:      googleUser.Email,
			ProfileUrl: googleUser.ProfileURL,
		}

		err = s.db.WithContext(ctx).Transaction(
			func(tx *gorm.DB) error {
				txUserRepo := repository.NewUserRepository(tx)
				txIdentityRepo := repository.NewAuthRepo(tx)

				if err := txUserRepo.Create(user); err != nil {
					return err
				}
				var identityUser = &model.AuthIdentity{
					Provider:       "google",
					ProviderUserId: googleUser.ProviderUserID,
					UserId:         user.ID,
				}
				if err := txIdentityRepo.Create(identityUser); err != nil {
					return err
				}

				return nil

			},
		)
		if err != nil {
			return nil, "", fmt.Errorf(
				"register Google user: %w", err,
			)
		}

	}

	// 4. Issue a Room Mate API JWT.
	accessToken, err := s.createAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil

}

func (s *AuthService) LoginWithApple(ctx context.Context, rawToken string) (*model.User, string, error) {
	appleUserInfo, err := s.appleAuth.GetAppleUserInfo(ctx, rawToken)
	if err != nil {
		return nil, "", err
	}
	if appleUserInfo.Email == "" || !appleUserInfo.EmailVerified {
		return nil, "", errors.New("Apple account must have a verified email")
	}
	// check already register or not
	existingUser, err := s.authRepo.FindByProvider("apple", appleUserInfo.ProviderUserID)
	var user *model.User
	if err == nil {
		// Returning user.
		user = &existingUser.User
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", fmt.Errorf(
			"find auth identity: %w", err,
		)
	} else {
		user = &model.User{
			Email:      appleUserInfo.Email,
			ProfileUrl: appleUserInfo.ProfileURL,
			Name:       appleUserInfo.Name,
		}
		err = s.db.WithContext(ctx).Transaction(
			func(tx *gorm.DB) error {
				txUserRepo := repository.NewUserRepository(tx)
				txIdentityRepo := repository.NewAuthRepo(tx)

				if err := txUserRepo.Create(user); err != nil {
					return err
				}
				var authIdenties = &model.AuthIdentity{
					Provider:       "apple",
					ProviderUserId: appleUserInfo.ProviderUserID,
					UserId:         user.ID,
				}

				if err := txIdentityRepo.Create(authIdenties); err != nil {
					return err
				}
				return nil
			})
	}

	if err != nil {
		return nil, "", fmt.Errorf(
			"register Apple user: %w", err,
		)
	}

	accessToken, err := s.createAccessToken(user)
	if err != nil {
		return nil, "", err
	}

	return user, accessToken, nil

}

func (s *AuthService) createAccessToken(user *model.User) (string, error) {
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Subject:   user.ID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
	}
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)
	signedToken, err := token.SignedString(
		[]byte(s.jwtSecret),
	)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// Used to map the database user into the API response.
func toUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         user.ID,
		Name:       user.Name,
		Email:      user.Email,
		ProfileUrl: user.ProfileUrl,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

}
