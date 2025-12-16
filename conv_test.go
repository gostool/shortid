package shortid

import (
	"fmt"
	"testing"
	"time"
)

// TestBasicEncoding 测试基础编码功能
func TestBasicEncoding(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{10, "a"},
		{35, "z"},
		{36, "A"},
		{61, "Z"},
		{62, "10"},
		{123, "1Z"},
		{1000, "g8"},
		{999999, "4c91"},
		{-1, "-1"},
		{-10, "-a"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("int64_%d", tt.input), func(t *testing.T) {
			result := ToBase64(tt.input)
			if result != tt.expected {
				t.Errorf("ToBase64(%d) = %q, expect %q", tt.input, result, tt.expected)
			}

			// 测试解码
			decoded, err := FromBase64(result)
			if err != nil {
				t.Errorf("FromBase64(%q) error: %v", result, err)
			} else if decoded != tt.input {
				t.Errorf("FromBase64(%q) = %d, expect %d", result, decoded, tt.input)
			}
		})
	}
}

// TestTimestampEncoding 测试时间戳编码
func TestTimestampEncoding(t *testing.T) {
	testCases := []struct {
		name      string
		timestamp int64
		desc      string
	}{
		{"unix_epoch", 0, "Unix纪元"},
		{"2020_start", 1577836800, "2020-01-01"},
		{"2021_start", 1609459200, "2021-01-01"},
		{"today", time.Now().Unix(), "今天"},
		{"2100_end", 4133951999, "2100-12-31"},
	}

	t.Run("short_encoding", func(t *testing.T) {
		for _, tc := range testCases {
			encoded := ToTimestampShort(tc.timestamp)
			t.Logf("%s (%s): %d -> %s (length: %d)",
				tc.name, tc.desc, tc.timestamp, encoded, len(encoded))

			// 验证解码
			decoded, err := FromTimestampShort(encoded)
			if err != nil {
				t.Errorf("Failed to decode %s: %v", encoded, err)
			} else if decoded != tc.timestamp {
				t.Errorf("Decode mismatch: %d -> %s -> %d",
					tc.timestamp, encoded, decoded)
			}
		}
	})

	t.Run("date_encoding", func(t *testing.T) {
		dates := []struct {
			year, month, day int
		}{
			{2024, 12, 30},
			{2024, 12, 31},
			{2025, 1, 1},
			{2025, 6, 15},
			{2030, 1, 1},
		}

		for _, d := range dates {
			encoded := EncodeDate(d.year, d.month, d.day)
			t.Logf("%04d-%02d-%02d -> %s", d.year, d.month, d.day, encoded)
		}
	})
}

// TestIDGeneration 测试ID生成
func TestIDGeneration(t *testing.T) {
	t.Run("order_id", func(t *testing.T) {
		for i := int64(1); i <= 10; i++ {
			orderID := GenerateOrderID(i)
			t.Logf("Order %d: %s", i, orderID)

			if len(orderID) < 3 {
				t.Errorf("Order ID too short: %s", orderID)
			}
		}
	})

	t.Run("share_code", func(t *testing.T) {
		testData := []int64{123, 456, 789, 999999}
		for _, data := range testData {
			shareCode := GenerateShareCode(BusinessShare, 7*24*time.Hour, data)
			t.Logf("Share code for data %d: %s", data, shareCode)
		}
	})

	t.Run("invite_code", func(t *testing.T) {
		userIDs := []int64{123456789, 987654321, 1, 999999}
		for _, uid := range userIDs {
			inviteCode := GenerateInviteCode(uid)
			t.Logf("User %d invite code: %s", uid, inviteCode)
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
		}

		for _, tc := range testCases {
			cacheKey := GenerateCacheKey(tc.prefix, tc.id)
			t.Logf("Cache key %s:%d -> %s", tc.prefix, tc.id, cacheKey)
		}
	})
}

// TestIDGenerator 测试ID生成器
func TestIDGenerator(t *testing.T) {
	config := IDGeneratorConfig{
		BusinessType: BusinessOrder,
		EnableDate:   true,
		DateBase:     "2024-01-01",
		MachineID:    1,
	}

	gen := NewIDGenerator(config)

	t.Run("generate", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 1000; i++ {
			id := gen.Generate()
			if ids[id] {
				t.Errorf("Duplicate ID generated: %s", id)
			}
			ids[id] = true
		}
	})

	t.Run("parse", func(t *testing.T) {
		id := gen.Generate()
		info, err := gen.Parse(id)
		if err != nil {
			t.Errorf("Failed to parse ID %s: %v", id, err)
		} else {
			t.Logf("Parsed ID %s: Business=%s, Timestamp=%v",
				id, info.Business, info.Timestamp)
		}
	})
}

