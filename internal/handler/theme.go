package handler

import (
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
)

type ThemeHandler struct {
	themeService *service.ThemeService
}

func NewThemeHandler(themeService *service.ThemeService) *ThemeHandler {
	return &ThemeHandler{
		themeService: themeService,
	}
}

func (h *ThemeHandler) GetThemeById(c *gin.Context) {
	themeID := c.Param("id")
	// log.Println("Received theme ID:", themeID)
	if themeID == " " {
		c.JSON(400, gin.H{"error": "theme id is required"})
		return
	}

	theme, err := h.themeService.GetByThemeID(themeID)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, theme)
}
