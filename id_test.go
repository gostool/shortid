package shortid

import (
	"sync"
	"testing"
	"time"
)

// TestIDGeneratorBasics 测试ID生成器基础功能
func TestIDGeneratorBasics(t *testing.T) {
	config := IDGeneratorConfig{
		BusinessType: BusinessOrder,
		EnableDate:   true,
		MachineID:    1,
	}
	gen := NewIDGenerator(config)

	t.Run("generate_single", func(t *testing.T) {
		id := gen.Generate()
		if len(id) < 3 {
			t.Errorf("ID too short: %s", id)
		}
		t.Logf("Generated ID: %s", id)
	})

	t.Run("generate_sequence", func(t *testing.T) {
		ids := make([]string, 10)
		for i := 0; i < 10; i++ {
			ids[i] = gen.GenerateWithTimestamp(time.Now().Unix())
		}

		// 检查唯一性
		idMap := make(map[string]bool)
		for _, id := range ids {
			if idMap[id] {
				t.Errorf("Duplicate ID: %s", id)
			}
			idMap[id] = true
		}

		// 检查业务前缀
		for _, id := range ids {
			if string(id[0]) != config.BusinessType.String() {
				t.Errorf("Wrong business prefix in ID: %s", id)
			}
		}
	})
}

// TestBusinessTypes 测试各种业务类型
func TestBusinessTypes(t *testing.T) {
	businessTypes := []struct {
		name   string
		typ    BusinessType
		prefix string
	}{
		{"订单", BusinessOrder, "3"},
		{"用户", BusinessUser, "2"},
		{"支付", BusinessPayment, "4"},
		{"分享", BusinessShare, "A"},
		{"邀请", BusinessInvite, "B"},
	}

	for _, bt := range businessTypes {
		t.Run(bt.name, func(t *testing.T) {
			config := IDGeneratorConfig{
				BusinessType: bt.typ,
				EnableDate:   true,
			}
			gen := NewIDGenerator(config)

			id := gen.Generate()
			if len(id) < 2 {
				t.Errorf("ID too short for %s: %s", bt.name, id)
			}

			if !hasPrefix(id, bt.prefix) {
				t.Errorf("Wrong prefix for %s, expected %s, got %s",
					bt.name, bt.prefix, id[:1])
			}

			t.Logf("%s ID: %s", bt.name, id)
		})
	}
}

// TestConcurrentGeneration 测试并发生成
func TestConcurrentGeneration(t *testing.T) {
	config := IDGeneratorConfig{
		BusinessType: BusinessOrder,
		EnableDate:   true,
		MachineID:    1,
	}
	gen := NewIDGenerator(config)

	const numGoroutines = 10
	const idsPerGoroutine = 100

	var wg sync.WaitGroup
	idChan := make(chan string, numGoroutines*idsPerGoroutine)

	// 启动多个goroutine并发生成ID
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGoroutine; j++ {
				id := gen.Generate()
				idChan <- id
			}
		}()
	}

	wg.Wait()
	close(idChan)

	// 检查唯一性
	idMap := make(map[string]bool)
	count := 0
	for id := range idChan {
		if idMap[id] {
			t.Errorf("Duplicate ID: %s", id)
		}
		idMap[id] = true
		count++
	}

	expectedCount := numGoroutines * idsPerGoroutine
	if count != expectedCount {
		t.Errorf("Expected %d IDs, got %d", expectedCount, count)
	}

	t.Logf("Generated %d unique IDs concurrently", count)
}

