package shortid

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// RedisSequenceProvider Redis 实现（生产用）
// ============================================================================

// RedisSequenceProvider Redis 实现的 SequenceProvider
// 用于生产环境，支持分布式部署
// 使用 Redis INCR 命令实现原子递增，保证并发安全
type RedisSequenceProvider struct {
	client *redis.Client
}

// NewRedisSequenceProvider 创建 Redis 序列号提供者
//
// 参数：
//   - addr: Redis 服务器地址，格式：host:port，例如 "localhost:6379"
//
// 返回：
//   - *RedisSequenceProvider: Redis 序列号提供者实例
//   - error: 如果创建失败，返回错误
//
// 示例：
//
//	provider, err := NewRedisSequenceProvider("localhost:6379")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Close()
func NewRedisSequenceProvider(addr string) (*RedisSequenceProvider, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisSequenceProvider{
		client: client,
	}, nil
}

// NewRedisSequenceProviderWithOptions 使用自定义选项创建 Redis 序列号提供者
//
// 参数：
//   - opts: Redis 客户端选项
//
// 返回：
//   - *RedisSequenceProvider: Redis 序列号提供者实例
//   - error: 如果创建失败，返回错误
func NewRedisSequenceProviderWithOptions(opts *redis.Options) (*RedisSequenceProvider, error) {
	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisSequenceProvider{
		client: client,
	}, nil
}

// GetSequence 获取序列号（原子递增，取模128）
//
// 使用 Redis INCR 命令实现原子递增，然后取模128确保序列号在有效范围内（0-127）
//
// 参数：
//   - ctx: 上下文
//   - key: 序列号键名，通常基于时间戳（10ms单位）生成唯一键
//
// 返回：
//   - uint16: 序列号（0-127）
//   - error: 如果操作失败，返回错误
func (r *RedisSequenceProvider) GetSequence(ctx context.Context, key string) (uint16, error) {
	// 构建完整的Redis键名
	redisKey := fmt.Sprintf("shortid:sequence:%s", key)

	// 使用 Redis INCR 原子递增
	val, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment sequence counter: %w", err)
	}

	// 取模128确保序列号在有效范围内（0-127）
	return uint16(val % 128), nil
}

// SetSequenceExpiration 设置序列号键的过期时间
//
// 用于清理过期的序列号键，避免内存/存储泄漏
// 过期时间建议设置为时间单位的2-3倍（例如：30ms）
//
// 参数：
//   - ctx: 上下文
//   - key: 序列号键名
//   - expiration: 过期时间
//
// 返回：
//   - error: 如果操作失败，返回错误
func (r *RedisSequenceProvider) SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error {
	redisKey := fmt.Sprintf("shortid:sequence:%s", key)
	return r.client.Expire(ctx, redisKey, expiration).Err()
}

// HealthCheck 健康检查
//
// 通过 PING 命令检查 Redis 连接是否可用
//
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - error: 如果连接不可用，返回错误
func (r *RedisSequenceProvider) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接，释放资源
//
// 关闭 Redis 客户端连接，释放相关资源
//
// 返回：
//   - error: 如果关闭失败，返回错误
func (r *RedisSequenceProvider) Close() error {
	return r.client.Close()
}

