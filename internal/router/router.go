package router

import (
	"log"
	"room-mate-api/internal/config"
	"room-mate-api/internal/handler"
	"room-mate-api/internal/middleware"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type routeHandler struct {
	themeHandler *handler.ThemeHandler
}

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {

	router := gin.Default()

	themeRepository := repository.NewThemeRepository(db)
	themeService := service.NewThemeService(themeRepository)
	themeHandler := handler.NewThemeHandler(themeService)

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(*&userService)

	roomRepository := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepository)
	roomHandler := handler.NewRoomHandler(roomService)

	roomMemberRepository := repository.NewRoomMemberRepository(db)
	roomMemberService := service.NewRoomMemberService(roomMemberRepository)
	roomMemberHandler := handler.NewRoomMemberHandler(roomMemberService)

	googleAuth, err := service.NewGoogleAuthService(
		cfg.GoogleClientId,
	)
	if err != nil {
		log.Fatal(err)
	}

	authIdentityRepo := repository.NewAuthRepo(db)

	authService := service.NewAuthService(
		googleAuth,
		authIdentityRepo,
		cfg.JWTSecret,
		db,
	)

	authHandler := handler.NewAuthHandler(authService)

	api := router.Group("api/v1")
	// Public route: users need this to sign in.
	api.POST("/auth/google", authHandler.GoogleLogin)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	{

		protected.GET("/me", userHandler.GetAboutMe)

		user := protected.Group("users")
		{

			user.PUT("", userHandler.UpdateUser)

			// should move to admin portal
			user.POST("", userHandler.CreateUser)
			user.GET("/:id", userHandler.GetUserById)
			user.GET("", userHandler.GetUsers)
			user.DELETE("", userHandler.DeleteUser)

			//theme
			user.POST("/theme", userHandler.CreateUserTheme) //setting theme to user
			user.GET("/theme", userHandler.GetUserTheme)     //

		}

		theme := protected.Group("themes")
		{
			theme.GET("", themeHandler.GetThemes) //list

			// should move to admin side
			theme.GET("/:id", themeHandler.GetThemeById)
			theme.DELETE("/:id", themeHandler.DeleteTheme)
			theme.POST("", themeHandler.CreateTheme)
			theme.PUT("/:id", themeHandler.UpdateTheme)
		}

		room := protected.Group("rooms")
		{

			room.POST("", roomHandler.CreateRoom)
			room.PUT("/:id", roomHandler.UpdateRoom)
			room.GET("/:id", roomHandler.GetRoomById)

			room.GET("/:id/member", roomMemberHandler.GetRoomDetail) //member,room only now, next time will get a lot of data

			//should move to admin portal
			room.GET("", roomHandler.GetRooms) //list
			room.DELETE("/:id", roomHandler.DeleteRoom)
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
