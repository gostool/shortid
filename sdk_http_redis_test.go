package shortid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestSDK_HTTPRedis_NextID 测试HTTP服务生成NextID
// 场景：通过HTTP API生成原始数字ID（uint64）
func TestSDK_HTTPRedis_NextID(t *testing.T) {
	redisAddr := getRedisAddr()
	if !checkRedisAvailable(redisAddr) {
		t.Skipf("Redis not available at %s, skipping test", redisAddr)
	}

	// 创建HTTP服务器
	server, err := NewHTTPServer(":0", redisAddr, BusinessOrder) // :0 表示自动分配端口
	if err != nil {
		t.Fatalf("Failed to create HTTP server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// 创建测试HTTP客户端
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 测试生成多个ID
	t.Log("=== HTTP服务：生成NextID ===")
	const count = 100
	ids := make([]uint64, count)
	idMap := make(map[uint64]bool, count)

	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(server.handleNextID))
	defer testServer.Close()

	startTime := time.Now()
	for i := 0; i < count; i++ {
		resp, err := client.Get(testServer.URL + "/nextid")
		if err != nil {
			t.Fatalf("HTTP request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
			continue
		}

		var response IDResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response.Error != "" {
			t.Errorf("Response error: %s", response.Error)
			continue
		}

		ids[i] = response.ID

		// 验证唯一性
		if idMap[response.ID] {
			t.Errorf("Duplicate ID found at index %d: %d", i, response.ID)
		}
		idMap[response.ID] = true
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
	t.Log("\n=== 前10个生成的ID（原始数字ID，10进制）===")
	for i := 0; i < 10 && i < len(ids); i++ {
		t.Logf("ID[%d]: %d (10进制)", i, ids[i])
	}
}

// TestSDK_HTTPRedis_Health 测试健康检查端点
func TestSDK_HTTPRedis_Health(t *testing.T) {
	redisAddr := getRedisAddr()
	if !checkRedisAvailable(redisAddr) {
		t.Skipf("Redis not available at %s, skipping test", redisAddr)
	}

	server, err := NewHTTPServer(":0", redisAddr, BusinessOrder)
	if err != nil {
		t.Fatalf("Failed to create HTTP server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(server.handleHealth))
	defer testServer.Close()

	// 测试健康检查
	resp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if result["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", result["status"])
	}
}

// TestSDK_HTTPRedis_Concurrent 测试并发HTTP请求
func TestSDK_HTTPRedis_Concurrent(t *testing.T) {
	redisAddr := getRedisAddr()
	if !checkRedisAvailable(redisAddr) {
		t.Skipf("Redis not available at %s, skipping test", redisAddr)
	}

	server, err := NewHTTPServer(":0", redisAddr, BusinessOrder)
	if err != nil {
		t.Fatalf("Failed to create HTTP server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(server.handleNextID))
	defer testServer.Close()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// 并发生成ID
	results := make(chan uint64, 100)
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		go func() {
			resp, err := client.Get(testServer.URL + "/nextid")
			if err != nil {
				errors <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errors <- fmt.Errorf("unexpected status: %d", resp.StatusCode)
				return
			}

			var response IDResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				errors <- err
				return
			}

			if response.Error != "" {
				errors <- fmt.Errorf("response error: %s", response.Error)
				return
			}

			results <- response.ID
		}()
	}

	// 收集结果
	idMap := make(map[uint64]bool)
	for i := 0; i < 100; i++ {
		select {
		case id := <-results:
			if idMap[id] {
				t.Errorf("Duplicate ID found in concurrent test: %d", id)
			}
			idMap[id] = true
		case err := <-errors:
			t.Errorf("HTTP request error: %v", err)
		}
	}

	t.Logf("✓ 并发测试：成功生成 100 个唯一ID")
	t.Logf("✓ 唯一ID数量: %d", len(idMap))
}

// TestSDK_HTTPRedis_MethodNotAllowed 测试不支持的方法
func TestSDK_HTTPRedis_MethodNotAllowed(t *testing.T) {
	redisAddr := getRedisAddr()
	if !checkRedisAvailable(redisAddr) {
		t.Skipf("Redis not available at %s, skipping test", redisAddr)
	}

	server, err := NewHTTPServer(":0", redisAddr, BusinessOrder)
	if err != nil {
		t.Fatalf("Failed to create HTTP server: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	// 创建测试服务器
	testServer := httptest.NewServer(http.HandlerFunc(server.handleNextID))
	defer testServer.Close()

	// 测试PUT方法（不支持）
	req, _ := http.NewRequest(http.MethodPut, testServer.URL+"/nextid", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", resp.StatusCode)
	}
}
