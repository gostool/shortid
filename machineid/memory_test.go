package shortid

import (
	"context"
	"testing"
)

func TestMemoryMachineIDProvider_GetMachineID(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	// 测试基本功能
	id1, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}
	if id1 != 1 {
		t.Errorf("GetMachineID() = %d, want 1", id1)
	}

	id2, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}
	if id2 != 2 {
		t.Errorf("GetMachineID() = %d, want 2", id2)
	}

	// 测试取模64 - 验证ID在有效范围内
	// 前面已经调用了2次，再调用62次，总共64次
	for i := 0; i < 62; i++ {
		id, err := provider.GetMachineID(ctx)
		if err != nil {
			t.Fatalf("GetMachineID() error = %v", err)
		}
		if id > 63 {
			t.Errorf("GetMachineID() returned invalid id = %d, want 0-63", id)
		}
	}

	// 第65个应该在有效范围内
	id65, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}
	if id65 > 63 {
		t.Errorf("GetMachineID() returned invalid id = %d, want 0-63", id65)
	}
}

func TestMemoryMachineIDProvider_Concurrent(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	// 并发测试
	results := make(chan uint16, 100)
	for i := 0; i < 100; i++ {
		go func() {
			id, err := provider.GetMachineID(ctx)
			if err != nil {
				t.Errorf("GetMachineID() error = %v", err)
				return
			}
			results <- id
		}()
	}

	// 收集结果
	ids := make(map[uint16]bool)
	for i := 0; i < 100; i++ {
		id := <-results
		if id > 63 {
			t.Errorf("GetMachineID() returned invalid id = %d, want 0-63", id)
		}
		ids[id] = true
	}

	// 验证所有ID都在有效范围内
	if len(ids) == 0 {
		t.Error("No IDs generated")
	}
}

func TestMemoryMachineIDProvider_OtherMethods(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	// 测试 SetMachineIDExpiration（应该总是成功）
	if err := provider.SetMachineIDExpiration(ctx, 1, 0); err != nil {
		t.Errorf("SetMachineIDExpiration() error = %v", err)
	}

	// 测试 HealthCheck（应该总是成功）
	if err := provider.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// 测试 Close（应该总是成功）
	if err := provider.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
