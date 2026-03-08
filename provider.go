package shortid

import (
	"context"
	"time"
)

// MachineIDLease 表示机器ID租约。
//
// Token 为租约唯一令牌，续租/释放时必须携带，防止误操作其他实例持有的租约。
type MachineIDLease struct {
	MachineID uint16
	Token     string
	ExpiresAt time.Time
}

// MachineIDLeaseProvider 机器ID租约提供者接口。
//
// 该接口用于分布式/Serverless场景，通过“租约”模式分配机器ID，
// 支持 Redis、etcd 或其他第三方组件实现。
type MachineIDLeaseProvider interface {
	// AcquireMachineIDLease 申请机器ID租约。
	AcquireMachineIDLease(ctx context.Context, ttl time.Duration) (*MachineIDLease, error)

	// RenewMachineIDLease 续租。
	// 返回值 ok=false 表示租约已失效/不属于当前持有者。
	RenewMachineIDLease(ctx context.Context, lease *MachineIDLease, ttl time.Duration) (ok bool, err error)

	// ReleaseMachineIDLease 主动释放租约。
	ReleaseMachineIDLease(ctx context.Context, lease *MachineIDLease) error

	// HealthCheck 健康检查。
	HealthCheck(ctx context.Context) error

	// Close 关闭连接，释放资源。
	Close() error
}

// ============================================================================
// MachineIDProvider 接口定义
// ============================================================================

// MachineIDProvider 机器ID提供者接口。
//
// 契约约束（稳定 API）：
//   - 实现方负责原子分配 machine id，并保证返回值在 [0, 63]。
//   - 实现方不应在内部做无限重试；超时、重试策略由调用方通过 ctx 控制。
//   - 返回错误应尽量保留底层依赖错误（便于调用方分类处理）。
//
// 用于 Serverless 环境动态分配机器ID，支持多种实现：Redis（生产环境）、内存（测试环境）。
type MachineIDProvider interface {
	// GetMachineID 获取机器ID（原子递增，取模64）
	// 要求：必须保证原子性，返回 0-63 范围内的机器ID
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回：
	//   - uint16: 机器ID（0-63）
	//   - error: 如果操作失败，返回错误
	GetMachineID(ctx context.Context) (uint16, error)

	// SetMachineIDExpiration 设置机器ID过期时间
	// 要求：用于 Serverless 环境，设置机器ID的过期时间，支持自动回收
	//
	// 参数：
	//   - ctx: 上下文
	//   - machineID: 机器ID
	//   - expiration: 过期时间，建议设置为函数最大运行时间 + 缓冲时间（例如：20分钟）
	//
	// 返回：
	//   - error: 如果操作失败，返回错误
	SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error

	// HealthCheck 健康检查
	// 要求：验证连接是否可用，在ID生成前检查存储系统是否正常
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回：
	//   - error: 如果连接不可用，返回错误
	HealthCheck(ctx context.Context) error

	// Close 关闭连接，释放资源
	// 要求：在程序退出或资源释放时调用，释放相关资源
	//
	// 返回：
	//   - error: 如果关闭失败，返回错误
	Close() error
}

// ============================================================================
// SequenceProvider 接口定义
// ============================================================================

// SequenceProvider 序列号提供者接口。
//
// 契约约束（稳定 API）：
//   - 实现方负责同 key 下的原子序列分配，并保证返回值在 [0, 127]。
//   - 实现方不应在内部做无限重试；超时、重试策略由调用方通过 ctx 控制。
//   - 返回错误应尽量保留底层依赖错误（便于调用方分类处理）。
//
// 用于分布式环境生成序列号，支持多种实现：Redis（生产环境）、内存（测试环境）。
type SequenceProvider interface {
	// GetSequence 获取序列号（原子递增，取模128）
	// 要求：必须保证原子性，返回 0-127 范围内的序列号
	//
	// 参数：
	//   - ctx: 上下文
	//   - key: 序列号键名，通常基于时间戳（10ms单位）生成唯一键
	//
	// 返回：
	//   - uint16: 序列号（0-127）
	//   - error: 如果操作失败，返回错误
	//
	// 说明：
	//   - 序列号在同一时间单位（10ms）内递增
	//   - 不同时间单位使用不同的 key，序列号从0开始
	//   - 如果序列号溢出（达到128），调用方需要等待下一时间单位
	GetSequence(ctx context.Context, key string) (uint16, error)

	// SetSequenceExpiration 设置序列号键的过期时间
	// 要求：用于清理过期的序列号键，避免内存/存储泄漏
	//
	// 参数：
	//   - ctx: 上下文
	//   - key: 序列号键名
	//   - expiration: 过期时间，建议设置为时间单位的2-3倍（例如：1s，Redis最小支持值）
	//
	// 返回：
	//   - error: 如果操作失败，返回错误
	SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error

	// HealthCheck 健康检查
	// 要求：验证连接是否可用，在ID生成前检查存储系统是否正常
	//
	// 参数：
	//   - ctx: 上下文
	//
	// 返回：
	//   - error: 如果连接不可用，返回错误
	HealthCheck(ctx context.Context) error

	// Close 关闭连接，释放资源
	// 要求：在程序退出或资源释放时调用，释放相关资源
	//
	// 返回：
	//   - error: 如果关闭失败，返回错误
	Close() error
}
