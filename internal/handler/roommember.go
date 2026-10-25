package handler

import (
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RoomMember struct {
	roomMemberService *service.RoomMemberService
}

func NewRoomMemberHandler(roomMemberService *service.RoomMemberService) *RoomMember {
	return &RoomMember{
		roomMemberService: roomMemberService,
	}
}

func (s *RoomMember) GetRoomDetail(c *gin.Context) {
	roomId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(400, "Bad Request")
		return
	}
	roomMembers, err := s.roomMemberService.GetRoomDetail(roomId)
	if err != nil {
		c.JSON(400, err)
		return
	}
	c.JSON(200, roomMembers)
}
