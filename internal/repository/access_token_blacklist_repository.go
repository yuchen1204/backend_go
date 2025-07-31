package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
)

// AccessTokenBlacklistRepository Access Token黑名单仓储接口
type AccessTokenBlacklistRepository interface {
	// Add adds an access token to the blacklist
	Add(ctx context.Context, userID uuid.UUID, accessToken string, expiration time.Duration) error
	// IsBlacklisted checks if an access token is blacklisted
	IsBlacklisted(ctx context.Context, accessToken string) (bool, error)
	// RemoveExpiredTokens removes expired tokens from blacklist (cleanup method)
	RemoveExpiredTokens(ctx context.Context) error
}

// redisAccessTokenBlacklistRepository Redis Access Token黑名单仓储实现
type redisAccessTokenBlacklistRepository struct {
	rdb *redis.Client
}

// NewAccessTokenBlacklistRepository 创建 Access Token 黑名单仓储实例
func NewAccessTokenBlacklistRepository(rdb *redis.Client) AccessTokenBlacklistRepository {
	return &redisAccessTokenBlacklistRepository{rdb: rdb}
}

// Add adds an access token to the blacklist
func (r *redisAccessTokenBlacklistRepository) Add(ctx context.Context, userID uuid.UUID, accessToken string, expiration time.Duration) error {
	key := r.getRedisKey(accessToken)
	
	// Store the token with the remaining TTL until its natural expiration
	err := r.rdb.Set(ctx, key, userID.String(), expiration).Err()
	if err != nil {
		return fmt.Errorf("无法将access token加入黑名单: %w", err)
	}

	return nil
}

// IsBlacklisted checks if an access token is blacklisted
func (r *redisAccessTokenBlacklistRepository) IsBlacklisted(ctx context.Context, accessToken string) (bool, error) {
	key := r.getRedisKey(accessToken)
	
	_, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // Token not in blacklist
	}
	if err != nil {
		return false, fmt.Errorf("无法检查access token黑名单状态: %w", err)
	}

	return true, nil // Token is blacklisted
}

// RemoveExpiredTokens removes expired tokens from blacklist
// This is handled automatically by Redis TTL, but we provide this method for manual cleanup if needed
func (r *redisAccessTokenBlacklistRepository) RemoveExpiredTokens(ctx context.Context) error {
	// Redis automatically removes expired keys, so this is mainly for monitoring/logging purposes
	return nil
}

// getRedisKey generates the Redis key for blacklisted access tokens
func (r *redisAccessTokenBlacklistRepository) getRedisKey(accessToken string) string {
	return fmt.Sprintf("blacklist:access_token:%s", accessToken)
} 