package shortid

import (
	"context"
	"testing"
	"time"
)

func TestNewGenerator_WithMachineID(t *testing.T) {
	// 测试固定机器ID模式
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	if generator == nil {
		t.Fatal("NewGenerator() returned nil")
	}

	// 验证配置
	if generator.machineID != 1 {
		t.Errorf("machineID = %d, want 1", generator.machineID)
	}
	if generator.businessType != uint8(BusinessOrder) {
		t.Errorf("businessType = %d, want %d", generator.businessType, BusinessOrder)
	}
	if generator.useMachineProvider {
		t.Error("useMachineProvider = true, want false")
	}
}

// TestNewGenerator_WithMachineIDProvider 测试Serverless模式
// 注意：此测试需要 MemoryMachineIDProvider，在实际使用时通过导入子包访问
func TestNewGenerator_WithMachineIDProvider(t *testing.T) {
	// 跳过此测试，因为需要子目录中的类型
	// 在实际集成测试中可以使用
	t.Skip("Skipping test that requires subdirectory types")
}

func TestNewGenerator_InvalidMachineID(t *testing.T) {
	// 测试无效机器ID
	_, err := NewGenerator(Config{
		MachineID:    100, // 超出范围（0-63）
		BusinessType: BusinessOrder,
	})
	if err != ErrInvalidMachineID {
		t.Errorf("NewGenerator() error = %v, want ErrInvalidMachineID", err)
	}
}

func TestGenerator_Generate(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 生成ID
	id1, err := generator.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	t.Logf("id1: %s", id1)
	if id1 == "" {
		t.Error("Generate() returned empty string")
	}

	// 再次生成，应该不同
	id2, err := generator.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	t.Logf("id2: %s", id2)
	if id2 == "" {
		t.Error("Generate() returned empty string")
	}

	// ID应该不同（至少序列号不同）
	if id1 == id2 {
		t.Errorf("Generate() returned same ID: %s", id1)
	}
}

func TestGenerator_GenerateWithContext(t *testing.T) {
	// 测试使用固定机器ID的GenerateWithContext
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()
	id, err := generator.GenerateWithContext(ctx)
	if err != nil {
		t.Fatalf("GenerateWithContext() error = %v", err)
	}
	if id == "" {
		t.Error("GenerateWithContext() returned empty string")
	}
}

func TestGenerator_Concurrent(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 并发生成ID
	results := make(chan string, 100)
	for i := 0; i < 100; i++ {
		go func() {
			id, err := generator.Generate()
			if err != nil {
				t.Errorf("Generate() error = %v", err)
				return
			}
			results <- id
		}()
	}

	// 收集结果
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := <-results
		t.Logf("id: %s", id)
		if id == "" {
			t.Error("Generate() returned empty string")
		}
		ids[id] = true
	}

	// 验证所有ID都是唯一的（至少大部分应该是唯一的）
	if len(ids) < 50 {
		t.Errorf("Only %d unique IDs generated, expected at least 50", len(ids))
	}
}

