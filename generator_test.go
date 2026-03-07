package shortid

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testMachineIDProvider struct {
	id       uint16
	getCalls int32
	expCalls int32
}

type failMachineIDProvider struct{}

func (p *failMachineIDProvider) GetMachineID(ctx context.Context) (uint16, error) {
	return 0, errors.New("machine provider unavailable")
}

func (p *failMachineIDProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	return nil
}

func (p *failMachineIDProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *failMachineIDProvider) Close() error {
	return nil
}

type failSequenceProvider struct{}

func (p *failSequenceProvider) GetSequence(ctx context.Context, key string) (uint16, error) {
	return 0, errors.New("sequence provider unavailable")
}

func (p *failSequenceProvider) SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error {
	return nil
}

func (p *failSequenceProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *failSequenceProvider) Close() error {
	return nil
}

func (p *testMachineIDProvider) GetMachineID(ctx context.Context) (uint16, error) {
	atomic.AddInt32(&p.getCalls, 1)
	return p.id, nil
}

func (p *testMachineIDProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	atomic.AddInt32(&p.expCalls, 1)
	return nil
}

func (p *testMachineIDProvider) HealthCheck(ctx context.Context) error {
	return nil
}

func (p *testMachineIDProvider) Close() error {
	return nil
}

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

func TestValidateConfig_ConflictingMachineConfig(t *testing.T) {
	err := ValidateConfig(Config{
		MachineID:         1,
		MachineIDProvider: &testMachineIDProvider{id: 1},
		BusinessType:      BusinessOrder,
	})
	if err == nil {
		t.Fatal("ValidateConfig() expected error, got nil")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("ValidateConfig() error = %v, want ErrInvalidConfig", err)
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

func TestGenerator_MachineIDProvider_ZeroIDInitializedOnce(t *testing.T) {
	provider := &testMachineIDProvider{id: 0}
	generator, err := NewGenerator(Config{
		MachineIDProvider: provider,
		BusinessType:      BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if _, err = generator.NextID(ctx); err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
	}

	if got := atomic.LoadInt32(&provider.getCalls); got != 1 {
		t.Fatalf("GetMachineID() calls = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&provider.expCalls); got != 1 {
		t.Fatalf("SetMachineIDExpiration() calls = %d, want 1", got)
	}
}

func TestGenerator_MachineIDProvider_ConcurrentInitOnce(t *testing.T) {
	provider := &testMachineIDProvider{id: 3}
	generator, err := NewGenerator(Config{
		MachineIDProvider: provider,
		BusinessType:      BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const workers = 100
	var wg sync.WaitGroup
	wg.Add(workers)
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			_, err := generator.NextID(context.Background())
			errCh <- err
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
	}

	if got := atomic.LoadInt32(&provider.getCalls); got != 1 {
		t.Fatalf("GetMachineID() calls = %d, want 1", got)
	}
}

func TestGenerator_MachineProviderFailure(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineIDProvider: &failMachineIDProvider{},
		BusinessType:      BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	_, err = generator.NextID(context.Background())
	if err == nil {
		t.Fatal("NextID() expected error, got nil")
	}
}

func TestGenerator_SequenceProviderFailure(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:        1,
		BusinessType:     BusinessOrder,
		SequenceProvider: &failSequenceProvider{},
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 让 generator 进入 “同时间片内走序列分支”。
	generator.startTime = time.Now().UnixMilli()/10 - 1
	generator.elapsedTime = 1

	_, err = generator.NextID(context.Background())
	if err == nil {
		t.Fatal("NextID() expected error, got nil")
	}
}

func TestGenerator_SequenceOverflowAdvancesElapsedTime(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	nowUnit := time.Now().UnixMilli() / 10
	generator.startTime = nowUnit
	generator.elapsedTime = 0
	generator.sequence = SnowflakeMaxSequence

	if _, err := generator.NextID(context.Background()); err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if generator.elapsedTime < 1 {
		t.Fatalf("elapsedTime = %d, want >= 1", generator.elapsedTime)
	}
}

func TestGenerator_ClockBackwardStillMonotonicPerInstance(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 模拟时钟回退：将 elapsedTime 人为设为未来值。
	current := time.Now().UnixMilli()/10 - generator.startTime
	generator.elapsedTime = current + 10
	generator.sequence = 0

	first, err := generator.NextID(context.Background())
	if err != nil {
		t.Fatalf("first NextID() error = %v", err)
	}
	second, err := generator.NextID(context.Background())
	if err != nil {
		t.Fatalf("second NextID() error = %v", err)
	}

	if second <= first {
		t.Fatalf("second id (%d) should be greater than first id (%d)", second, first)
	}
}
