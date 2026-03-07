package shortid

import (
	"context"
	"net"
	"os"
	"testing"
	"time"
)

const requireRedisEnv = "SHORTID_REQUIRE_REDIS_TESTS"

// getRedisAddr 获取Redis地址，支持环境变量配置
func getRedisAddr() string {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		return "localhost:6379"
	}
	return addr
}

func requireRedisForSDKTest(t *testing.T) string {
	t.Helper()

	redisAddr := getRedisAddr()
	if checkRedisAvailable(redisAddr) {
		return redisAddr
	}

	if os.Getenv(requireRedisEnv) == "1" {
		t.Fatalf("Redis is required but unavailable at %s (env %s=1)", redisAddr, requireRedisEnv)
	}
	t.Skipf("Redis not available at %s, skipping test. Set %s=1 to make Redis tests mandatory.", redisAddr, requireRedisEnv)
	return ""
}

// checkRedisAvailable 检查Redis是否可用
// 直接创建Redis客户端来测试连接
func checkRedisAvailable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// createRedisMachineIDProvider 创建Redis机器ID提供者
// 使用http_server.go中定义的函数
func createRedisMachineIDProvider(addr string) (MachineIDProvider, error) {
	return createRedisMachineIDProviderForHTTP(addr)
}

// createRedisSequenceProvider 创建Redis序列号提供者
// 使用http_server.go中定义的函数
func createRedisSequenceProvider(addr string) (SequenceProvider, error) {
	return createRedisSequenceProviderForHTTP(addr)
}

// TestSDK_ServerlessRedis_ShortID 测试Serverless Redis模式生成短ID
// 场景：完整Serverless模式（机器ID + 序列号都使用Redis）
// 注意：需要Redis环境，如果Redis不可用会自动跳过
//
// 使用方法：
//  1. 启动Redis: docker run -d -p 6379:6379 redis
//  2. 运行测试: go test -v -run TestSDK_ServerlessRedis
//  3. 或设置环境变量: REDIS_ADDR=redishost:6379 go test -v -run TestSDK_ServerlessRedis
func TestSDK_ServerlessRedis_ShortID(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	ctx := context.Background()

	// 创建Redis机器ID提供者
	machineProvider, err := createRedisMachineIDProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create machine provider: %v", err)
	}
	defer machineProvider.Close()

	// 创建Redis序列号提供者
	sequenceProvider, err := createRedisSequenceProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create sequence provider: %v", err)
	}
	defer sequenceProvider.Close()

	// 创建ID生成器
	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		SequenceProvider:  sequenceProvider, // 使用分布式序列号
		BusinessType:      BusinessOrder,
		ReturnRawID:       false, // 返回短ID
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 生成短ID
	t.Log("=== Serverless Redis模式：生成短ID ===")
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

	// 显示前10个ID作为示例
	t.Log("\n=== 前10个生成的ID（短ID，Base62编码）===")
	for i := 0; i < 10 && i < len(ids); i++ {
		t.Logf("ID[%d]: %s (长度: %d)", i, ids[i], len(ids[i]))
	}
}

// TestSDK_ServerlessRedis_NextID 测试Serverless Redis模式生成原始数字ID
// 场景：完整Serverless模式（机器ID + 序列号都使用Redis）
func TestSDK_ServerlessRedis_NextID(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	ctx := context.Background()

	machineProvider, err := createRedisMachineIDProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create machine provider: %v", err)
	}
	defer machineProvider.Close()

	// sequenceProvider, err := createRedisSequenceProvider(redisAddr)
	// if err != nil {
	// 	t.Fatalf("Failed to create sequence provider: %v", err)
	// }
	// defer sequenceProvider.Close()

	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		// SequenceProvider:  sequenceProvider,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 使用NextID方法获取原始数字ID
	t.Log("=== Serverless Redis模式：生成原始数字ID（uint64）===")
	const count = 10000
	ids := make([]uint64, count)

	startTime := time.Now()
	for i := 0; i < count; i++ {
		id, err := generator.NextID(ctx)
		if err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
		ids[i] = id
	}
	duration := time.Since(startTime)

	// 验证唯一性
	idMap := make(map[uint64]bool)
	for i, id := range ids {
		if idMap[id] {
			t.Errorf("Duplicate ID found at index %d: %d", i, id)
		}
		idMap[id] = true
	}

	t.Logf("✓ 成功生成 %d 个唯一ID", count)
	t.Logf("✓ 唯一ID数量: %d", len(idMap))
	t.Logf("✓ 生成耗时: %v", duration)
	t.Logf("✓ 平均耗时: %v/ID", duration/time.Duration(count))
	t.Logf("✓ QPS: %.0f", float64(count)/duration.Seconds())

	// 显示前10个ID作为示例
	t.Log("\n=== 前10个生成的ID（原始数字ID，10进制）===")
	for i := 0; i < 10 && i < len(ids); i++ {
		t.Logf("ID[%d]: %d (10进制)", i, ids[i])
	}
}

// TestSDK_ServerlessRedis_Simplified 测试简化Serverless模式
// 场景：仅机器ID使用Redis，序列号使用本地模式
func TestSDK_ServerlessRedis_Simplified(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	ctx := context.Background()

	machineProvider, err := createRedisMachineIDProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create machine provider: %v", err)
	}
	defer machineProvider.Close()

	// 不设置SequenceProvider，使用默认的本地序列号
	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		// SequenceProvider 不设置，使用默认的本地序列号
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 生成短ID
	t.Log("=== 简化Serverless模式：仅机器ID使用Redis ===")
	const count = 10000
	ids := make([]string, count)
	idMap := make(map[string]bool, count)

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

	t.Logf("✓ 成功生成 %d 个唯一ID", count)
	t.Logf("✓ 唯一ID数量: %d", len(idMap))

	// 显示前5个ID作为示例
	t.Log("\n=== 前5个生成的ID ===")
	for i := 0; i < 5 && i < len(ids); i++ {
		t.Logf("ID[%d]: %s", i, ids[i])
	}
}

// TestSDK_ServerlessRedis_Concurrent 测试并发场景
func TestSDK_ServerlessRedis_Concurrent(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	ctx := context.Background()

	machineProvider, err := createRedisMachineIDProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create machine provider: %v", err)
	}
	defer machineProvider.Close()

	sequenceProvider, err := createRedisSequenceProvider(redisAddr)
	if err != nil {
		t.Fatalf("Failed to create sequence provider: %v", err)
	}
	defer sequenceProvider.Close()

	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		SequenceProvider:  sequenceProvider,
		BusinessType:      BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 并发生成ID
	results := make(chan string, 100)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func() {
			id, err := generator.GenerateWithContext(ctx)
			if err != nil {
				errors <- err
				return
			}
			results <- id
		}()
	}

	// 收集结果
	idMap := make(map[string]bool)
	for i := 0; i < 100; i++ {
		select {
		case id := <-results:
			if idMap[id] {
				t.Errorf("Duplicate ID found in concurrent test: %s", id)
			}
			idMap[id] = true
		case err := <-errors:
			t.Errorf("GenerateWithContext() error = %v", err)
		}
	}

	t.Logf("✓ 并发测试：成功生成 100 个唯一ID")
	t.Logf("✓ 唯一ID数量: %d", len(idMap))
}
