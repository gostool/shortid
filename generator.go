package shortid

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ============================================================================
// Generator 核心实现
// ============================================================================

// Generator ID生成器
// 基于 Sonyflake 算法实现分布式唯一ID生成器
// 支持 Serverless 部署，结合时间戳压缩和 Base62 编码生成短ID
type Generator struct {
	mu           sync.Mutex
	startTime    int64  // 基准时间（10ms单位）
	elapsedTime  int64  // 已过时间（10ms单位）
	sequence     uint16 // 序列号
	machineID    uint16 // 机器ID
	machineReady bool   // machineID是否已初始化
	businessType uint8  // 业务类型

	// Serverless模式
	machineIDProvider  MachineIDProvider
	useMachineProvider bool
	machineLease       *MachineIDLease

	// 分布式序列号模式
	sequenceProvider    SequenceProvider
	useSequenceProvider bool

	// 租约模式（推荐）
	machineIDLeaseProvider  MachineIDLeaseProvider
	useMachineLeaseProvider bool
	leaseDuration           time.Duration
	leaseRenewAt            time.Time

	// 输出格式
	returnRawID bool // true: 返回原始数字ID, false: 返回短ID
}

// Config 生成器配置
type Config struct {
	// MachineID 固定机器ID（传统部署）
	// 与 MachineIDProvider 二选一
	MachineID uint16

	// BusinessType 业务类型
	BusinessType BusinessType

	// MachineIDProvider Serverless模式机器ID提供者
	// 与 MachineID 二选一
	MachineIDProvider MachineIDProvider

	// MachineIDLeaseProvider 机器ID租约提供者（推荐）
	// 与 MachineID / MachineIDProvider 二选一
	MachineIDLeaseProvider MachineIDLeaseProvider

	// SequenceProvider 分布式序列号提供者
	// 可选，如果不设置则使用本地序列号
	SequenceProvider SequenceProvider

	// StartTime 基准时间（可选，默认2024-01-01）
	StartTime time.Time

	// ReturnRawID 是否返回原始数字ID（uint64）
	// false: 返回短ID（Base62编码，默认）
	// true: 返回原始数字ID（uint64，10进制）
	ReturnRawID bool

	// MachineIDLeaseDuration 机器ID租约时长（默认20分钟）
	MachineIDLeaseDuration time.Duration
}

// ValidateConfig 验证生成器配置是否合法。
//
// 兼容策略：
//   - 该函数只做参数与组合校验，不做外部依赖探测。
//   - NewGenerator 会调用该函数，建议调用方在启动阶段显式调用以提前失败。
func ValidateConfig(config Config) error {
	if config.BusinessType > BusinessReservedZ {
		return ErrInvalidBusinessType
	}
	if config.MachineID > SnowflakeMaxMachine {
		return ErrInvalidMachineID
	}
	if config.MachineIDProvider != nil && config.MachineID != 0 {
		return fmt.Errorf("%w: MachineID and MachineIDProvider are mutually exclusive", ErrInvalidConfig)
	}
	if config.MachineIDLeaseProvider != nil && config.MachineIDProvider != nil {
		return fmt.Errorf("%w: MachineIDProvider and MachineIDLeaseProvider are mutually exclusive", ErrInvalidConfig)
	}
	if config.MachineIDLeaseProvider != nil && config.MachineID != 0 {
		return fmt.Errorf("%w: MachineID and MachineIDLeaseProvider are mutually exclusive", ErrInvalidConfig)
	}
	if config.MachineIDLeaseDuration < 0 {
		return fmt.Errorf("%w: MachineIDLeaseDuration must be >= 0", ErrInvalidConfig)
	}
	return nil
}

// New 是面向常见场景的便捷构造函数。
//
// 适用：固定机器ID部署（单机/固定节点）。
func New(machineID uint16, businessType BusinessType) (*Generator, error) {
	return NewGenerator(Config{
		MachineID:    machineID,
		BusinessType: businessType,
	})
}

// MustNew 是便捷构造函数的 panic 版本。
//
// 适合在程序启动阶段初始化全局实例时使用。
func MustNew(machineID uint16, businessType BusinessType) *Generator {
	g, err := New(machineID, businessType)
	if err != nil {
		panic(err)
	}
	return g
}

// NewGenerator 创建ID生成器
//
// 参数：
//   - config: 生成器配置
//
// 返回：
//   - *Generator: ID生成器实例
//   - error: 如果配置无效，返回错误
func NewGenerator(config Config) (*Generator, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	g := &Generator{
		businessType:  uint8(config.BusinessType),
		sequence:      SnowflakeMaxSequence, // 初始化为最大值，首次生成时重置为0
		returnRawID:   config.ReturnRawID,   // 设置输出格式
		leaseDuration: 20 * time.Minute,
	}
	if config.MachineIDLeaseDuration > 0 {
		g.leaseDuration = config.MachineIDLeaseDuration
	}

	// 设置基准时间
	if config.StartTime.IsZero() {
		g.startTime = DefaultSnowflakeEpochMs / 10 // 转换为10ms单位
	} else {
		g.startTime = config.StartTime.UnixMilli() / 10
	}

	// 机器ID分配（优先租约模式）
	if config.MachineIDLeaseProvider != nil {
		g.machineIDLeaseProvider = config.MachineIDLeaseProvider
		g.useMachineLeaseProvider = true
		g.machineReady = false
	} else if config.MachineIDProvider != nil {
		g.machineIDProvider = config.MachineIDProvider
		g.useMachineProvider = true
		g.machineReady = false // 延迟获取机器ID（首次生成时获取）
	} else {
		g.machineID = config.MachineID
		g.machineReady = true
		g.useMachineProvider = false
	}

	// 序列号提供者
	if config.SequenceProvider != nil {
		g.sequenceProvider = config.SequenceProvider
		g.useSequenceProvider = true
	} else {
		g.useSequenceProvider = false
	}

	return g, nil
}

