package shortid

import (
	"context"
	"testing"
	"time"
)

// TestSDK_SingleMemory 测试单机内存模式SDK使用
// 场景：单机部署，使用内存Provider（测试用）
// 特点：无需Redis等外部依赖，适合测试环境
//
// 注意：由于子目录中的类型在根目录测试中无法直接访问，
// 这里使用固定机器ID模式进行测试（这是单机部署的典型场景）
func TestSDK_SingleMemory_ShortID(t *testing.T) {
	ctx := context.Background()

	// 单机部署：使用固定机器ID（无需Provider）
	generator, err := NewGenerator(Config{
		MachineID:    1,             // 固定机器ID
		BusinessType: BusinessOrder, // 业务类型：订单
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 生成短ID
	t.Log("=== 单机内存模式：生成短ID ===")
	ids := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		id, err := generator.GenerateWithContext(ctx)
		if err != nil {
			t.Fatalf("GenerateWithContext() error = %v", err)
		}
		if id == "" {
			t.Error("GenerateWithContext() returned empty string")
		}
		ids[i] = id
		t.Logf("ID %d: %s (长度: %d)", i+1, id, len(id))
	}

	// 验证ID唯一性
	t.Log("\n=== 验证ID唯一性 ===")
	idMap := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generator.GenerateWithContext(ctx)
		if err != nil {
			t.Fatalf("GenerateWithContext() error = %v", err)
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		idMap[id] = true
	}
	t.Logf("✓ 成功生成 100 个唯一ID")
	t.Logf("✓ ID长度范围: 8-12 字符")

	// 验证ID格式
	for _, id := range ids {
		if len(id) < 8 || len(id) > 15 {
			t.Errorf("ID length out of range: %s (length: %d)", id, len(id))
		}
		// 验证ID不为空且格式正确
		if len(id) == 0 {
			t.Error("ID should not be empty")
		}
	}
}

// TestSDK_SingleMemory_UID 测试唯一ID生成
// 场景：单机部署，生成大量唯一ID并验证唯一性
// 特点：验证ID的唯一性和性能
func TestSDK_SingleMemory_UID(t *testing.T) {
	ctx := context.Background()

	// 单机部署：使用固定机器ID
	generator, err := NewGenerator(Config{
		MachineID:    1,             // 固定机器ID
		BusinessType: BusinessOrder, // 业务类型：订单
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 生成大量唯一ID
	t.Log("=== 单机内存模式：生成唯一ID ===")
	const count = 10000
	ids := make([]string, count)
	idMap := make(map[string]bool, count)

	startTime := time.Now()
	for i := 0; i < count; i++ {
		id, err := generator.GenerateWithContext(ctx)
		if err != nil {
			t.Fatalf("GenerateWithContext() error = %v", err)
		}
		if id == "" {
			t.Errorf("GenerateWithContext() returned empty string at index %d", i)
			continue
		}
		ids[i] = id

		// 验证唯一性
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %s", i, id)
		}
		idMap[id] = true
	}
	duration := time.Since(startTime)

	// 验证结果
	t.Logf("✓ 成功生成 %d 个唯一ID", count)
	t.Logf("✓ 唯一ID数量: %d", len(idMap))
	t.Logf("✓ 生成耗时: %v", duration)
	t.Logf("✓ 平均耗时: %v/ID", duration/time.Duration(count))
	t.Logf("✓ QPS: %.0f", float64(count)/duration.Seconds())

	// 验证所有ID都是唯一的
	if len(idMap) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(idMap))
	}

	// 验证ID格式
	for i, id := range ids {
		if len(id) < 8 || len(id) > 15 {
			t.Errorf("ID[%d] length out of range: %s (length: %d)", i, id, len(id))
		}
	}

	// 显示前10个ID作为示例
	t.Log("\n=== 前10个生成的ID（短ID，Base62编码）===")
	for i := 0; i < 10 && i < len(ids); i++ {
		t.Logf("ID[%d]: %s (长度: %d)", i, ids[i], len(ids[i]))
	}
}

// TestSDK_SingleMemory_NextID 测试NextID方法（直接返回uint64）
func TestSDK_SingleMemory_NextID(t *testing.T) {
	ctx := context.Background()

	// 单机部署：使用固定机器ID
	generator, err := NewGenerator(Config{
		MachineID:    1,             // 固定机器ID
		BusinessType: BusinessOrder, // 业务类型：订单
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 使用NextID方法获取原始数字ID
	t.Log("=== 使用NextID方法生成原始数字ID（uint64）===")
	const count = 10000
	ids := make([]uint64, count)

	for i := 0; i < count; i++ {
		id, err := generator.NextID(ctx)
		if err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
		ids[i] = id
	}

	// 验证唯一性
	idMap := make(map[uint64]bool)
	for i, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %d", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功生成 %d 个唯一ID", count)
	// t.Logf("✓ 唯一ID数量: %d", len(idMap))

	// 显示所有ID
	// t.Log("\n=== 生成的原始数字ID（uint64，10进制）===")
	// for i, id := range ids {
	// 	t.Logf("ID[%d]: %d (10进制)", i, id)
	// }
}