// TestBatchOperations 测试批量操作
func TestBatchOperations(t *testing.T) {
	t.Run("batch_order_ids", func(t *testing.T) {
		ids := BatchOrderIDs(100)
		if len(ids) != 100 {
			t.Errorf("Expected 100 IDs, got %d", len(ids))
		}

		// 检查唯一性
		idMap := make(map[string]bool)
		for _, id := range ids {
			if idMap[id] {
				t.Errorf("Duplicate ID in batch: %s", id)
			}
			idMap[id] = true
		}
	})

	t.Run("batch_share_codes", func(t *testing.T) {
		data := make([]int64, 50)
		for i := range data {
			data[i] = int64(i + 1000)
		}

		codes := BatchShareCodes(50, data)
		if len(codes) != 50 {
			t.Errorf("Expected 50 codes, got %d", len(codes))
		}
	})
}

// BenchmarkPerformance 性能基准测试
func BenchmarkPerformance(b *testing.B) {
	b.Run("base64_encode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToBase64(123456789)
		}
	})

	b.Run("base64_decode", func(b *testing.B) {
		s := "1Z"
		for i := 0; i < b.N; i++ {
			FromBase64(s)
		}
	})

	b.Run("timestamp_encode", func(b *testing.B) {
		ts := time.Now().Unix()
		for i := 0; i < b.N; i++ {
			ToTimestampShort(ts)
		}
	})

	b.Run("id_generate", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			DefaultOrderGenerator.Generate()
		}
	})
}

// TestInternetScenarios 互联网应用场景测试
func TestInternetScenarios(t *testing.T) {
	t.Run("ecommerce", func(t *testing.T) {
		t.Log("=== 电商场景 ===")

		// 订单号
		orderID := GenerateOrderID(12345)
		t.Logf("订单号: %s", orderID)

		// 支付流水
		paymentID := NewIDGenerator(IDGeneratorConfig{
			BusinessType: BusinessPayment,
			EnableDate:   true,
		}).Generate()
		t.Logf("支付流水号: %s", paymentID)

		// 物流单号
		logisticsID := NewIDGenerator(IDGeneratorConfig{
			BusinessType: BusinessLogistics,
			Prefix:       "SF",
		}).Generate()
		t.Logf("物流单号: %s", logisticsID)
	})

	t.Run("social", func(t *testing.T) {
		t.Log("=== 社交场景 ===")

		// 用户邀请码
		inviteCode := QuickInviteCode(123456789)
		t.Logf("邀请码: %s", inviteCode)

		// 活动分享码
		activityCode := GenerateShareCode(BusinessActivity, 3*24*time.Hour, 10086)
		t.Logf("活动分享码: %s", activityCode)

		// 会话ID
		sessionID := GenerateSessionID(12345, 2*time.Hour)
		t.Logf("会话ID: %s", sessionID)
	})

	t.Run("content", func(t *testing.T) {
		t.Log("=== 内容场景 ===")

		// 文章ID
		articleID := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'C', // Content
			EnableDate:   false,
		}).Generate()
		t.Logf("文章ID: %s", articleID)

		// 评论ID
		commentID := NewIDGenerator(IDGeneratorConfig{
			BusinessType: 'c', // comment
		}).Generate()
		t.Logf("评论ID: %s", commentID)
	})

	t.Run("marketing", func(t *testing.T) {
		t.Log("=== 营销场景 ===")

		// 优惠券码
		couponCode := NewIDGenerator(IDGeneratorConfig{
			BusinessType: BusinessCoupon,
			Prefix:       "CPN",
		}).Generate()
		t.Logf("优惠券码: %s", couponCode)

		// 营销活动码
		promoCode := GenerateShareCode(BusinessMarketing, 30*24*time.Hour, 520)
		t.Logf("促销码: %s", promoCode)
	})
}

// TestEdgeCases 边界情况测试
func TestEdgeCases(t *testing.T) {
	t.Run("zero_values", func(t *testing.T) {
		if ToBase64(0) != "0" {
			t.Error("ToBase64(0) should be '0'")
		}
		if ToTimestampShort(0) != "0" {
			t.Error("ToTimestampShort(0) should be '0'")
		}
	})

	t.Run("large_numbers", func(t *testing.T) {
		large := int64(9223372036854775807) // MaxInt64
		encoded := ToBase64(large)
		decoded, err := FromBase64(encoded)
		if err != nil {
			t.Errorf("Failed to decode large number: %v", err)
		} else if decoded != large {
			t.Errorf("Large number decode mismatch")
		}
	})

	t.Run("negative_numbers", func(t *testing.T) {
		negative := int64(-12345)
		encoded := ToBase64(negative)
		decoded, err := FromBase64(encoded)
		if err != nil {
			t.Errorf("Failed to decode negative number: %v", err)
		} else if decoded != negative {
			t.Errorf("Negative number decode mismatch")
		}
	})
}
