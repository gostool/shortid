package shortid

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// MemoryMachineIDProvider 内存实现（测试用）
// ============================================================================

// MemoryMachineIDProvider 内存实现的 MachineIDProvider
// 用于测试环境，无需外部依赖（如 Redis）
// 注意：内存实现不支持真正的分布式，仅用于单机测试
type MemoryMachineIDProvider struct {
	counter int64
	mu      sync.Mutex
}

// NewMemoryMachineIDProvider 创建内存机器ID提供者
func NewMemoryMachineIDProvider() *MemoryMachineIDProvider {
	return &MemoryMachineIDProvider{}
}

// GetMachineID 获取机器ID（原子递增，取模64）
func (m *MemoryMachineIDProvider) GetMachineID(ctx context.Context) (uint16, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.counter++
	return uint16(m.counter % 64), nil
}

// SetMachineIDExpiration 设置机器ID过期时间
// 内存实现无需过期机制，直接返回 nil
func (m *MemoryMachineIDProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	// 内存实现无需过期机制
	return nil
}

// HealthCheck 健康检查
// 内存实现始终健康，直接返回 nil
func (m *MemoryMachineIDProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭连接
// 内存实现无需关闭操作，直接返回 nil
func (m *MemoryMachineIDProvider) Close() error {
	return nil
}

