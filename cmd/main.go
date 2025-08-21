// Package main provides the entry point for the backend application.
//
//	@title			Backend API
//	@version		1.0
//	@description	这是一个用户注册和认证系统的后端API
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.url	http://www.example.com/support
//	@contact.email	support@example.com
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api/v1
//
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							header
//	@name						Authorization
//	@description				Description for what is this security definition being used
//
//	@schemes	http https
package main

import (
	_ "backend/docs" // 导入生成的docs包
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func startPanelServer() {
	panelMux := http.NewServeMux()
	panelDir := "./panel"

	// Check if panel directory exists
	if _, err := os.Stat(panelDir); os.IsNotExist(err) {
		log.Printf("管理面板目录 '%s' 不存在，将不启动管理面板服务。", panelDir)
		return
	}

	panelFS := http.FileServer(http.Dir(panelDir))
	panelMux.Handle("/", panelFS)

	port := getEnv("PANEL_PORT", "8081")
	log.Printf("管理面板服务器启动在端口 %s", port)
	// 使用 http.ListenAndServe 启动服务
	if err := http.ListenAndServe(":"+port, panelMux); err != nil {
		log.Printf("管理面板服务器启动失败: %v", err)
	}
}

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("未找到 .env 文件，将使用系统环境变量")
	}

	// 初始化数据库
	db, err := config.InitDatabase()
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	// 初始化Redis
	rdb, err := config.InitRedis()
	if err != nil {
		log.Fatalf("Redis初始化失败: %v", err)
	}

	// 初始化仓储层
	userRepo := repository.NewUserRepository(db)
	fileRepo := repository.NewFileRepository(db)
	codeRepo := repository.NewCodeRepository(rdb)
	refreshTokenRepo := repository.NewRefreshTokenRepository(rdb)
	rateLimitRepo := repository.NewRateLimitRepository(rdb)
	accessTokenBlacklistRepo := repository.NewAccessTokenBlacklistRepository(rdb)
	deviceRepo := repository.NewDeviceRepository(db)
	adminLogRepo := repository.NewAdminLogRepository(db)
	userActionLogRepo := repository.NewUserActionLogRepository(db)

	// 初始化服务层
	securityCfg := config.GetSecurityConfig()
	fileStorageCfg := config.GetFileStorageConfig()
	smtpCfg := config.GetSMTPConfig()
	log.Printf("启动时SMTP配置: host=%s port=%d username=%s from=%s password_set=%t", smtpCfg.Host, smtpCfg.Port, smtpCfg.Username, smtpCfg.From, smtpCfg.Password != "")
	mailSvc := service.NewMailService(smtpCfg)
	jwtSvc := service.NewJwtService(securityCfg)
	fileStorageSvc := service.NewFileStorageService(fileStorageCfg)
	userService := service.NewUserService(userRepo, deviceRepo, codeRepo, refreshTokenRepo, rateLimitRepo, accessTokenBlacklistRepo, mailSvc, jwtSvc, securityCfg)
	fileService := service.NewFileService(fileRepo, fileStorageSvc)
	adminLogService := service.NewAdminLogService(adminLogRepo)
	userActionLogService := service.NewUserActionLogService(userActionLogRepo)
	adminCfg := config.GetAdminConfig()

	// 初始化处理器层
	userHandler := handler.NewUserHandler(userService, userActionLogService)
	fileHandler := handler.NewFileHandler(fileService)
	adminHandler := handler.NewAdminHandler(*adminCfg, jwtSvc, userService, adminLogService, userActionLogService, fileService)

	// 验证文件存储配置
	if err := fileStorageCfg.ValidateConfigs(); err != nil {
		log.Fatalf("文件存储配置验证失败: %v", err)
	}

	// 设置路由
	r := router.SetupRoutes(userHandler, fileHandler, adminHandler, jwtSvc, accessTokenBlacklistRepo)

	// 启动管理面板服务器
	go startPanelServer()

	// 启动API服务器
	port := getEnv("PORT", "8080")
	log.Printf("API 服务器启动在端口 %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("API 服务器启动失败: %v", err)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 