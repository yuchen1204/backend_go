package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// RefreshTokenRepository 刷新Token仓储接口
type RefreshTokenRepository interface {
	// Store stores a refresh token for a user
	Store(ctx context.Context, userID uuid.UUID, refreshToken string, expiration time.Duration) error
	// Validate checks if a refresh token is valid for a user
	Validate(ctx context.Context, userID uuid.UUID, refreshToken string) (bool, error)
	// Delete removes a refresh token for a user
	Delete(ctx context.Context, userID uuid.UUID) error
	// DeleteByToken removes a specific refresh token
	DeleteByToken(ctx context.Context, refreshToken string) error
}

// redisRefreshTokenRepository Redis刷新Token仓储实现
type redisRefreshTokenRepository struct {
	rdb *redis.Client
}

// NewRefreshTokenRepository 创建刷新Token仓储实例
func NewRefreshTokenRepository(rdb *redis.Client) RefreshTokenRepository {
	return &redisRefreshTokenRepository{rdb: rdb}
}

// Store stores a refresh token for a user
func (r *redisRefreshTokenRepository) Store(ctx context.Context, userID uuid.UUID, refreshToken string, expiration time.Duration) error {
	key := r.getRedisKey(userID)
	err := r.rdb.Set(ctx, key, refreshToken, expiration).Err()
	if err != nil {
		return fmt.Errorf("无法存储refresh token: %w", err)
	}

	// Also store a mapping from token to userID for validation
	tokenKey := r.getTokenKey(refreshToken)
	err = r.rdb.Set(ctx, tokenKey, userID.String(), expiration).Err()
	if err != nil {
		return fmt.Errorf("无法存储token映射: %w", err)
	}

	return nil
}

// Validate checks if a refresh token is valid for a user
func (r *redisRefreshTokenRepository) Validate(ctx context.Context, userID uuid.UUID, refreshToken string) (bool, error) {
	key := r.getRedisKey(userID)
	storedToken, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Token not found
	}
	if err != nil {
		return false, fmt.Errorf("无法验证refresh token: %w", err)
	}

	return storedToken == refreshToken, nil
}

// Delete removes a refresh token for a user
func (r *redisRefreshTokenRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	// First get the token to remove the token mapping
	key := r.getRedisKey(userID)
	refreshToken, err := r.rdb.Get(ctx, key).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("无法获取refresh token进行删除: %w", err)
	}

	// Delete the user -> token mapping
	err = r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("无法删除用户refresh token: %w", err)
	}

	// Delete the token -> user mapping if token exists
	if refreshToken != "" {
		tokenKey := r.getTokenKey(refreshToken)
		_ = r.rdb.Del(ctx, tokenKey).Err() // Ignore error for cleanup
	}

	return nil
}

// DeleteByToken removes a specific refresh token
func (r *redisRefreshTokenRepository) DeleteByToken(ctx context.Context, refreshToken string) error {
	// Get the userID from the token mapping
	tokenKey := r.getTokenKey(refreshToken)
	userIDStr, err := r.rdb.Get(ctx, tokenKey).Result()
	if err == redis.Nil {
		return nil // Token not found, nothing to delete
	}
	if err != nil {
		return fmt.Errorf("无法获取token对应的用户ID: %w", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fmt.Errorf("无效的用户ID格式: %w", err)
	}

	// Delete both mappings
	userKey := r.getRedisKey(userID)
	pipe := r.rdb.Pipeline()
	pipe.Del(ctx, userKey)
	pipe.Del(ctx, tokenKey)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("无法删除refresh token映射: %w", err)
	}

	return nil
}

// getRedisKey generates the Redis key for user refresh tokens
func (r *redisRefreshTokenRepository) getRedisKey(userID uuid.UUID) string {
	return fmt.Sprintf("refresh_token:user:%s", userID.String())
}

// getTokenKey generates the Redis key for token to user mapping
func (r *redisRefreshTokenRepository) getTokenKey(token string) string {
	return fmt.Sprintf("refresh_token:token:%s", token)
} 