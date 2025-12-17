package shortid

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// StateProvider 接口定义
// ============================================================================

// StateProvider 分布式状态提供者接口
// 定义了ID生成器所需的分布式状态操作能力
// 支持多种实现：Redis（生产环境）、内存（测试环境）
type StateProvider interface {
	// GetMachineID 获取机器ID（原子递增，取模64）
	// 在 Serverless 环境中，每次函数启动时动态分配机器ID
	GetMachineID(ctx context.Context) (uint16, error)

	// SetMachineIDExpiration 设置机器ID过期时间
	// 用于 Serverless 环境，设置机器ID的过期时间，支持自动回收
	SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error

	// HealthCheck 健康检查
	// 用于验证连接是否可用，在ID生成前检查存储系统是否正常
	HealthCheck(ctx context.Context) error

	// Close 关闭连接，释放资源
	// 在程序退出或资源释放时调用
	Close() error
}

// ============================================================================
// MemoryProvider 内存实现（测试用）
// ============================================================================

// MemoryProvider 内存实现的 StateProvider
// 用于测试环境，无需外部依赖（如 Redis）
// 注意：内存实现不支持真正的分布式，仅用于单机测试
type MemoryProvider struct {
	counter int64
	mu      sync.Mutex
}

// NewMemoryProvider 创建内存 Provider
func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{}
}

// GetMachineID 获取机器ID（原子递增，取模64）
func (m *MemoryProvider) GetMachineID(ctx context.Context) (uint16, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter++
	return uint16(m.counter % 64), nil
}

// SetMachineIDExpiration 设置机器ID过期时间
// 内存实现无需过期机制，直接返回 nil
func (m *MemoryProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	// 内存实现无需过期机制
	return nil
}

// HealthCheck 健康检查
// 内存实现始终健康，直接返回 nil
func (m *MemoryProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭连接
// 内存实现无需关闭操作，直接返回 nil
func (m *MemoryProvider) Close() error {
	return nil
}

