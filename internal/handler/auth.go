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

func (h *AuthHandler) LoginWithApple(c *gin.Context) {
	ctx := c.Request.Context()
	var rawToken dto.SocialLoginRequest
	err := c.ShouldBindJSON(&rawToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	user, accessToken, err := h.authService.LoginWithApple(ctx, rawToken.IdToken)
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
			Email:      user.Email,
			ProfileUrl: user.ProfileUrl,
			Name:       user.Name,
		},
	})
}
