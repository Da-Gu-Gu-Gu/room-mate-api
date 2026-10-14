package handler

import (
	"net/http"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUsers(c *gin.Context) {
	userList, err := h.userService.GetUsers()
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Something Wrong with user",
		})
		return
	}
	c.JSON(200, userList)

}