// TestIDParsing 测试ID解析
func TestIDParsing(t *testing.T) {
	config := IDGeneratorConfig{
		BusinessType: BusinessOrder,
		EnableDate:   true,
		DateBase:     "2024-01-01",
	}
	gen := NewIDGenerator(config)

	t.Run("parse_generated_id", func(t *testing.T) {
		id := gen.Generate()
		info, err := gen.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse ID %s: %v", id, err)
			return
		}

		if info.Business != config.BusinessType.String() {
			t.Errorf("Wrong business type, expected %s, got %s",
				config.BusinessType.String(), info.Business)
		}

		t.Logf("Parsed ID: %s", id)
		t.Logf("  Business: %s", info.Business)
		t.Logf("  Timestamp: %v", info.Timestamp)
		t.Logf("  Sequence: %d", info.Sequence)
	})
}

// TestIDGenerationSpeed 测试ID生成速度
func TestIDGenerationSpeed(t *testing.T) {
	configs := []struct {
		name   string
		config IDGeneratorConfig
	}{
		{
			name: "simple",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
			},
		},
		{
			name: "with_date",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
				EnableDate:   true,
			},
		},
		{
			name: "with_machine",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
				EnableDate:   true,
				MachineID:    12345,
			},
		},
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewIDGenerator(tc.config)

			start := time.Now()
			const count = 10000

			for i := 0; i < count; i++ {
				gen.Generate()
			}

			duration := time.Since(start)
			speed := float64(count) / duration.Seconds()

			t.Logf("%s: Generated %d IDs in %v (%.0f IDs/sec)",
				tc.name, count, duration, speed)

			if speed < 10000 {
				t.Errorf("Generation speed too low: %.0f IDs/sec", speed)
			}
		})
	}
}

// TestSpecialIDGenerators 测试特殊ID生成器
func TestSpecialIDGenerators(t *testing.T) {
	t.Run("share_code", func(t *testing.T) {
		testCases := []struct {
			name     string
			business BusinessType
			expireIn time.Duration
			data     int64
		}{
			{"7_day_share", BusinessShare, 7 * 24 * time.Hour, 12345},
			{"1_hour_share", BusinessShare, time.Hour, 67890},
			{"30_day_share", BusinessShare, 30 * 24 * time.Hour, 55555},
		}

		for _, tc := range testCases {
			code := GenerateShareCode(tc.business, tc.expireIn, tc.data)
			t.Logf("%s: %s", tc.name, code)

			if len(code) < 3 {
				t.Errorf("Share code too short: %s", code)
			}
		}
	})

	t.Run("invite_code", func(t *testing.T) {
		userIDs := []int64{
			1,
			123,
			12345,
			123456789,
			999999999999,
		}

		for _, uid := range userIDs {
			code := GenerateInviteCode(uid)
			t.Logf("User %d invite code: %s", uid, code)

			if len(code) != 2 { // B + 1-2 chars
				t.Errorf("Unexpected invite code length: %s", code)
			}
		}
	})

	t.Run("cache_key", func(t *testing.T) {
		testCases := []struct {
			prefix string
			id     int64
		}{
			{"user", 12345},
			{"session", 987654321},
			{"product", 5555},
			{"order", 123456789012345},
			{"temp", 0},
		}

		for _, tc := range testCases {
			key := GenerateCacheKey(tc.prefix, tc.id)
			t.Logf("Cache key %s:%d -> %s", tc.prefix, tc.id, key)

			if !hasPrefix(key, tc.prefix+":") {
				t.Errorf("Wrong cache key prefix: %s", key)
			}
		}
	})
}

// TestBatchGeneration 测试批量生成
func TestBatchGeneration(t *testing.T) {
	t.Run("batch_orders", func(t *testing.T) {
		ids := BatchOrderIDs(100)

		if len(ids) != 100 {
			t.Errorf("Expected 100 IDs, got %d", len(ids))
		}

		// 检查前缀
		for _, id := range ids {
			if string(id[0]) != BusinessOrder.String() {
				t.Errorf("Wrong prefix in batch order ID: %s", id)
			}
		}

		// 检查唯一性
		idMap := make(map[string]bool)
		duplicates := 0
		for _, id := range ids {
			if idMap[id] {
				duplicates++
			}
			idMap[id] = true
		}

		if duplicates > 0 {
			t.Errorf("Found %d duplicates in batch generation", duplicates)
		}
	})

	t.Run("batch_custom", func(t *testing.T) {
		generator := func() string {
			return DefaultUserGenerator.Generate()
		}

		ids := BatchQuick(50, generator)

		if len(ids) != 50 {
			t.Errorf("Expected 50 IDs, got %d", len(ids))
		}
	})
}

