package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimitRepository 频率限制仓库接口
type RateLimitRepository interface {
	// Increment a counter for a given IP address.
	// Returns the new value of the counter.
	Increment(ctx context.Context, ip string) (int64, error)
}

// redisRateLimitRepository Redis频率限制仓库实现
type redisRateLimitRepository struct {
	rdb *redis.Client
}

// NewRateLimitRepository 创建频率限制仓库实例
func NewRateLimitRepository(rdb *redis.Client) RateLimitRepository {
	return &redisRateLimitRepository{rdb: rdb}
}

// Increment increments the request count for a given IP.
// It uses a Redis pipeline to atomically increment the counter
// and set its expiry to 24 hours on the first increment.
func (r *redisRateLimitRepository) Increment(ctx context.Context, ip string) (int64, error) {
	key := r.getRedisKey(ip)
	
	var count int64
	pipe := r.rdb.Pipeline()
	
	// Increment the counter for the IP
	result := pipe.Incr(ctx, key)
	
	// Set the expiration only on the first request of the day
	pipe.ExpireNX(ctx, key, 24*time.Hour)

	// Execute the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("无法执行Redis pipeline进行频率计数: %w", err)
	}

	// Get the result of the INCR command
	count, err = result.Result()
	if err != nil {
		return 0, fmt.Errorf("无法获取INCR命令结果: %w", err)
	}

	return count, nil
}

// getRedisKey generates the Redis key for IP rate limiting.
func (r *redisRateLimitRepository) getRedisKey(ip string) string {
	return fmt.Sprintf("rate_limit:ip:%s", ip)
} 