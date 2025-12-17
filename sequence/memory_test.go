package shortid

import (
	"context"
	"testing"
)

func TestMemorySequenceProvider_GetSequence(t *testing.T) {
	provider := NewMemorySequenceProvider()
	ctx := context.Background()

	// 测试基本功能
	key := "test-key"
	seq1, err := provider.GetSequence(ctx, key)
	if err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if seq1 != 1 {
		t.Errorf("GetSequence() = %d, want 1", seq1)
	}

	seq2, err := provider.GetSequence(ctx, key)
	if err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if seq2 != 2 {
		t.Errorf("GetSequence() = %d, want 2", seq2)
	}

	// 测试不同key的序列号独立
	key2 := "test-key-2"
	seq3, err := provider.GetSequence(ctx, key2)
	if err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if seq3 != 1 {
		t.Errorf("GetSequence() with different key = %d, want 1", seq3)
	}

	// 测试取模128 - 需要调用126次才能让下一个变成0
	// 因为前面已经调用了3次（seq1=1, seq2=2, seq3=1），所以再调用125次
	for i := 0; i < 125; i++ {
		_, err := provider.GetSequence(ctx, key)
		if err != nil {
			t.Fatalf("GetSequence() error = %v", err)
		}
	}

	// 第129个应该是 0 (128 % 128 = 0)
	seq129, err := provider.GetSequence(ctx, key)
	if err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if seq129 != 0 {
		t.Errorf("GetSequence() after 128 calls = %d, want 0", seq129)
	}
}

func TestMemorySequenceProvider_Concurrent(t *testing.T) {
	provider := NewMemorySequenceProvider()
	ctx := context.Background()
	key := "concurrent-key"

	// 并发测试
	results := make(chan uint16, 100)
	for i := 0; i < 100; i++ {
		go func() {
			seq, err := provider.GetSequence(ctx, key)
			if err != nil {
				t.Errorf("GetSequence() error = %v", err)
				return
			}
			results <- seq
		}()
	}

	// 收集结果
	sequences := make(map[uint16]bool)
	for i := 0; i < 100; i++ {
		seq := <-results
		if seq > 127 {
			t.Errorf("GetSequence() returned invalid sequence = %d, want 0-127", seq)
		}
		sequences[seq] = true
	}

	// 验证所有序列号都在有效范围内
	if len(sequences) == 0 {
		t.Error("No sequences generated")
	}
}

func TestMemorySequenceProvider_OtherMethods(t *testing.T) {
	provider := NewMemorySequenceProvider()
	ctx := context.Background()

	// 测试 SetSequenceExpiration（应该总是成功）
	if err := provider.SetSequenceExpiration(ctx, "test-key", 0); err != nil {
		t.Errorf("SetSequenceExpiration() error = %v", err)
	}

	// 测试 HealthCheck（应该总是成功）
	if err := provider.HealthCheck(ctx); err != nil {
		t.Errorf("HealthCheck() error = %v", err)
	}

	// 测试 Close（应该清理资源）
	if err := provider.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
