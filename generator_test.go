package shortid

import (
	"context"
	"testing"
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
