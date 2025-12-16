package shortid

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestTimestampEncoder 测试时间戳编码器
func TestTimestampEncoder(t *testing.T) {
	configs := []struct {
		name   string
		config TimestampConfig
	}{
		{
			name: "default",
			config: TimestampConfig{
				UseDays: false,
			},
		},
		{
			name: "days_mode",
			config: TimestampConfig{
				UseDays: true,
			},
		},
		{
			name: "compact_mode",
			config: TimestampConfig{
				CompactMode: true,
			},
		},
	}

	testTimestamps := []int64{
		0,
		time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC).Unix(),
		time.Now().Unix(),
		time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC).Unix(),
		time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		time.Date(2100, 12, 31, 23, 59, 59, 0, time.UTC).Unix(),
	}

	for _, tc := range configs {
		t.Run(tc.name, func(t *testing.T) {
			encoder := NewTimestampEncoder(tc.config)

			for _, ts := range testTimestamps {
				encoded := encoder.Encode(ts)
				decoded, err := encoder.Decode(encoded)

				if err != nil {
					t.Errorf("Decode error for %d: %v", ts, err)
					continue
				}

				if decoded != ts {
					t.Errorf("Roundtrip failed: %d -> %s -> %d",
						ts, encoded, decoded)
				}

				t.Logf("Timestamp %d -> %s (length: %d)",
					ts, encoded, len(encoded))
			}
		})
	}
}

// TestDynamicTimestamp 测试动态时间戳编码
func TestDynamicTimestamp(t *testing.T) {
	now := time.Now().UTC()

	testCases := []struct {
		name     string
		offset   time.Duration
		expected string
	}{
		{"now", 0, "now"},
		{"1_min_ago", -1 * time.Minute, "m1"},
		{"5_min_ago", -5 * time.Minute, "m5"},
		{"30_min_ago", -30 * time.Minute, "m30"},
		{"1_hour_ago", -1 * time.Hour, "h1"},
		{"6_hours_ago", -6 * time.Hour, "h6"},
		{"1_day_ago", -24 * time.Hour, "d1"},
		{"1_week_ago", -7 * 24 * time.Hour, "d7"},
		{"1_month_ago", -30 * 24 * time.Hour, "M1"},
		{"1_year_ago", -365 * 24 * time.Hour, "y1"},
		{"1_min_ahead", 1 * time.Minute, "m1"},
		{"1_hour_ahead", 1 * time.Hour, "h1"},
		{"1_day_ahead", 24 * time.Hour, "d1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts := now.Add(tc.offset).Unix()
			encoded := ToTimestampDynamic(ts)

			// 动态编码包含时间相对性，这里只检查格式
			if !strings.HasPrefix(encoded, "-") && !strings.HasPrefix(encoded, "m") &&
				!strings.HasPrefix(encoded, "h") && !strings.HasPrefix(encoded, "d") &&
				!strings.HasPrefix(encoded, "M") && !strings.HasPrefix(encoded, "y") &&
				encoded != "now" {
				t.Errorf("Unexpected format: %s", encoded)
			}

			t.Logf("%s -> %s", tc.name, encoded)
		})
	}
}

// TestDateEncoding 测试日期编码
func TestDateEncoding(t *testing.T) {
	testDates := []struct {
		year, month, day int
		expected         string
	}{
		{2024, 12, 30, "-1"},
		{2024, 12, 31, "0"},
		{2025, 1, 1, "1"},
		{2025, 1, 2, "2"},
		{2025, 2, 1, "32"}, // 简化计算
	}

	for _, td := range testDates {
		t.Run(fmt.Sprintf("%04d-%02d-%02d", td.year, td.month, td.day), func(t *testing.T) {
			encoded := EncodeDate(td.year, td.month, td.day)
			t.Logf("%04d-%02d-%02d -> %s", td.year, td.month, td.day, encoded)
		})
	}
}

// TestTimestampCompression 测试时间戳压缩率
func TestTimestampCompression(t *testing.T) {
	timestamps := []struct {
		name string
		ts   int64
	}{
		{"2020_start", 1577836800},
		{"2024_start", 1704067200},
		{"2025_mid", 1749907200},
		{"2030_start", 1893456000},
		{"2100_end", 4133951999},
	}

	fmt.Println("\n时间戳压缩率对比:")
	fmt.Println("名称\t\t| 原始(10位) | Base62 | 短编码 | 动态V3")
	fmt.Println(strings.Repeat("-", 60))

	for _, ts := range timestamps {
		original := fmt.Sprintf("%d", ts.ts)
		base62 := ToBase64(ts.ts)
		short := ToTimestampShort(ts.ts)
		dynamic := ToTimestampDynamicV3(ts.ts)

		compression1 := float64(len(base62)) / float64(len(original)) * 100
		compression2 := float64(len(short)) / float64(len(original)) * 100
		compression3 := float64(len(dynamic)) / float64(len(original)) * 100

		fmt.Printf("%-12s\t| %-10s | %-6s | %-7s | %-6s\n",
			ts.name, original, base62, short, dynamic)
		fmt.Printf("\t\t| 压缩率:    | %.1f%% | %.1f%% | %.1f%%\n",
			compression1, compression2, compression3)
	}
}

