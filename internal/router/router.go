package router

import (
	"backend/internal/handler"
	"backend/internal/middleware"
	"backend/internal/repository"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRoutes 设置路由
func SetupRoutes(userHandler *handler.UserHandler, fileHandler *handler.FileHandler, jwtSvc service.JwtService, blacklistRepo repository.AccessTokenBlacklistRepository) *gin.Engine {
	// 创建Gin引擎
	r := gin.Default()

	// 添加中间件
	r.Use(CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// 静态文件服务（用于本地文件访问）
	r.Static("/uploads", "./uploads")

	// Swagger文档路由
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// API版本组
	v1 := r.Group("/api/v1")
	{
		// 用户相关路由
		users := v1.Group("/users")
		{
			users.POST("/register", userHandler.Register)
			users.POST("/login", userHandler.Login)
			users.POST("/refresh", userHandler.RefreshToken)
			users.POST("/logout", userHandler.Logout)
			users.POST("/send-code", userHandler.SendVerificationCode)
			users.POST("/send-reset-code", userHandler.SendResetPasswordCode)
			users.POST("/reset-password", userHandler.ResetPassword)
			
			// 需要认证的路由
			authRoutes := users.Group("/").Use(middleware.AuthMiddleware(jwtSvc, blacklistRepo))
			authRoutes.GET("/me", userHandler.GetMe)
			authRoutes.PUT("/me", userHandler.UpdateProfile)

			// 公开路由（不受保护）
			users.GET("/:id", userHandler.GetUserByID)
			users.GET("/username/:username", userHandler.GetUserByUsername)
		}

		// 文件相关路由
		files := v1.Group("/files")
		{
			// 公开路由
			files.GET("/public", fileHandler.GetPublicFiles)
			files.GET("/storages", fileHandler.GetStorageInfo)
			files.GET("/:id", fileHandler.GetFile) // 支持公开和私有文件访问

			// 需要认证的路由
			authFileRoutes := files.Group("/").Use(middleware.AuthMiddleware(jwtSvc, blacklistRepo))
			authFileRoutes.POST("/upload", fileHandler.UploadFile)
			authFileRoutes.POST("/upload-multiple", fileHandler.UploadFiles)
			authFileRoutes.GET("/my", fileHandler.GetUserFiles)
			authFileRoutes.PUT("/:id", fileHandler.UpdateFile)
			authFileRoutes.DELETE("/:id", fileHandler.DeleteFile)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		response.SuccessResponse(c, 200, "服务正常", gin.H{
			"status": "ok",
			"service": "backend",
		})
	})

	return r
}

// CORSMiddleware 跨域中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
} 