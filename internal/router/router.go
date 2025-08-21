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
func SetupRoutes(userHandler *handler.UserHandler, fileHandler *handler.FileHandler, adminHandler *handler.AdminHandler, friendHandler *handler.FriendHandler, wsHandler *handler.WSHandler, jwtSvc service.JwtService, blacklistRepo repository.AccessTokenBlacklistRepository) *gin.Engine {
	// 创建Gin引擎
	r := gin.Default()

	// 添加中间件
	r.Use(CORSMiddleware())
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 全局流量统计（入口/出口字节）
	r.Use(middleware.TrafficMiddleware())

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
			users.POST("/send-activation-code", userHandler.SendActivationCode)
			users.POST("/activate", userHandler.ActivateAccount)
			users.GET("/me", middleware.AuthMiddleware(jwtSvc, blacklistRepo), userHandler.GetMe)
			users.PUT("/me", middleware.AuthMiddleware(jwtSvc, blacklistRepo), userHandler.UpdateProfile)
			users.GET("/username/:username", userHandler.GetUserByUsername)
			users.GET("/:id", userHandler.GetUserByID)
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

		// 好友相关路由（需要认证）
		friends := v1.Group("/friends")
		friends.Use(middleware.AuthMiddleware(jwtSvc, blacklistRepo))
		{
			friends.POST("/requests", friendHandler.CreateRequest)
			friends.GET("/requests/incoming", friendHandler.ListIncomingRequests)
			friends.GET("/requests/outgoing", friendHandler.ListOutgoingRequests)
			friends.POST("/requests/:id/accept", friendHandler.AcceptRequest)
			friends.POST("/requests/:id/reject", friendHandler.RejectRequest)
			friends.DELETE("/requests/:id", friendHandler.CancelRequest)

			friends.GET("/list", friendHandler.ListFriends)
			friends.PATCH("/remarks/:friend_id", friendHandler.UpdateRemark)
			friends.DELETE("/:friend_id", friendHandler.DeleteFriend)

			friends.POST("/blocks/:user_id", friendHandler.Block)
			friends.DELETE("/blocks/:user_id", friendHandler.Unblock)
			friends.GET("/blocks", friendHandler.ListBlocks)
		}

		// WebSocket 路由（鉴权由 handler 内部处理：支持 Authorization 头或 query token）
		v1.GET("/ws/chat", wsHandler.Chat)

		// 管理员相关路由
		admin := v1.Group("/admin")
		{
			admin.POST("/login", adminHandler.Login)

			// 需要管理员认证的路由
			authAdminRoutes := admin.Group("/")
			authAdminRoutes.Use(middleware.AdminAuthMiddleware(jwtSvc))
			authAdminRoutes.GET("/dashboard", func(c *gin.Context) {
				adminUsername, _ := c.Get("admin_username")
				response.SuccessResponse(c, 200, "欢迎来到管理员面板", gin.H{
					"username": adminUsername,
				})
			})
			// 刷新管理员Token
			authAdminRoutes.POST("/refresh-token", adminHandler.AdminRefreshToken)
			// 用户管理相关路由
			authAdminRoutes.GET("/users", adminHandler.GetUsers)
			authAdminRoutes.GET("/users/:id", adminHandler.GetUserDetail)
			authAdminRoutes.PUT("/users/:id/status", adminHandler.UpdateUserStatus)
			authAdminRoutes.PUT("/users/:id/password", adminHandler.UpdateUserPassword)
			authAdminRoutes.DELETE("/users/:id", adminHandler.DeleteUser)
			// 好友功能封禁（管理员）
			authAdminRoutes.POST("/users/:id/friend-ban", adminHandler.AdminSetFriendBan)
			authAdminRoutes.DELETE("/users/:id/friend-ban", adminHandler.AdminRemoveFriendBan)
			authAdminRoutes.GET("/users/:id/friend-ban", adminHandler.AdminGetFriendBan)
			authAdminRoutes.GET("/stats/users", adminHandler.GetUserStats)
			// 用户行为日志（按用户）
			authAdminRoutes.GET("/users/:id/action-logs", adminHandler.ListUserActionLogs)

			// 文件管理相关路由（管理员）
			authAdminRoutes.GET("/files", adminHandler.AdminListFiles)
			authAdminRoutes.GET("/files/public", adminHandler.AdminListPublicFiles)
			authAdminRoutes.GET("/files/:id", adminHandler.AdminGetFile)
			authAdminRoutes.PUT("/files/:id", adminHandler.AdminUpdateFile)
			authAdminRoutes.DELETE("/files/:id", adminHandler.AdminDeleteFile)
			// 存储信息（管理员）
			authAdminRoutes.GET("/storage/info", adminHandler.AdminGetStorageInfo)

			// 管理员日志相关路由
			authAdminRoutes.POST("/logs", adminHandler.CreateAdminLog)
			authAdminRoutes.GET("/logs", adminHandler.ListAdminLogs)
			// 管理员统计：网络流量
			authAdminRoutes.GET("/stats/traffic", adminHandler.GetTrafficStats)
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
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
} 