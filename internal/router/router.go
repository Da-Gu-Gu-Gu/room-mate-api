package router

import (
	"room-mate-api/internal/handler"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type routeHandler struct {
	themeHandler *handler.ThemeHandler
}

func SetupRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	themeRepository := repository.NewThemeRepository(db)
	themeService := service.NewThemeService(themeRepository)
	themeHandler := handler.NewThemeHandler(themeService)

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(*&userRepository)
	userHandler := handler.NewUserHandler(*&userService)

	roomRepository := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepository)
	roomHandler := handler.NewRoomHandler(roomService)

	api := router.Group("api/v1")
	{
		user := api.Group("users")
		{
			user.GET("/:id", userHandler.GetUserById)
			user.POST("", userHandler.CreateUser)
			user.PUT("/:id", userHandler.UpdateUser)
			user.DELETE("/:id", userHandler.DeleteUser)
			user.GET("", userHandler.GetUsers) //list
		}

		theme := api.Group("themes")
		{
			theme.GET("/:id", themeHandler.GetThemeById)
			theme.POST("", themeHandler.CreateTheme)
			theme.PUT("/:id", themeHandler.UpdateTheme)
			theme.DELETE("/:id", themeHandler.DeleteTheme)
			theme.GET("", themeHandler.GetThemes) //list
		}

		room := api.Group("rooms")
		{
			room.GET("/:id", roomHandler.GetRoomById)
			room.POST("", roomHandler.CreateRoom)
			room.PUT("/:id", roomHandler.UpdateRoom)
			room.DELETE("/:id", roomHandler.DeleteRoom)
			room.GET("", roomHandler.GetRooms) //list
		}
	}
	// Define your routes here
	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, World!",
		})
	})

	return router
}

func SampleRouteHandler(c *gin.Context) {
	c.JSON(501, gin.H{
		"message": "This is a sample route",
	})
}
