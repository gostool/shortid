package shortid

import (
	"context"
	"sync"
	"time"
)

// ============================================================================
// MemorySequenceProvider 内存实现（测试用）
// ============================================================================

// MemorySequenceProvider 内存实现的 SequenceProvider
// 用于测试环境，无需外部依赖（如 Redis）
// 注意：内存实现不支持真正的分布式，仅用于单机测试
type MemorySequenceProvider struct {
	sequences map[string]int64
	mu        sync.RWMutex
}

// NewMemorySequenceProvider 创建内存序列号提供者
func NewMemorySequenceProvider() *MemorySequenceProvider {
	return &MemorySequenceProvider{
		sequences: make(map[string]int64),
	}
}

// GetSequence 获取序列号（原子递增，取模128）
func (m *MemorySequenceProvider) GetSequence(ctx context.Context, key string) (uint16, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 获取当前序列号并递增
	m.sequences[key]++
	sequence := m.sequences[key]

	// 取模128确保序列号在有效范围内（0-127）
	return uint16(sequence % 128), nil
}

// SetSequenceExpiration 设置序列号键的过期时间
// 内存实现可以定期清理，但这里简化处理，不实现过期机制
func (m *MemorySequenceProvider) SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error {
	// 内存实现可以忽略过期时间，或者实现定期清理逻辑
	// 这里简化处理，直接返回 nil
	return nil
}

// HealthCheck 健康检查
// 内存实现始终健康，直接返回 nil
func (m *MemorySequenceProvider) HealthCheck(ctx context.Context) error {
	return nil
}

// Close 关闭连接
// 内存实现清理资源
func (m *MemorySequenceProvider) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sequences = make(map[string]int64)
	return nil
}

