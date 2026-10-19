package dto

type GoogleUserInfo struct {
	ProviderUserID string
	Email          string
	Name           string
	ProfileURL     string
	EmailVerified  bool
}

type SocialLoginRequest struct {
	IdToken string `json:"idToken" binding:"required"`
}

type AuthResponse struct {
	AccessToken string
	TokenType   string
	User        UserResponse
}
