package router

import (
	"context"
	"log"

	"room-mate-api/internal/config"
	"room-mate-api/internal/handler"
	"room-mate-api/internal/middleware"
	"room-mate-api/internal/repository"
	"room-mate-api/internal/service"
	"room-mate-api/internal/storage"

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

	roomMemberRepository := repository.NewRoomMemberRepository(db)
	roomMemberService := service.NewRoomMemberService(roomMemberRepository)
	roomMemberHandler := handler.NewRoomMemberHandler(roomMemberService)

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

	billRepository := repository.NewBillRepository(db)
	billService := service.NewBillService(billRepository)
	billHandler := handler.NewBillHandler(billService)

	var imageStorage service.ImageObjectStorage
	var roomBillStorage service.RoomBillObjectStorage
	if cfg.AWSRegion != "" && cfg.AWSS3Bucket != "" {
		s3Storage, err := storage.NewS3Storage(context.Background(), cfg.AWSRegion, cfg.AWSS3Bucket)
		if err != nil {
			log.Fatal(err)
		}
		imageStorage = s3Storage
		roomBillStorage = s3Storage
	}

	roomBillRepository := repository.NewRoomBillRepository(db)
	roomBillService := service.NewRoomBillService(roomBillRepository, roomBillStorage)
	roomBillHandler := handler.NewRoomBillHandler(roomBillService)
	uploadService := service.NewUploadService(imageStorage)
	uploadHandler := handler.NewUploadHandler(uploadService)

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

		theme := protected.Group("/themes")
		theme.GET("", themeHandler.GetThemes)

		protected.POST("/uploads/images", uploadHandler.UploadImage)

		room := protected.Group("/rooms")
		{
			room.POST("", roomHandler.CreateRoom)
			room.PUT("/:id", roomHandler.UpdateRoom)
			room.GET("/:id", roomHandler.GetRoomById)
			room.DELETE("/:id", roomHandler.DeleteRoom)
			room.GET("/:id/members", roomMemberHandler.GetRoomDetail)
			room.POST("/:id/invitations", roomMemberHandler.Invite)
			room.POST("/:id/transfer-ownership", roomMemberHandler.TransferOwnership)
			room.POST("/:id/leave", roomMemberHandler.LeaveRoom)

			// Bills
			room.POST("/:id/bills", roomBillHandler.CreateRoomBill)
			room.GET("/:id/bills", roomBillHandler.GetRoomBills)
			room.GET("/:id/bills/:billId", roomBillHandler.GetRoomBillByID)
			room.POST("/:id/bills/:billId/shares", roomBillHandler.CreateRoomBillShares)
			room.POST("/:id/bills/:billId/payments", roomBillHandler.CreateRoomBillPayment)
			room.POST("/:id/bills/:billId/payments/:paymentId/approve", roomBillHandler.ApproveRoomBillPayment)
			room.POST("/:id/bills/:billId/payments/:paymentId/reject", roomBillHandler.RejectRoomBillPayment)

		}

		protected.GET("/invitations", roomMemberHandler.GetMyInvitations)
		protected.POST("/invitations/:invitationId/accept", roomMemberHandler.AcceptInvitation)
		protected.POST("/invitations/:invitationId/decline", roomMemberHandler.DeclineInvitation)

		protected.GET("/bill-categories", billHandler.GetBillCategories)
	}

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello, World!"})
	})
	return router
}

func SampleRouteHandler(c *gin.Context) {
	c.JSON(501, gin.H{"message": "This is a sample route"})
}
