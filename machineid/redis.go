package shortid

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ============================================================================
// RedisMachineIDProvider Redis 实现（生产用）
// ============================================================================

// RedisMachineIDProvider Redis 实现的 MachineIDProvider
// 用于生产环境，支持分布式部署
// 使用 Redis INCR 命令实现原子递增，保证并发安全
type RedisMachineIDProvider struct {
	client *redis.Client
}

// NewRedisMachineIDProvider 创建 Redis 机器ID提供者
//
// 参数：
//   - addr: Redis 服务器地址，格式：host:port，例如 "localhost:6379"
//
// 返回：
//   - *RedisMachineIDProvider: Redis 机器ID提供者实例
//   - error: 如果创建失败，返回错误
//
// 示例：
//
//	provider, err := NewRedisMachineIDProvider("localhost:6379")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Close()
func NewRedisMachineIDProvider(addr string) (*RedisMachineIDProvider, error) {
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

	return &RedisMachineIDProvider{
		client: client,
	}, nil
}

// NewRedisMachineIDProviderWithOptions 使用自定义选项创建 Redis 机器ID提供者
//
// 参数：
//   - opts: Redis 客户端选项
//
// 返回：
//   - *RedisMachineIDProvider: Redis 机器ID提供者实例
//   - error: 如果创建失败，返回错误
func NewRedisMachineIDProviderWithOptions(opts *redis.Options) (*RedisMachineIDProvider, error) {
	client := redis.NewClient(opts)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisMachineIDProvider{
		client: client,
	}, nil
}

// GetMachineID 获取机器ID（原子递增，取模64）
//
// 使用 Redis INCR 命令实现原子递增，然后取模64确保机器ID在有效范围内（0-63）
//
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - uint16: 机器ID（0-63）
//   - error: 如果操作失败，返回错误
func (r *RedisMachineIDProvider) GetMachineID(ctx context.Context) (uint16, error) {
	// 使用 Redis INCR 原子递增
	val, err := r.client.Incr(ctx, "shortid:machine:id").Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment machine id counter: %w", err)
	}

	// 取模64确保机器ID在有效范围内（0-63）
	return uint16(val % 64), nil
}

// SetMachineIDExpiration 设置机器ID过期时间
//
// 用于 Serverless 环境，设置机器ID的过期时间，支持自动回收
// 过期时间建议设置为函数最大运行时间 + 缓冲时间（例如：15分钟 + 5分钟 = 20分钟）
//
// 参数：
//   - ctx: 上下文
//   - machineID: 机器ID
//   - expiration: 过期时间
//
// 返回：
//   - error: 如果操作失败，返回错误
func (r *RedisMachineIDProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	key := fmt.Sprintf("shortid:machine:id:%d", machineID)
	return r.client.Expire(ctx, key, expiration).Err()
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
func (r *RedisMachineIDProvider) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close 关闭连接，释放资源
//
// 关闭 Redis 客户端连接，释放相关资源
//
// 返回：
//   - error: 如果关闭失败，返回错误
func (r *RedisMachineIDProvider) Close() error {
	return r.client.Close()
}

