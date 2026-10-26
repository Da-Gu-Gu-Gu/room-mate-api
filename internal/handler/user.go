package handler

import (
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/middleware"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (h *UserHandler) GetUserById(c *gin.Context) {
	actorID, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return
	}
	userId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	if userId != actorID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you can only access your own profile"})
		return
	}
	user, err := h.userService.GetUserById(userId)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, user)
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	var req dto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "Bad request")
		return
	}
	user, err := h.userService.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)

}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var userPayload dto.UserUpdateRequest
	if err := c.ShouldBindJSON(&userPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser, err := h.userService.UpdateUser(userId, &userPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(200, updatedUser)

}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	if err := h.userService.DeleteUser(userId); err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(200, "User deleted successfully")

}

func (h *UserHandler) CreateUserTheme(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var req *dto.UserThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, "Bad Request")
		return
	}

	userTheme, err := h.userService.CreateUserTheme(userId, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(200, userTheme)
}

func (h *UserHandler) GetUserTheme(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(400, "Bad Request")
		return
	}
	user, err := h.userService.GetUserTheme(userId)
	if err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}
	c.JSON(200, user)
}

func (h *UserHandler) GetAboutMe(c *gin.Context) {
	userId, err := uuid.Parse(c.GetString(middleware.UserIDContextKey))
	if err != nil {
		c.JSON(400, "Bad Request")
		return
	}
	//
	response, err := h.userService.GetAboutMe(userId)
	if err != nil {
		c.JSON(400, err.Error())
		return
	}
	c.JSON(200, response)
}