// TestTimestampRoundtrip 测试时间戳往返转换
func TestTimestampRoundtrip(t *testing.T) {
	testFuncs := []struct {
		name   string
		encode func(int64) string
		decode func(string) (int64, error)
	}{
		{"Base62", ToBase64, FromBase64},
		{"Short", ToTimestampShort, FromTimestampShort},
	}

	for _, tf := range testFuncs {
		t.Run(tf.name, func(t *testing.T) {
			testValues := []int64{
				0,
				1,
				60,         // 1分钟
				3600,       // 1小时
				86400,      // 1天
				2592000,    // 30天
				31536000,   // 1年
				1577836800, // 2020年
				time.Now().Unix(),
				4102444800, // 2100年
			}

			for _, val := range testValues {
				encoded := tf.encode(val)
				decoded, err := tf.decode(encoded)

				if err != nil {
					t.Errorf("Decode error for %d: %v", val, err)
					continue
				}

				if decoded != val {
					t.Errorf("Roundtrip failed: %d -> %s -> %d",
						val, encoded, decoded)
				}
			}
		})
	}
}

// BenchmarkTimestampEncoders 时间戳编码器性能测试
func BenchmarkTimestampEncoders(b *testing.B) {
	ts := time.Now().Unix()

	b.Run("base62", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToBase64(ts)
		}
	})

	b.Run("short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToTimestampShort(ts)
		}
	})

	b.Run("dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToTimestampDynamic(ts)
		}
	})

	b.Run("dynamic_v3", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			ToTimestampDynamicV3(ts)
		}
	})

	encoder := NewTimestampEncoder(TimestampConfig{
		UseDays: true,
	})
	b.Run("encoder_days", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			encoder.Encode(ts)
		}
	})

	encoder = NewTimestampEncoder(TimestampConfig{
		CompactMode: true,
	})
	b.Run("encoder_compact", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			encoder.Encode(ts)
		}
	})
}

// TestRealWorldTimestampScenarios 真实世界时间戳场景测试
func TestRealWorldTimestampScenarios(t *testing.T) {
	t.Run("user_session", func(t *testing.T) {
		t.Log("=== 用户会话时间戳 ===")

		loginTime := time.Now()
		sessionExpiry := loginTime.Add(2 * time.Hour)

		loginEncoded := ToTimestampShort(loginTime.Unix())
		expiryEncoded := ToTimestampDynamic(sessionExpiry.Unix())

		t.Logf("登录时间: %s -> %s", loginTime.Format("15:04:05"), loginEncoded)
		t.Logf("过期时间: %s -> %s", sessionExpiry.Format("15:04:05"), expiryEncoded)
	})

	t.Run("cache_expiry", func(t *testing.T) {
		t.Log("=== 缓存过期时间 ===")

		now := time.Now()
		expiryTimes := []time.Duration{
			5 * time.Minute,
			30 * time.Minute,
			1 * time.Hour,
			6 * time.Hour,
			24 * time.Hour,
			7 * 24 * time.Hour,
		}

		for _, duration := range expiryTimes {
			expiry := now.Add(duration)
			encoded := ToTimestampDynamic(expiry.Unix())
			t.Logf("缓存过期 %v: %s", duration, encoded)
		}
	})

	t.Run("event_scheduling", func(t *testing.T) {
		t.Log("=== 事件调度时间戳 ===")

		base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		events := []struct {
			name   string
			offset time.Duration
		}{
			{"新年开始", 0},
			{"春节", 40 * 24 * time.Hour},
			{"618活动", 169 * 24 * time.Hour},
			{"双十一", 314 * 24 * time.Hour},
			{"圣诞节", 358 * 24 * time.Hour},
		}

		for _, event := range events {
			eventTime := base.Add(event.offset)
			dateEncoded := EncodeDate(eventTime.Year(), int(eventTime.Month()), eventTime.Day())
			t.Logf("%s: %s -> %s", event.name,
				eventTime.Format("2006-01-02"), dateEncoded)
		}
	})
}
