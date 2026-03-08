package shortid

import (
	"context"
	"testing"
)

// TestSDK_SingleRedis_NextID 测试单机Redis模式（固定机器ID + Redis序列号）
func TestSDK_SingleRedis_NextID(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	sequenceProvider, err := createRedisSequenceProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create sequence provider: %v", err)
	}
	defer sequenceProvider.Close()

	generator, err := NewGenerator(Config{
		MachineID:        1,
		SequenceProvider: sequenceProvider,
		BusinessType:     BusinessOrder,
		ReturnRawID:      true,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const count = 5000
	idMap := make(map[uint64]struct{}, count)
	ctx := context.Background()

	for i := 0; i < count; i++ {
		id, genErr := generator.NextID(ctx)
		if genErr != nil {
			t.Fatalf("NextID() error = %v", genErr)
		}
		if _, exists := idMap[id]; exists {
			t.Fatalf("Duplicate ID found at index %d: %d", i, id)
		}
		idMap[id] = struct{}{}
	}

	if got := len(idMap); got != count {
		t.Fatalf("unique IDs = %d, want %d", got, count)
	}
}

// TestSDK_SingleRedis_GenerateShortID 测试单机Redis模式短ID输出
func TestSDK_SingleRedis_GenerateShortID(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	sequenceProvider, err := createRedisSequenceProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create sequence provider: %v", err)
	}
	defer sequenceProvider.Close()

	generator, err := NewGenerator(Config{
		MachineID:        2,
		SequenceProvider: sequenceProvider,
		BusinessType:     BusinessOrder,
		ReturnRawID:      false,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	id, err := generator.GenerateWithContext(context.Background())
	if err != nil {
		t.Fatalf("GenerateWithContext() error = %v", err)
	}
	if id == "" {
		t.Fatal("GenerateWithContext() returned empty id")
	}
}