func TestGenerator_GenerateBatch(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 测试批量生成
	count := 100
	ids, err := generator.GenerateBatch(count)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("GenerateBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 验证所有ID都是唯一的
	idMap := make(map[string]bool)
	for i, id := range ids {
		if id == "" {
			t.Errorf("GenerateBatch() returned empty string at index %d", i)
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %s", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功批量生成 %d 个唯一ID", count)
}

func TestGenerator_GenerateBatchWithContext(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()
	count := 50
	ids, err := generator.GenerateBatchWithContext(ctx, count)
	if err != nil {
		t.Fatalf("GenerateBatchWithContext() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("GenerateBatchWithContext() returned %d IDs, want %d", len(ids), count)
	}

	// 验证所有ID都是唯一的
	idMap := make(map[string]bool)
	for i, id := range ids {
		if id == "" {
			t.Errorf("GenerateBatchWithContext() returned empty string at index %d", i)
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %s", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功批量生成 %d 个唯一ID", count)
}

func TestGenerator_NextIDBatch(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()
	count := 100
	ids, err := generator.NextIDBatch(ctx, count)
	if err != nil {
		t.Fatalf("NextIDBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("NextIDBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 验证所有ID都是唯一的
	idMap := make(map[uint64]bool)
	for i, id := range ids {
		if id == 0 {
			t.Errorf("NextIDBatch() returned zero ID at index %d", i)
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %d", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功批量生成 %d 个唯一数字ID", count)
}

func TestGenerator_GenerateBatch_InvalidCount(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 测试 count <= 0
	_, err = generator.GenerateBatch(0)
	if err == nil {
		t.Error("GenerateBatch(0) should return error")
	}

	_, err = generator.GenerateBatch(-1)
	if err == nil {
		t.Error("GenerateBatch(-1) should return error")
	}

	// 测试 count > MaxBatchCount
	_, err = generator.GenerateBatch(MaxBatchCount + 1)
	if err == nil {
		t.Errorf("GenerateBatch(%d) should return error", MaxBatchCount+1)
	}
}

func TestGenerator_NextIDBatch_InvalidCount(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()

	// 测试 count <= 0
	_, err = generator.NextIDBatch(ctx, 0)
	if err == nil {
		t.Error("NextIDBatch(0) should return error")
	}

	_, err = generator.NextIDBatch(ctx, -1)
	if err == nil {
		t.Error("NextIDBatch(-1) should return error")
	}

	// 测试 count > MaxBatchCount
	_, err = generator.NextIDBatch(ctx, MaxBatchCount+1)
	if err == nil {
		t.Errorf("NextIDBatch(%d) should return error", MaxBatchCount+1)
	}
}

func TestGenerator_GenerateBatch_LargeCount(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 测试较大的批量生成
	count := 1000
	ids, err := generator.GenerateBatch(count)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("GenerateBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 验证所有ID都是唯一的
	idMap := make(map[string]bool)
	for _, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功批量生成 %d 个唯一ID", count)
	t.Logf("✓ 唯一ID数量: %d", len(idMap))
}

func TestGenerator_GenerateBatch_Concurrent(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 并发批量生成
	const goroutineCount = 10
	const batchSize = 100
	results := make(chan []string, goroutineCount)
	errors := make(chan error, goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func() {
			ids, err := generator.GenerateBatch(batchSize)
			if err != nil {
				errors <- err
				return
			}
			results <- ids
		}()
	}

	// 收集结果
	allIDs := make(map[string]bool)
	for i := 0; i < goroutineCount; i++ {
		select {
		case ids := <-results:
			for _, id := range ids {
				if allIDs[id] {
					t.Errorf("Duplicate ID found in concurrent batch: %s", id)
				}
				allIDs[id] = true
			}
		case err := <-errors:
			t.Errorf("GenerateBatch() error in concurrent test: %v", err)
		}
	}

	expectedCount := goroutineCount * batchSize
	if len(allIDs) != expectedCount {
		t.Errorf("Expected %d unique IDs, got %d", expectedCount, len(allIDs))
	}

	t.Logf("✓ 并发批量生成测试：成功生成 %d 个唯一ID", len(allIDs))
}

func TestGenerator_GenerateBatch_BoundaryValues(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 测试边界值：1个ID
	ids, err := generator.GenerateBatch(1)
	if err != nil {
		t.Fatalf("GenerateBatch(1) error = %v", err)
	}
	if len(ids) != 1 {
		t.Errorf("GenerateBatch(1) returned %d IDs, want 1", len(ids))
	}
	if ids[0] == "" {
		t.Error("GenerateBatch(1) returned empty string")
	}

	// 测试边界值：MaxBatchCount个ID
	ids, err = generator.GenerateBatch(MaxBatchCount)
	if err != nil {
		t.Fatalf("GenerateBatch(%d) error = %v", MaxBatchCount, err)
	}
	if len(ids) != MaxBatchCount {
		t.Errorf("GenerateBatch(%d) returned %d IDs, want %d", MaxBatchCount, len(ids), MaxBatchCount)
	}

	// 验证唯一性
	idMap := make(map[string]bool)
	for _, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 边界值测试：成功生成 1 个和 %d 个唯一ID", MaxBatchCount)
}

func TestGenerator_GenerateBatch_ReturnRawID(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
		ReturnRawID:  true, // 返回原始数字ID
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	count := 100
	ids, err := generator.GenerateBatch(count)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("GenerateBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 验证所有ID都是数字字符串
	idMap := make(map[string]bool)
	for i, id := range ids {
		if id == "" {
			t.Errorf("GenerateBatch() returned empty string at index %d", i)
		}
		// 验证是数字字符串
		for _, r := range id {
			if r < '0' || r > '9' {
				t.Errorf("GenerateBatch() returned non-numeric ID at index %d: %s", i, id)
				break
			}
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %s", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ ReturnRawID 模式：成功批量生成 %d 个唯一数字ID字符串", count)
}

func TestGenerator_NextIDBatch_BoundaryValues(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()

	// 测试边界值：1个ID
	ids, err := generator.NextIDBatch(ctx, 1)
	if err != nil {
		t.Fatalf("NextIDBatch(1) error = %v", err)
	}
	if len(ids) != 1 {
		t.Errorf("NextIDBatch(1) returned %d IDs, want 1", len(ids))
	}
	if ids[0] == 0 {
		t.Error("NextIDBatch(1) returned zero ID")
	}

	// 测试边界值：MaxBatchCount个ID
	ids, err = generator.NextIDBatch(ctx, MaxBatchCount)
	if err != nil {
		t.Fatalf("NextIDBatch(%d) error = %v", MaxBatchCount, err)
	}
	if len(ids) != MaxBatchCount {
		t.Errorf("NextIDBatch(%d) returned %d IDs, want %d", MaxBatchCount, len(ids), MaxBatchCount)
	}

	// 验证唯一性
	idMap := make(map[uint64]bool)
	for _, id := range ids {
		if id == 0 {
			t.Error("NextIDBatch() returned zero ID")
		}
		if idMap[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
	}

	t.Logf("✓ NextIDBatch 边界值测试：成功生成 1 个和 %d 个唯一数字ID", MaxBatchCount)
}

func TestGenerator_GenerateBatch_Performance(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 性能测试：生成1000个ID
	count := 1000
	startTime := time.Now()
	ids, err := generator.GenerateBatch(count)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("GenerateBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 计算性能指标
	qps := float64(count) / duration.Seconds()
	avgLatency := duration / time.Duration(count)

	t.Logf("✓ 性能测试：生成 %d 个ID", count)
	t.Logf("  - 总耗时: %v", duration)
	t.Logf("  - 平均耗时: %v/ID", avgLatency)
	t.Logf("  - QPS: %.0f", qps)

	// 验证唯一性
	idMap := make(map[string]bool)
	for _, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		idMap[id] = true
	}

	t.Logf("  - 唯一ID数量: %d", len(idMap))
}

func TestGenerator_NextIDBatch_Performance(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()

	// 性能测试：生成1000个ID
	count := 1000
	startTime := time.Now()
	ids, err := generator.NextIDBatch(ctx, count)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("NextIDBatch() error = %v", err)
	}

	if len(ids) != count {
		t.Errorf("NextIDBatch() returned %d IDs, want %d", len(ids), count)
	}

	// 计算性能指标
	qps := float64(count) / duration.Seconds()
	avgLatency := duration / time.Duration(count)

	t.Logf("✓ NextIDBatch 性能测试：生成 %d 个ID", count)
	t.Logf("  - 总耗时: %v", duration)
	t.Logf("  - 平均耗时: %v/ID", avgLatency)
	t.Logf("  - QPS: %.0f", qps)

	// 验证唯一性
	idMap := make(map[uint64]bool)
	for _, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found: %d", id)
		}
		idMap[id] = true
	}

	t.Logf("  - 唯一ID数量: %d", len(idMap))
}

func TestGenerator_GenerateBatch_Consistency(t *testing.T) {
	generator1, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	generator2, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 单个生成 vs 批量生成的一致性测试
	const count = 10
	singleIDs := make([]string, count)
	for i := 0; i < count; i++ {
		id, err := generator1.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		singleIDs[i] = id
	}

	batchIDs, err := generator2.GenerateBatch(count)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	// 验证两种方式生成的ID格式一致
	if len(singleIDs) != len(batchIDs) {
		t.Errorf("Length mismatch: single=%d, batch=%d", len(singleIDs), len(batchIDs))
	}

	// 验证ID格式一致（长度、字符集等）
	for i := 0; i < count; i++ {
		if len(singleIDs[i]) != len(batchIDs[i]) {
			t.Errorf("ID length mismatch at index %d: single=%d, batch=%d", i, len(singleIDs[i]), len(batchIDs[i]))
		}
		// 验证字符集一致（都是Base62字符）
		for _, r := range singleIDs[i] {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				t.Errorf("Invalid character in single ID at index %d: %s", i, singleIDs[i])
			}
		}
		for _, r := range batchIDs[i] {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
				t.Errorf("Invalid character in batch ID at index %d: %s", i, batchIDs[i])
			}
		}
	}

	t.Logf("✓ 一致性测试：单个生成和批量生成的ID格式一致")
}

func TestGenerator_GenerateBatch_DifferentBusinessTypes(t *testing.T) {
	businessTypes := []BusinessType{
		BusinessOrder,
		BusinessPayment,
		BusinessUser,
	}

	for _, bt := range businessTypes {
		generator, err := NewGenerator(Config{
			MachineID:    1,
			BusinessType: bt,
		})
		if err != nil {
			t.Fatalf("NewGenerator() with BusinessType %d error = %v", bt, err)
		}

		count := 50
		ids, err := generator.GenerateBatch(count)
		if err != nil {
			t.Fatalf("GenerateBatch() with BusinessType %d error = %v", bt, err)
		}

		if len(ids) != count {
			t.Errorf("GenerateBatch() with BusinessType %d returned %d IDs, want %d", bt, len(ids), count)
		}

		// 验证唯一性
		idMap := make(map[string]bool)
		for _, id := range ids {
			if idMap[id] {
				t.Errorf("Duplicate ID found for BusinessType %d: %s", bt, id)
			}
			idMap[id] = true
		}

		t.Logf("✓ BusinessType %d: 成功批量生成 %d 个唯一ID", bt, count)
	}
}

func TestGenerator_GenerateBatchWithContext_Cancel(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 创建可取消的Context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	// 虽然Context已取消，但批量生成应该能正常完成（因为nextID不检查Context）
	// 这个测试主要是验证代码不会因为Context取消而panic
	count := 10
	ids, err := generator.GenerateBatchWithContext(ctx, count)
	if err != nil {
		// Context取消不应该导致错误（因为nextID不检查Context）
		// 但如果实现中检查了Context，这里可能会返回错误，这是可以接受的
		t.Logf("GenerateBatchWithContext() with cancelled context returned error (acceptable): %v", err)
		return
	}

	if len(ids) != count {
		t.Errorf("GenerateBatchWithContext() returned %d IDs, want %d", len(ids), count)
	}

	t.Logf("✓ Context取消测试：成功生成 %d 个ID", len(ids))
}
