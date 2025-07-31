package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// CodeRepository 验证码缓存接口
type CodeRepository interface {
	Set(ctx context.Context, email, code string, expiration time.Duration) error
	Get(ctx context.Context, email string) (string, error)
	Delete(ctx context.Context, email string) error
}

// redisCodeRepository Redis验证码缓存实现
type redisCodeRepository struct {
	rdb *redis.Client
}

// NewCodeRepository 创建验证码缓存实例
func NewCodeRepository(rdb *redis.Client) CodeRepository {
	return &redisCodeRepository{rdb: rdb}
}

// Set 将验证码存入Redis
func (r *redisCodeRepository) Set(ctx context.Context, email, code string, expiration time.Duration) error {
	key := r.getRedisKey(email)
	err := r.rdb.Set(ctx, key, code, expiration).Err()
	if err != nil {
		return fmt.Errorf("无法将验证码存入Redis: %w", err)
	}
	return nil
}

// Get 从Redis获取验证码
func (r *redisCodeRepository) Get(ctx context.Context, email string) (string, error) {
	key := r.getRedisKey(email)
	code, err := r.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // 验证码不存在或已过期
	}
	if err != nil {
		return "", fmt.Errorf("无法从Redis获取验证码: %w", err)
	}
	return code, nil
}

// Delete 从Redis删除验证码
func (r *redisCodeRepository) Delete(ctx context.Context, email string) error {
	key := r.getRedisKey(email)
	err := r.rdb.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("无法从Redis删除验证码: %w", err)
	}
	return nil
}

// getRedisKey 生成验证码在Redis中的键
func (r *redisCodeRepository) getRedisKey(email string) string {
	return fmt.Sprintf("verification_code:%s", email)
} 