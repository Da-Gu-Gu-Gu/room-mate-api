package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"room-mate-api/internal/dto"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"
)

type TaskHandler struct{ service *service.TaskService }

func NewTaskHandler(s *service.TaskService) *TaskHandler { return &TaskHandler{service: s} }

func taskIDs(c *gin.Context) (uuid.UUID, uuid.UUID, uuid.UUID, bool) {
	userID, ok := currentUserID(c)
	if !ok {
		c.Abort()
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room ID"})
		return uuid.Nil, uuid.Nil, uuid.Nil, false
	}
	taskID := uuid.Nil
	if c.Param("taskId") != "" {
		taskID, err = uuid.Parse(c.Param("taskId"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
			return uuid.Nil, uuid.Nil, uuid.Nil, false
		}
	}
	return roomID, taskID, userID, true
}

func taskPage(c *gin.Context) (int, int, bool) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if err != nil || limit < 1 || limit > 100 {
		c.JSON(400, gin.H{"error": "limit must be between 1 and 100"})
		return 0, 0, false
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		c.JSON(400, gin.H{"error": "offset must be nonnegative"})
		return 0, 0, false
	}
	return limit, offset, true
}

func taskError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrTaskInvalid):
		c.JSON(400, gin.H{"error": err.Error()})
	case errors.Is(err, repository.ErrRoomMemberForbidden):
		c.JSON(403, gin.H{"error": "task access forbidden"})
	case errors.Is(err, gorm.ErrRecordNotFound):
		c.JSON(404, gin.H{"error": "task or occurrence not found"})
	case errors.Is(err, repository.ErrTaskConflict):
		c.JSON(409, gin.H{"error": err.Error()})
	default:
		c.JSON(500, gin.H{"error": "internal server error"})
	}
}

func (h *TaskHandler) Create(c *gin.Context) {
	roomID, _, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid task payload"})
		return
	}
	task, err := h.service.Create(c.Request.Context(), roomID, userID, req)
	if err != nil {
		taskError(c, err)
		return
	}
	c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) List(c *gin.Context) {
	roomID, _, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	limit, offset, ok := taskPage(c)
	if !ok {
		return
	}
	tasks, err := h.service.Repository.List(c.Request.Context(), roomID, userID, limit, offset)
	if err != nil {
		taskError(c, err)
		return
	}
	c.JSON(200, gin.H{"tasks": tasks})
}

func (h *TaskHandler) Get(c *gin.Context) {
	roomID, taskID, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	task, err := h.service.Repository.Get(c.Request.Context(), roomID, taskID, userID)
	if err != nil {
		taskError(c, err)
		return
	}
	c.JSON(200, task)
}

func (h *TaskHandler) Occurrences(c *gin.Context) {
	roomID, taskID, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	limit, offset, ok := taskPage(c)
	if !ok {
		return
	}
	occurrences, err := h.service.Repository.Occurrences(c.Request.Context(), roomID, taskID, userID, limit, offset)
	if err != nil {
		taskError(c, err)
		return
	}
	c.JSON(200, gin.H{"occurrences": occurrences})
}

func (h *TaskHandler) Complete(c *gin.Context) {
	roomID, taskID, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	occurrenceID, err := uuid.Parse(c.Param("occurrenceId"))
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid occurrence ID"})
		return
	}
	var req dto.CompleteTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid completion payload"})
		return
	}
	occurrence, err := h.service.Complete(c.Request.Context(), roomID, taskID, occurrenceID, userID, req)
	if err != nil {
		taskError(c, err)
		return
	}
	c.JSON(200, occurrence)
}

func (h *TaskHandler) Status(c *gin.Context) {
	roomID, taskID, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || (req.Status != "active" && req.Status != "paused" && req.Status != "archived") {
		c.JSON(400, gin.H{"error": "status must be active, paused, or archived"})
		return
	}
	if err := h.service.Repository.UpdateStatus(c.Request.Context(), roomID, taskID, userID, req.Status, time.Now()); err != nil {
		taskError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *TaskHandler) Members(c *gin.Context) {
	roomID, taskID, userID, ok := taskIDs(c)
	if !ok {
		return
	}
	var req struct {
		MemberIDs []uuid.UUID `json:"memberIds"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid members payload"})
		return
	}
	if err := h.service.ReplaceMembers(c.Request.Context(), roomID, taskID, userID, req.MemberIDs); err != nil {
		taskError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
