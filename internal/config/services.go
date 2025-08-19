package config

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// SMTPConfig SMTP配置
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	Sender   string
}

// AdminConfig 存储管理员面板的认证信息
type AdminConfig struct {
	User     string
	Password string
}

// SecurityConfig 安全相关配置
type SecurityConfig struct {
	MaxRequestsPerIPPerDay         int
	JwtSecret                      string
	JwtAccessTokenExpiresInMinutes int
	JwtRefreshTokenExpiresInDays   int
}

// GetRedisConfig 获取Redis配置
func GetRedisConfig() *RedisConfig {
	db, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
	return &RedisConfig{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

// GetSMTPConfig 获取SMTP配置
// GetAdminConfig 获取管理员配置
func GetAdminConfig() *AdminConfig {
	return &AdminConfig{
		User:     getEnv("PANEL_USER", "admin"),
		Password: getEnv("PANEL_PASSWORD", "password"),
	}
}

// GetSMTPConfig 获取SMTP配置
func GetSMTPConfig() *SMTPConfig {
	port, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	return &SMTPConfig{
		Host:     getEnv("SMTP_HOST", "smtp.example.com"),
		Port:     port,
		Username: getEnv("SMTP_USERNAME", "user@example.com"),
		Password: getEnv("SMTP_PASSWORD", "password"),
		From:     getEnv("SMTP_FROM", getEnv("SMTP_FROM", "user@example.com")),
		Sender:   getEnv("SMTP_SENDER", "Sender"),
	}
}

// GetSecurityConfig 获取安全配置
func GetSecurityConfig() *SecurityConfig {
	maxRequests, _ := strconv.Atoi(getEnv("MAX_IP_REQUESTS_PER_DAY", "10"))
	accessTokenExpires, _ := strconv.Atoi(getEnv("JWT_ACCESS_TOKEN_EXPIRES_IN_MINUTES", "30"))
	refreshTokenExpires, _ := strconv.Atoi(getEnv("JWT_REFRESH_TOKEN_EXPIRES_IN_DAYS", "7"))
	return &SecurityConfig{
		MaxRequestsPerIPPerDay:         maxRequests,
		JwtSecret:                      getEnv("JWT_SECRET", "a-very-secret-key-that-should-be-changed"),
		JwtAccessTokenExpiresInMinutes: accessTokenExpires,
		JwtRefreshTokenExpiresInDays:   refreshTokenExpires,
	}
}

// InitRedis 初始化Redis连接
func InitRedis() (*redis.Client, error) {
	config := GetRedisConfig()
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("无法连接到Redis: %w", err)
	}

	log.Println("Redis连接成功")
	return rdb, nil
}
