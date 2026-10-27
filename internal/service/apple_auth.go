package service

import (
	"context"
	"fmt"
	"room-mate-api/internal/dto"

	"github.com/coreos/go-oidc/v3/oidc"
)

type AppleAuthService struct {
	verifier *oidc.IDTokenVerifier
}

func NewAppleAuthService(clientID string) (*AppleAuthService, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, "https://appleid.apple.com") //replace with apple

	if err != nil {
		return nil, err
	}

	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &AppleAuthService{
		verifier: idTokenVerifier,
	}, nil
}

func (s *AppleAuthService) GetAppleUserInfo(ctx context.Context, rawIdToken string) (*dto.AppleUserInfo, error) {
	token, err := s.verifier.Verify(ctx, rawIdToken)
	if err != nil {
		return nil, err
	}

	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := token.Claims(&claims); err != nil {
		// handle error
		return nil, err
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("Apple account has no email")
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("Apple email is not verified")
	}
	//google sdk user
	return &dto.AppleUserInfo{
		ProviderUserID: token.Subject,
		Email:          claims.Email,
		Name:           claims.Name,
		ProfileURL:     claims.Picture,
		EmailVerified:  claims.EmailVerified,
	}, nil

}
