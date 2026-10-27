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

func SetupRouter(db *gorm.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	themeRepository := repository.NewThemeRepository(db)
	themeService := service.NewThemeService(themeRepository)
	themeHandler := handler.NewThemeHandler(themeService)

	userRepository := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	roomRepository := repository.NewRoomRepository(db)
	roomService := service.NewRoomService(roomRepository)
	roomHandler := handler.NewRoomHandler(roomService)

	googleAuth, err := service.NewGoogleAuthService(cfg.GoogleClientId)
	if err != nil {
		log.Fatal(err)
	}
	appleAuth, err := service.NewAppleAuthService(cfg.AppleClientId)
	if err != nil {
		log.Fatal(err)
	}
	authIdentityRepo := repository.NewAuthRepo(db)
	authService := service.NewAuthService(googleAuth, appleAuth, authIdentityRepo, cfg.JWTSecret, db)
	authHandler := handler.NewAuthHandler(authService)

	api := router.Group("/api/v1")
	api.POST("/auth/google", authHandler.GoogleLogin)
	api.POST("/auth/apple", authHandler.LoginWithApple)

	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		protected.GET("/me", userHandler.GetAboutMe)

		user := protected.Group("/users")
		{
			user.PUT("", userHandler.UpdateUser)
			user.GET("/:id", userHandler.GetUserById)
			user.DELETE("", userHandler.DeleteUser)
			user.POST("/theme", userHandler.CreateUserTheme)
			user.GET("/theme", userHandler.GetUserTheme)
		}

		// Theme catalog is readable by authenticated users. Mutating global themes
		// is intentionally not exposed here; it should be an admin-only API.
		theme := protected.Group("/themes")
		theme.GET("", themeHandler.GetThemes)

		room := protected.Group("/rooms")
		{
			room.POST("", roomHandler.CreateRoom)
			room.PUT("/:id", roomHandler.UpdateRoom)
			room.GET("/:id", roomHandler.GetRoomById)
			room.DELETE("/:id", roomHandler.DeleteRoom)
		}
	}

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})
	return router
}

func SampleRouteHandler(c *gin.Context) {
	c.JSON(501, gin.H{"message": "This is a sample route"})
}
