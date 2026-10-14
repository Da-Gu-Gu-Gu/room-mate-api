package handler

import (
	"log"
	"net/http"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
	themeID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid theme id"})
		return
	}
	// log.Println("Received theme ID:", themeID)
	if themeID == uuid.Nil {
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

func (h *ThemeHandler) GetThemes(c *gin.Context) {
	thems, err := h.themeService.GetThemes()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, thems)
}

func (h *ThemeHandler) CreateTheme(c *gin.Context) {
	var req dto.ThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Print("here got error")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	theme, err := h.themeService.CreateTheme(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, theme)
}

func (h *ThemeHandler) UpdateTheme(c *gin.Context) {
	themeId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid theme id"})
		return
	}
	var req dto.UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedTheme, err := h.themeService.UpdateTheme(themeId, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, updatedTheme)

}

func (h *ThemeHandler) DeleteTheme(c *gin.Context) {
	themeId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid theme id"})
		return
	}
	err = h.themeService.DeleteTheme(themeId)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Theme deleted successfully"})

}
