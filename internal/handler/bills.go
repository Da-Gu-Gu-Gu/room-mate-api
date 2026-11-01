package handler

import (
	"net/http"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
)

type BillHandler struct {
	billService *service.BillService
}

func NewBillHandler(billService *service.BillService) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

func (h *BillHandler) GetBillCategories(c *gin.Context) {
	categories, err := h.billService.GetBillCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve bill categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}