// Generate 生成ID（固定机器ID模式）
//
// 返回：
//   - string: ID字符串（短ID或数字ID字符串，取决于配置）
//   - error: 如果生成失败，返回错误
func (g *Generator) Generate() (string, error) {
	return g.GenerateWithContext(context.Background())
}

// GenerateWithContext 生成ID（支持Serverless模式）
//
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - string: ID字符串（短ID或数字ID字符串）
//   - error: 如果生成失败，返回错误
func (g *Generator) GenerateWithContext(ctx context.Context) (string, error) {
	// 生成64位数字ID
	id, err := g.nextID(ctx)
	if err != nil {
		return "", err
	}

	// 根据配置返回短ID或原始数字ID
	if g.returnRawID {
		return fmt.Sprintf("%d", id), nil
	}

	// 转换为短ID
	return g.toShortID(id)
}

// NextID 生成原始数字ID（uint64）
// 返回64位数字ID，不进行Base62编码
//
// 参数：
//   - ctx: 上下文
//
// 返回：
//   - uint64: 64位数字ID
//   - error: 如果生成失败，返回错误
func (g *Generator) NextID(ctx context.Context) (uint64, error) {
	// 生成64位数字ID
	return g.nextID(ctx)
}

// nextID 生成64位数字ID（基于Sonyflake算法）
func (g *Generator) nextID(ctx context.Context) (uint64, error) {
	const maskSequence = uint16(1<<SnowflakeSequenceBits - 1)
	const timeUnit = 10 // 10ms单位（Sonyflake使用10ms）

	g.mu.Lock()
	defer g.mu.Unlock()

	nowTime := time.Now()
	if err := g.ensureMachineIdentity(ctx, nowTime); err != nil {
		return 0, err
	}

	// 计算当前已过时间（10ms单位）
	now := nowTime.UnixMilli() / timeUnit
	current := now - g.startTime

	if g.elapsedTime < current {
		g.elapsedTime = current
		g.sequence = 0
	} else {
		// 使用分布式序列号或本地序列号
		if g.useSequenceProvider {
			// 分布式序列号模式
			sequenceKey := fmt.Sprintf("%d", g.elapsedTime)
			sequence, err := g.sequenceProvider.GetSequence(ctx, sequenceKey)
			if err != nil {
				return 0, fmt.Errorf("failed to get sequence: %w", err)
			}
			g.sequence = sequence
			// 设置序列号键的过期时间（1s，Redis最小支持值）
			// 注意：Redis不支持小于1s的过期时间，所以使用1s而不是30ms
			_ = g.sequenceProvider.SetSequenceExpiration(ctx, sequenceKey, 1*time.Second)
		} else {
			// 本地序列号模式
			g.sequence = (g.sequence + 1) & maskSequence
		}

		if g.sequence == 0 {
			// 序列号溢出，等待下一时间单位
			g.elapsedTime++
			overtime := g.elapsedTime - current
			if overtime > 0 {
				time.Sleep(time.Duration(overtime*timeUnit) * time.Millisecond)
			}
		}
	}

	// 检查时间溢出
	if g.elapsedTime >= 1<<SnowflakeTimestampBits {
		return 0, ErrOverTimeLimit
	}

	// 组装64位ID
	return uint64(g.elapsedTime)<<(SnowflakeBusinessShift+SnowflakeMachineShift+SnowflakeSequenceBits) |
		uint64(g.businessType)<<(SnowflakeMachineShift+SnowflakeSequenceBits) |
		uint64(g.machineID)<<SnowflakeSequenceBits |
		uint64(g.sequence), nil
}

func nextLeaseRenewTime(now time.Time, leaseDuration time.Duration) time.Time {
	if leaseDuration <= 0 {
		return now
	}
	return now.Add(leaseDuration / 2)
}

// toShortID 转换为短ID
func (g *Generator) toShortID(id uint64) (string, error) {
	// 提取各部分
	elapsedTime := id >> (SnowflakeBusinessShift + SnowflakeMachineShift + SnowflakeSequenceBits)
	businessType := uint8((id >> (SnowflakeMachineShift + SnowflakeSequenceBits)) & SnowflakeMaxBusiness)
	machineID := uint16((id >> SnowflakeSequenceBits) & SnowflakeMaxMachine)
	sequence := uint16(id & SnowflakeMaxSequence)

	// 计算实际时间戳（毫秒）
	timestampMs := (g.startTime*10 + int64(elapsedTime)*10)

	// 时间戳压缩（毫秒级）
	compressedTime := ToTimestampShortMs(timestampMs)

	// Base62编码各部分
	businessStr := EncodeBase62(uint64(businessType))
	machineStr := EncodeBase62(uint64(machineID))
	sequenceStr := EncodeBase62(uint64(sequence))

	// 拼接
	return businessStr + compressedTime + machineStr + sequenceStr, nil
}
