package handler

import (
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(
	authService *service.AuthService,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	var req dto.SocialLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	user, accessToken, err := h.authService.LoginWithGoogle(
		c.Request.Context(),
		// "eyJhbGciOiJSUzI1NiIsImtpZCI6Ijk0M2EzYTVkN2Q5MTk2MjVhNDU0ZTQ4OWI3NWMyOWFkYWI1N2FjYmEiLCJ0eXAiOiJKV1QifQ.eyJpc3MiOiJodHRwczovL2FjY291bnRzLmdvb2dsZS5jb20iLCJhenAiOiI0MDc0MDg3MTgxOTIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJhdWQiOiI0MDc0MDg3MTgxOTIuYXBwcy5nb29nbGV1c2VyY29udGVudC5jb20iLCJzdWIiOiIxMDQyNTE3Njk1MzQyNjQ0MDUxMTgiLCJlbWFpbCI6ImhlaW5odGV0YXhuZ0BnbWFpbC5jb20iLCJlbWFpbF92ZXJpZmllZCI6dHJ1ZSwiYXRfaGFzaCI6InJBb013NlZ6dXozdDNkZlcwMmNrNHciLCJuYW1lIjoiSGVpbiBIdGV0IEF1bmciLCJwaWN0dXJlIjoiaHR0cHM6Ly9saDMuZ29vZ2xldXNlcmNvbnRlbnQuY29tL2EvQUNnOG9jS2FWdmo0WnJxM0h2dzNHTzgtUHVyQ1ZnTVMwaTUxSHhuUVB0TVF6cTgzYlRiZHU3Yz1zOTYtYyIsImdpdmVuX25hbWUiOiJIZWluIEh0ZXQgQXVuZyIsImlhdCI6MTc5MjQyNDk1NSwiZXhwIjoxNzkyNDI4NTU1fQ.PXpJSiVHflPPnDoOe97HXjWaQQjYTVzhSDnOlpPYIVXTGF2BBSg29R79HKHZLiAjIauvHURWe8Y9OPLXCEqFv31TSxJOJZKpa5Svj5Zvks5pSPw8Hl8HUPPFMO3C8udu51Q5GoPHh5qDIwI83QBtn8m7C4WwXCfc5emaySY7B29gQYzejma6A3ntw4pbe5q9w7otTM-06XDcXQ1xPUjG7CFmwx3t0dVCiD-A-jqczHuTHPuJLIb_bSp_CghVsBE6wDkvUcLhOUwmA1Zi7FA103MqwJOCdBlkx2IUcJ9PnMDOIbTl2WO7N9iydMHYNHcRQasbGKYN6qoCxqw5MWsGpw",
		req.IdToken,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		User: dto.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			Email:      user.Email,
			ProfileUrl: user.ProfileUrl,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		},
	})
}
