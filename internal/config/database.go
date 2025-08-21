package config

import (
	"backend/internal/model"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库配置
type Database struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig 获取数据库配置
func GetDatabaseConfig() *Database {
	return &Database{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "postgres"),
		DBName:   getEnv("DB_NAME", "backend"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// InitDatabase 初始化数据库连接
func InitDatabase() (*gorm.DB, error) {
	config := GetDatabaseConfig()
	
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		config.Host, config.User, config.Password, config.DBName, config.Port, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	// 自动迁移模型
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("数据库连接成功")
	return db, nil
}

// AutoMigrate 自动迁移数据库模型
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.File{}, // 添加文件模型
		&model.UserDevice{},
		&model.DeviceVerification{},
		&model.AdminActionLog{},
		&model.UserActionLog{},
		// 在此处添加其他模型
		&model.FriendRequest{},
		&model.Friendship{},
		&model.BlockList{},
		&model.FriendBan{},
		&model.ChatRoom{},
	)
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
} 