// TestRealWorldIDScenarios 真实世界ID场景测试
func TestRealWorldIDScenarios(t *testing.T) {
	t.Run("ecommerce_platform", func(t *testing.T) {
		t.Log("=== 电商平台ID场景 ===")

		// 订单ID
		orderID := GenerateOrderID(1001)
		t.Logf("订单ID: %s", orderID)

		// 支付流水
		paymentGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: BusinessPayment,
			EnableDate:   true,
			MachineID:    100,
		})
		paymentID := paymentGen.Generate()
		t.Logf("支付流水: %s", paymentID)

		// 退款单
		refundGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'R', // Refund
			EnableDate:   true,
			Prefix:       "RF",
		})
		refundID := refundGen.Generate()
		t.Logf("退款单号: %s", refundID)
	})

	t.Run("social_platform", func(t *testing.T) {
		t.Log("=== 社交平台ID场景 ===")

		// 用户ID
		userGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: BusinessUser,
			MachineID:    1,
		})
		userID := userGen.Generate()
		t.Logf("用户ID: %s", userID)

		// 动态ID
		postGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'P', // Post
			EnableDate:   true,
		})
		postID := postGen.Generate()
		t.Logf("动态ID: %s", postID)

		// 评论ID
		commentGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'c', // comment
			Prefix:       "cmt",
		})
		commentID := commentGen.Generate()
		t.Logf("评论ID: %s", commentID)

		// 分享链接（7天有效）
		shareCode := GenerateShareCode(BusinessShare, 7*24*time.Hour, 12345)
		t.Logf("分享链接: %s", shareCode)
	})

	t.Run("iot_platform", func(t *testing.T) {
		t.Log("=== 物联网平台ID场景 ===")

		// 设备ID
		deviceGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'D', // Device
			MachineID:    1001,
			EnableDate:   false, // 设备ID不需要日期
		})
		deviceID := deviceGen.Generate()
		t.Logf("设备ID: %s", deviceID)

		// 传感器数据ID
		sensorGen := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'S', // Sensor
			EnableDate:   true,
			MachineID:    1001,
		})
		sensorID := sensorGen.Generate()
		t.Logf("传感器数据ID: %s", sensorID)
	})
}

// TestIDGeneratorConfiguration 测试ID生成器配置
func TestIDGeneratorConfiguration(t *testing.T) {
	testConfigs := []struct {
		name   string
		config IDGeneratorConfig
		valid  bool
	}{
		{
			name: "minimal",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
			},
			valid: true,
		},
		{
			name: "with_date",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
				EnableDate:   true,
				DateBase:     "2024-01-01",
			},
			valid: true,
		},
		{
			name: "with_machine",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
				EnableDate:   true,
				MachineID:    9999,
			},
			valid: true,
		},
		{
			name: "with_prefix",
			config: IDGeneratorConfig{
				BusinessType: BusinessOrder,
				Prefix:       "TEST",
			},
			valid: true,
		},
	}

	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewIDGenerator(tc.config)
			id := gen.Generate()

			if len(id) == 0 {
				t.Errorf("Generated empty ID")
			}

			// 检查前缀
			if tc.config.Prefix != "" && !hasPrefix(id, tc.config.Prefix) {
				t.Errorf("Missing prefix %s in ID %s", tc.config.Prefix, id)
			}

			t.Logf("Config %s generated ID: %s", tc.name, id)
		})
	}
}

// 辅助函数
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
