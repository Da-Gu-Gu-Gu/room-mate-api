package service

import (
	"context"
	"fmt"
	"room-mate-api/internal/dto"

	"github.com/coreos/go-oidc/v3/oidc"
)

type GoogleAuthService struct {
	verifier *oidc.IDTokenVerifier
}

func NewGoogleAuthService(clientID string) (*GoogleAuthService, error) {
	ctx := context.Background()

	provider, err := oidc.NewProvider(
		ctx,
		"https://accounts.google.com",
	)
	if err != nil {
		return nil, fmt.Errorf("initialize Google provider: %w", err)
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientID,
	})

	return &GoogleAuthService{
		verifier: verifier,
	}, nil
}

func (s *GoogleAuthService) VerifyIDToken(
	ctx context.Context,
	rawToken string,
) (*oidc.IDToken, error) {
	token, err := s.verifier.Verify(ctx, rawToken)
	if err != nil {
		return nil, fmt.Errorf("verify Google ID token: %w", err)
	}

	if token.Subject == "" {
		return nil, fmt.Errorf("Google ID token has no subject")
	}

	return token, nil
}

func (s *GoogleAuthService) VerifyGoogleUser(
	ctx context.Context,
	rawToken string,
) (*dto.GoogleUserInfo, error) {
	token, err := s.VerifyIDToken(ctx, rawToken)
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
		return nil, fmt.Errorf("read Google token claims: %w", err)
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("Google account has no email")
	}

	if !claims.EmailVerified {
		return nil, fmt.Errorf("Google email is not verified")
	}
	//google sdk user
	return &dto.GoogleUserInfo{
		ProviderUserID: token.Subject,
		Email:          claims.Email,
		Name:           claims.Name,
		ProfileURL:     claims.Picture,
		EmailVerified:  claims.EmailVerified,
	}, nil
}
