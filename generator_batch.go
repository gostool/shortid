package shortid

import (
	"context"
	"fmt"
	"time"
)

// ============================================================================
// Generator 批量生成接口
// ============================================================================

// GenerateBatch 批量生成ID（固定机器ID模式）
//
// 参数：
//   - count: 需要生成的ID数量
//
// 返回：
//   - []string: ID字符串数组（短ID或数字ID字符串，取决于配置）
//   - error: 如果生成失败，返回错误
func (g *Generator) GenerateBatch(count int) ([]string, error) {
	return g.GenerateBatchWithContext(context.Background(), count)
}

// GenerateBatchWithContext 批量生成ID（支持Serverless模式）
//
// 参数：
//   - ctx: 上下文
//   - count: 需要生成的ID数量
//
// 返回：
//   - []string: ID字符串数组（短ID或数字ID字符串，取决于配置）
//   - error: 如果生成失败，返回错误
func (g *Generator) GenerateBatchWithContext(ctx context.Context, count int) ([]string, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}
	if count > MaxBatchCount {
		return nil, fmt.Errorf("count exceeds maximum limit of %d", MaxBatchCount)
	}

	// Serverless模式：首次获取机器ID
	if g.useMachineProvider && g.machineID == 0 {
		machineID, err := g.machineIDProvider.GetMachineID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get machine id: %w", err)
		}
		g.machineID = machineID
		// 设置过期时间（20分钟）
		_ = g.machineIDProvider.SetMachineIDExpiration(ctx, machineID, 20*time.Minute)
	}

	// 预分配切片
	ids := make([]string, 0, count)

	// 批量生成
	for i := 0; i < count; i++ {
		// 生成64位数字ID
		id, err := g.nextID(ctx)
		if err != nil {
			return ids, fmt.Errorf("failed to generate ID at index %d: %w", i, err)
		}

		// 根据配置返回短ID或原始数字ID
		if g.returnRawID {
			ids = append(ids, fmt.Sprintf("%d", id))
		} else {
			// 转换为短ID
			shortID, err := g.toShortID(id)
			if err != nil {
				return ids, fmt.Errorf("failed to convert to short ID at index %d: %w", i, err)
			}
			ids = append(ids, shortID)
		}
	}

	return ids, nil
}

// NextIDBatch 批量生成原始数字ID（uint64）
// 返回64位数字ID数组，不进行Base62编码
//
// 参数：
//   - ctx: 上下文
//   - count: 需要生成的ID数量
//
// 返回：
//   - []uint64: 64位数字ID数组
//   - error: 如果生成失败，返回错误
func (g *Generator) NextIDBatch(ctx context.Context, count int) ([]uint64, error) {
	if count <= 0 {
		return nil, fmt.Errorf("count must be greater than 0")
	}
	if count > MaxBatchCount {
		return nil, fmt.Errorf("count exceeds maximum limit of %d", MaxBatchCount)
	}

	// Serverless模式：首次获取机器ID
	if g.useMachineProvider && g.machineID == 0 {
		machineID, err := g.machineIDProvider.GetMachineID(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get machine id: %w", err)
		}
		g.machineID = machineID
		// 设置过期时间（20分钟）
		_ = g.machineIDProvider.SetMachineIDExpiration(ctx, machineID, 20*time.Minute)
	}

	// 预分配切片
	ids := make([]uint64, 0, count)

	// 批量生成
	for i := 0; i < count; i++ {
		id, err := g.nextID(ctx)
		if err != nil {
			return ids, fmt.Errorf("failed to generate ID at index %d: %w", i, err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}
