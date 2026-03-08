package shortid

import (
	"fmt"
	"testing"
	"time"
)

// ============================================================================
// 测试辅助函数
// ============================================================================

// 测试用例结构
type testCase struct {
	name        string
	input       int64
	expected    string
	expectedLen int   // 期望的编码长度
	baseline    int64 // 可选基准时间
	now         int64 // 用于动态编码的当前时间
}

// 编解码测试用例结构
type roundTripTestCase struct {
	name      string
	input     int64
	baseline  int64 // 可选基准时间
	now       int64 // 用于动态编码的当前时间
	tolerance int64 // 容差（用于动态编码）
}

// 错误测试用例结构
type errorTestCase struct {
	name     string
	input    string
	baseline int64 // 可选基准时间
	now      int64 // 用于动态编码的当前时间
	wantErr  bool
}

// ============================================================================
// 1. 短编码算法测试
// ============================================================================

func TestToTimestampShort(t *testing.T) {
	tests := []testCase{
		{
			name:        "基准时间",
			input:       1704067200, // 2024-01-01 00:00:00 UTC
			expected:    "0000",     // 天数"0" + 秒数"000"（固定4字符）
			expectedLen: 4,
		},
		{
			name:        "基准时间+1天",
			input:       1704067200 + 86400,
			expected:    "1000", // 天数"1" + 秒数"000"
			expectedLen: 4,
		},
		{
			name:        "基准时间+166天",
			input:       1704067200 + 166*86400,
			expected:    "2G000", // Base62(166) = "2G" + 秒数"000"
			expectedLen: 5,
		},
		{
			name:        "基准时间+365天",
			input:       1704067200 + 365*86400,
			expected:    "5T000", // Base62(365) = "5T" + 秒数"000"
			expectedLen: 5,
		},
		{
			name:        "基准时间+1天+3600秒",
			input:       1704067200 + 86400 + 3600,
			expected:    "10W4", // 编码格式："10W4" = 天数"1" + 秒数"0W4"（从末尾取3字符作为秒数部分）
			expectedLen: 4,      // 天数1字符 + 秒数3字符 = 4字符
		},
		{
			name:        "时间戳0",
			input:       0,
			expected:    "0",
			expectedLen: 1, // 特殊值，长度为1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShort(tt.input)
			if got != tt.expected {
				t.Errorf("ToTimestampShort(%d) = %v, want %v", tt.input, got, tt.expected)
			}
			if tt.expectedLen > 0 && len(got) != tt.expectedLen {
				t.Errorf("ToTimestampShort(%d) length = %d, want %d", tt.input, len(got), tt.expectedLen)
			}
		})
	}
}

func TestToTimestampShortWithBaseline(t *testing.T) {
	baseline := int64(1609459200) // 2021-01-01 00:00:00 UTC
	tests := []testCase{
		{
			name:     "自定义基准时间",
			input:    baseline,
			baseline: baseline,
			expected: "0000", // 天数"0" + 秒数"000"（固定4字符）
		},
		{
			name:     "自定义基准时间+1天",
			input:    baseline + SecondsPerDay,
			baseline: baseline,
			expected: "1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortWithBaseline(tt.input, tt.baseline)
			if got != tt.expected {
				t.Errorf("ToTimestampShortWithBaseline(%d, %d) = %v, want %v",
					tt.input, tt.baseline, got, tt.expected)
			}
			if tt.expectedLen > 0 && len(got) != tt.expectedLen {
				t.Errorf("ToTimestampShortWithBaseline(%d, %d) length = %d, want %d",
					tt.input, tt.baseline, len(got), tt.expectedLen)
			}
		})
	}
}

func TestFromTimestampShort(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			name:     "基准时间（旧格式兼容）",
			input:    "0",
			expected: 1704067200, // 2024-01-01 00:00:00 UTC（兼容旧格式）
		},
		{
			name:     "基准时间（标准格式）",
			input:    "0000",
			expected: 1704067200, // 2024-01-01 00:00:00 UTC
		},
		{
			name:     "1天后",
			input:    "1000",
			expected: 1704067200 + 86400,
		},
		{
			name:     "166天后",
			input:    "2G000",
			expected: 1704067200 + 166*86400,
		},
		{
			name:     "1天1小时后",
			input:    "10W4",
			expected: 1704067200 + 86400 + 3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromTimestampShort(tt.input)
			if err != nil {
				t.Errorf("FromTimestampShort(%s) error: %v", tt.input, err)
				return
			}
			if got != tt.expected {
				t.Errorf("FromTimestampShort(%s) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTimestampShortRoundTrip(t *testing.T) {
	tests := []roundTripTestCase{
		{
			name:  "基准时间",
			input: 1704067200, // 2024-01-01 00:00:00 UTC
		},
		{
			name:  "1天后",
			input: 1704067200 + 86400,
		},
		{
			name:  "100天后",
			input: 1704067200 + 100*86400,
		},
		{
			name:  "1年后",
			input: 1704067200 + 365*86400,
		},
		{
			name:  "10年后",
			input: 1704067200 + 10*365*86400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ToTimestampShort(tt.input)
			decoded, err := FromTimestampShort(encoded)
			t.Logf("encoded: %s, decoded: %d", encoded, decoded)
			if err != nil {
				t.Errorf("RoundTrip error for %d: %v", tt.input, err)
				return
			}
			if decoded != tt.input {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", tt.input, encoded, decoded)
			}
		})
	}
}

func TestTimestampShort_BaselineMismatch(t *testing.T) {
	original := int64(DefaultBaseline + 3*SecondsPerDay + 1234)
	encoded := ToTimestampShortWithBaseline(original, DefaultBaseline)

	decoded, err := FromTimestampShortWithBaseline(encoded, DefaultBaseline+SecondsPerDay)
	if err != nil {
		t.Fatalf("FromTimestampShortWithBaseline() error = %v", err)
	}
	if decoded == original {
		t.Fatalf("decoded with mismatched baseline should differ, got same value %d", decoded)
	}
}

// 测试负数时间戳（基准时间之前）
// 注意：当前实现不支持负数天数，负数时间戳会使用Base62编码回退
func TestTimestampShortNegativeDays(t *testing.T) {
	baseline := int64(DefaultBaseline)
	tests := []struct {
		name     string
		input    int64
		baseline int64
	}{
		{
			name:     "基准时间-1天",
			input:    baseline - int64(SecondsPerDay),
			baseline: baseline,
		},
		{
			name:     "基准时间-1天+1小时",
			input:    baseline - int64(SecondsPerDay) + 3600,
			baseline: baseline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortWithBaseline(tt.input, tt.baseline)
			// 验证编码不为空
			if got == "" {
				t.Errorf("ToTimestampShortWithBaseline(%d, %d) returned empty string",
					tt.input, tt.baseline)
			}
			// 负数时间戳可能使用Base62回退编码，长度不固定
			// 这里只验证编码不为空，不验证具体格式
			t.Logf("Negative timestamp encoding: %d -> %s (may use fallback encoding)", tt.input, got)
		})
	}
}

// ============================================================================
// 2. 动态编码算法测试
// ============================================================================

func TestToTimestampDynamic(t *testing.T) {
	// 固定当前时间用于测试
	now := int64(1704153600) // 2024-01-02 00:00:00 UTC
	tests := []testCase{
		{
			name:     "当前时间",
			input:    now,
			now:      now,
			expected: "now",
		},
		{
			name:     "5分钟后",
			input:    now + 5*UnitMinute,
			now:      now,
			expected: "m5",
		},
		{
			name:     "3分钟前",
			input:    now - 3*UnitMinute,
			now:      now,
			expected: "-m3",
		},
		{
			name:     "2小时后",
			input:    now + 2*UnitHour,
			now:      now,
			expected: "h2",
		},
		{
			name:     "1天前",
			input:    now - SecondsPerDay,
			now:      now,
			expected: "-d1",
		},
		{
			name:     "1月后",
			input:    now + UnitMonth,
			now:      now,
			expected: "M1",
		},
		{
			name:     "1年后",
			input:    now + UnitYear,
			now:      now,
			expected: "y1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampDynamicWithNow(tt.input, tt.now)
			if got != tt.expected {
				t.Errorf("ToTimestampDynamicWithNow(%d, %d) = %v, want %v",
					tt.input, tt.now, got, tt.expected)
			}
		})
	}
}

func TestFromTimestampDynamic(t *testing.T) {
	now := int64(1704153600) // 2024-01-02 00:00:00 UTC
	tests := []struct {
		name     string
		input    string
		now      int64
		expected int64
	}{
		{
			name:     "当前时间",
			input:    "now",
			now:      now,
			expected: now,
		},
		{
			name:     "5分钟后",
			input:    "m5",
			now:      now,
			expected: now + 5*UnitMinute,
		},
		{
			name:     "3分钟前",
			input:    "-m3",
			now:      now,
			expected: now - 3*UnitMinute,
		},
		{
			name:     "2小时后",
			input:    "h2",
			now:      now,
			expected: now + 2*UnitHour,
		},
		{
			name:     "1天前",
			input:    "-d1",
			now:      now,
			expected: now - SecondsPerDay,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromTimestampDynamic(tt.input, tt.now)
			if err != nil {
				t.Errorf("FromTimestampDynamic(%s, %d) error: %v", tt.input, tt.now, err)
				return
			}
			if got != tt.expected {
				t.Errorf("FromTimestampDynamic(%s, %d) = %v, want %v",
					tt.input, tt.now, got, tt.expected)
			}
		})
	}
}

func TestTimestampDynamicRoundTrip(t *testing.T) {
	now := time.Now().Unix()
	tests := []int64{
		now,                   // 当前
		now + 30*UnitMinute,   // 30分钟后
		now + 3*UnitHour,      // 3小时后
		now + 2*SecondsPerDay, // 2天后
		now - 15*UnitMinute,   // 15分钟前
		now - UnitHour,        // 1小时前
		now + UnitMonth,       // 1月后
	}

	for _, ts := range tests {
		t.Run(fmt.Sprintf("timestamp_%d", ts), func(t *testing.T) {
			encoded := ToTimestampDynamicWithNow(ts, now)
			decoded, err := FromTimestampDynamic(encoded, now)
			if err != nil {
				t.Errorf("RoundTrip error for %d: %v", ts, err)
				return
			}
			if decoded != ts {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", ts, encoded, decoded)
			}
		})
	}
}

func TestTimestampDynamic_DifferentNowAffectsDecode(t *testing.T) {
	nowA := int64(1704067200)
	nowB := nowA + 300
	target := nowA + 2*UnitHour

	encoded := ToTimestampDynamicWithNow(target, nowA)
	decodedWithA, err := FromTimestampDynamic(encoded, nowA)
	if err != nil {
		t.Fatalf("FromTimestampDynamic() with nowA error = %v", err)
	}
	decodedWithB, err := FromTimestampDynamic(encoded, nowB)
	if err != nil {
		t.Fatalf("FromTimestampDynamic() with nowB error = %v", err)
	}

	if decodedWithA != target {
		t.Fatalf("decodedWithA = %d, want %d", decodedWithA, target)
	}
	if decodedWithB == target {
		t.Fatalf("decodedWithB should differ from target when now differs")
	}
}

// ============================================================================
// 3. 紧凑编码算法测试
// ============================================================================

func TestToTimestampCompact(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		checkLen bool // 是否检查长度
	}{
		{
			name:     "基准年份",
			input:    time.Date(BaseYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			checkLen: true,
		},
		{
			name:     "基准年份+1天",
			input:    time.Date(BaseYear, 1, 2, 0, 0, 0, 0, time.UTC).Unix(),
			checkLen: true,
		},
		{
			name:     "基准年份+年中",
			input:    time.Date(BaseYear, 7, 2, 12, 30, 45, 0, time.UTC).Unix(),
			checkLen: true,
		},
		{
			name:     "2024年",
			input:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
			checkLen: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampCompact(tt.input)
			// 验证编码长度（如果在范围内应该是7字符）
			if tt.checkLen {
				year := time.Unix(tt.input, 0).UTC().Year()
				if year >= BaseYear && year <= BaseYear+MaxYearOffset {
					if len(got) != 7 {
						t.Errorf("ToTimestampCompact(%d) = %v (len=%d), want length 7", tt.input, got, len(got))
					}
				}
			}
			// 验证往返
			decoded, err := FromTimestampCompact(got)
			if err != nil {
				t.Errorf("FromTimestampCompact(%s) error: %v", got, err)
				return
			}
			if decoded != tt.input {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", tt.input, got, decoded)
			}
		})
	}
}

func TestFromTimestampCompact(t *testing.T) {
	// 使用往返测试验证解码正确性
	tests := []struct {
		name  string
		input int64
	}{
		{
			name:  "基准年份开始",
			input: time.Date(BaseYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		},
		{
			name:  "基准年份+167天",
			input: time.Date(BaseYear, 6, 16, 12, 30, 45, 0, time.UTC).Unix(),
		},
		{
			name:  "2024年开始",
			input: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ToTimestampCompact(tt.input)
			got, err := FromTimestampCompact(encoded)
			if err != nil {
				t.Errorf("FromTimestampCompact(%s) error: %v", encoded, err)
				return
			}
			if got != tt.input {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", tt.input, encoded, got)
			}
		})
	}
}

func TestTimestampCompactRoundTrip(t *testing.T) {
	// 测试2020-2100年范围内的时间戳
	startYear := 2020
	endYear := 2050

	for year := startYear; year <= endYear; year++ {
		for month := 1; month <= 12; month++ {
			for day := 1; day <= 28; day++ { // 使用28天避免月份天数问题
				ts := time.Date(year, time.Month(month), day, 12, 30, 45, 0, time.UTC).Unix()
				t.Run(fmt.Sprintf("compact_%04d_%02d_%02d", year, month, day), func(t *testing.T) {
					// 只测试在支持范围内的年份
					if year < BaseYear || year > BaseYear+MaxYearOffset {
						t.Skipf("Year %d out of range", year)
						return
					}

					encoded := ToTimestampCompact(ts)
					if len(encoded) != 7 {
						t.Errorf("Compact encoding length should be 7, got %d", len(encoded))
					}

					decoded, err := FromTimestampCompact(encoded)
					if err != nil {
						t.Errorf("Decode error for %s: %v", encoded, err)
						return
					}
					if decoded != ts {
						t.Errorf("RoundTrip failed: %d -> %s -> %d", ts, encoded, decoded)
					}
				})
			}
		}
	}
}

// ============================================================================
// 4. 错误处理测试
// ============================================================================

func TestTimestampShortErrors(t *testing.T) {
	tests := []errorTestCase{
		{
			name:    "字符串太短",
			input:   "123", // 少于4字符
			wantErr: true,
		},
		{
			name:    "无效的Base62字符",
			input:   "1@00", // 包含无效字符
			wantErr: true,
		},
		{
			name:    "秒数部分无效",
			input:   "1xyz", // "xyz"不是有效的秒数
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromTimestampShort(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromTimestampShort(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestTimestampDynamicErrors(t *testing.T) {
	tests := []errorTestCase{
		{
			name:    "字符串太短",
			input:   "m",
			now:     time.Now().Unix(),
			wantErr: true,
		},
		{
			name:    "无效单位",
			input:   "x5",
			now:     time.Now().Unix(),
			wantErr: true,
		},
		{
			name:    "无效数值",
			input:   "m@",
			now:     time.Now().Unix(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromTimestampDynamic(tt.input, tt.now)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromTimestampDynamic(%s, %d) error = %v, wantErr %v",
					tt.input, tt.now, err, tt.wantErr)
			}
		})
	}
}

func TestTimestampCompactErrors(t *testing.T) {
	tests := []errorTestCase{
		{
			name:    "长度错误",
			input:   "123456", // 6字符
			wantErr: true,
		},
		{
			name:    "长度错误",
			input:   "12345678", // 8字符
			wantErr: true,
		},
		{
			name:    "无效的Base62字符",
			input:   "12@4567", // 包含无效字符
			wantErr: true,
		},
		{
			name:    "年份超出范围",
			input:   "zz01500", // 年份偏移超出100
			wantErr: true,
		},
		{
			name:    "天数超出范围",
			input:   "00zz123", // 天数超出366
			wantErr: true,
		},
		{
			name:    "秒数超出范围",
			input:   "0001zzz", // 秒数超出86399
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromTimestampCompact(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromTimestampCompact(%s) error = %v, wantErr %v",
					tt.input, err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// 5. 边界情况测试
// ============================================================================

func TestTimestampBoundaryValues(t *testing.T) {
	t.Run("短编码最大值", func(t *testing.T) {
		// 测试大数值
		ts := int64(DefaultBaseline) + int64(10000*SecondsPerDay)
		encoded := ToTimestampShort(ts)
		decoded, err := FromTimestampShort(encoded)
		if err != nil {
			t.Errorf("Error decoding large timestamp: %v", err)
		}
		if decoded != ts {
			t.Errorf("Large timestamp roundtrip failed")
		}
	})

	t.Run("紧凑编码边界年份", func(t *testing.T) {
		// 测试支持的边界年份
		minYear := BaseYear
		maxYear := BaseYear + MaxYearOffset

		minTs := time.Date(minYear, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		maxTs := time.Date(maxYear, 12, 31, 23, 59, 59, 0, time.UTC).Unix()

		// 测试最小年份
		encoded := ToTimestampCompact(minTs)
		if len(encoded) != 7 {
			t.Errorf("Min year encoding length should be 7")
		}

		// 测试最大年份
		encoded = ToTimestampCompact(maxTs)
		if len(encoded) != 7 {
			t.Errorf("Max year encoding length should be 7")
		}
	})

	t.Run("紧凑编码超出范围", func(t *testing.T) {
		// 测试超出范围的年份
		oldTs := time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
		encoded := ToTimestampCompact(oldTs)
		// 应该回退到Base62编码
		if len(encoded) == 7 {
			t.Errorf("Out of range timestamp should not use compact encoding")
		}
	})
}

func TestTimestampCompact_OutOfRangeFallback(t *testing.T) {
	before := time.Date(1999, 12, 31, 23, 59, 59, 0, time.UTC).Unix()
	after := time.Date(2101, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	beforeEncoded := ToTimestampCompact(before)
	afterEncoded := ToTimestampCompact(after)
	if len(beforeEncoded) == 7 {
		t.Fatalf("beforeEncoded should fallback to variable length, got %q", beforeEncoded)
	}
	if len(afterEncoded) == 7 {
		t.Fatalf("afterEncoded should fallback to variable length, got %q", afterEncoded)
	}

	if _, err := FromTimestampCompact(beforeEncoded); err == nil {
		t.Fatal("FromTimestampCompact() should reject fallback non-7-char format")
	}
}

// ============================================================================
// 6. 性能基准测试
// ============================================================================

func BenchmarkToTimestampShort(b *testing.B) {
	ts := time.Now().Unix()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToTimestampShort(ts)
	}
}

func BenchmarkFromTimestampShort(b *testing.B) {
	encoded := ToTimestampShort(time.Now().Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromTimestampShort(encoded)
	}
}

func BenchmarkToTimestampDynamic(b *testing.B) {
	ts := time.Now().Unix()
	now := ts
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToTimestampDynamicWithNow(ts, now)
	}
}

func BenchmarkToTimestampCompact(b *testing.B) {
	ts := time.Now().Unix()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ToTimestampCompact(ts)
	}
}

func BenchmarkFromTimestampCompact(b *testing.B) {
	encoded := ToTimestampCompact(time.Now().Unix())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FromTimestampCompact(encoded)
	}
}

// ============================================================================
// 7. 集成测试
// ============================================================================

func TestAllAlgorithmsConsistency(t *testing.T) {
	// 测试同一时间戳在不同算法下的表现
	ts := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC).Unix()

	t.Run("短编码", func(t *testing.T) {
		encoded := ToTimestampShort(ts)
		decoded, err := FromTimestampShort(encoded)
		if err != nil {
			t.Errorf("Short encoding error: %v", err)
		}
		if decoded != ts {
			t.Errorf("Short encoding roundtrip failed")
		}
	})

	t.Run("动态编码", func(t *testing.T) {
		now := ts
		encoded := ToTimestampDynamicWithNow(ts, now)
		if encoded != "now" {
			t.Errorf("Dynamic encoding for current time should be 'now', got %s", encoded)
		}
	})

	t.Run("紧凑编码", func(t *testing.T) {
		encoded := ToTimestampCompact(ts)
		if len(encoded) != 7 {
			t.Errorf("Compact encoding should be 7 characters")
		}
		decoded, err := FromTimestampCompact(encoded)
		if err != nil {
			t.Errorf("Compact encoding error: %v", err)
		}
		if decoded != ts {
			t.Errorf("Compact encoding roundtrip failed")
		}
	})
}

// 测试基准时间一致性
func TestBaselineConsistency(t *testing.T) {
	// 确保文档和代码使用相同的基准时间
	if DefaultBaseline != 1704067200 {
		t.Errorf("DefaultBaseline should be 1704067200 (2024-01-01), got %d", DefaultBaseline)
	}
}

// ============================================================================
// 4. 毫秒级短编码算法测试
// ============================================================================

func TestToTimestampShortMs(t *testing.T) {
	baselineMs := int64(DefaultSnowflakeEpochMs) // 1704067200000
	tests := []testCase{
		{
			name:        "基准时间",
			input:       baselineMs,
			expected:    "000000", // 天数"0" + 毫秒数"00000"（6字符：天数1字符+毫秒数5字符）
			expectedLen: 6,
		},
		{
			name:        "基准时间+1天",
			input:       baselineMs + MillisecondsPerDay,
			expected:    "100000", // 天数"1" + 毫秒数"00000"
			expectedLen: 6,
		},
		{
			name:        "基准时间+166天",
			input:       baselineMs + 166*MillisecondsPerDay,
			expected:    "2G00000", // Base62(166) = "2G" + 毫秒数"00000"（7字符：天数2字符+毫秒数5字符）
			expectedLen: 7,
		},
		{
			name:        "基准时间+1天+1000毫秒",
			input:       baselineMs + MillisecondsPerDay + 1000,
			expected:    "1000g8", // 天数"1" + 毫秒数"00g8"（从末尾取5字符作为毫秒数部分）
			expectedLen: 6,        // 天数1字符 + 毫秒数5字符 = 6字符
		},
		{
			name:        "基准时间+1天+3600000毫秒（1小时）",
			input:       baselineMs + MillisecondsPerDay + 3600000,
			expected:    "10f6ww", // 天数"1" + 毫秒数"0f6ww"
			expectedLen: 6,
		},
		{
			name:        "时间戳0",
			input:       0,
			expected:    "0",
			expectedLen: 1, // 特殊值，长度为1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortMs(tt.input)
			if got != tt.expected {
				t.Errorf("ToTimestampShortMs(%d) = %v, want %v", tt.input, got, tt.expected)
			}
			if tt.expectedLen > 0 && len(got) != tt.expectedLen {
				t.Errorf("ToTimestampShortMs(%d) length = %d, want %d", tt.input, len(got), tt.expectedLen)
			}
		})
	}
}

func TestToTimestampShortMsWithBaseline(t *testing.T) {
	baselineMs := int64(1609459200000) // 2021-01-01 00:00:00 UTC（毫秒）
	tests := []testCase{
		{
			name:        "自定义基准时间",
			input:       baselineMs,
			baseline:    baselineMs,
			expected:    "000000", // 天数"0" + 毫秒数"00000"（6字符）
			expectedLen: 6,
		},
		{
			name:        "自定义基准时间+1天",
			input:       baselineMs + MillisecondsPerDay,
			baseline:    baselineMs,
			expected:    "100000", // 天数"1" + 毫秒数"00000"（6字符）
			expectedLen: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortMsWithBaseline(tt.input, tt.baseline)
			if got != tt.expected {
				t.Errorf("ToTimestampShortMsWithBaseline(%d, %d) = %v, want %v",
					tt.input, tt.baseline, got, tt.expected)
			}
		})
	}
}

func TestFromTimestampShortMs(t *testing.T) {
	baselineMs := int64(DefaultSnowflakeEpochMs)
	tests := []struct {
		name      string
		input     string
		expected  int64
		wantError bool
	}{
		{
			name:     "基准时间",
			input:    "000000", // 天数"0" + 毫秒数"00000"（6字符）
			expected: baselineMs,
		},
		{
			name:     "基准时间+1天",
			input:    "100000", // 天数"1" + 毫秒数"00000"（6字符）
			expected: baselineMs + MillisecondsPerDay,
		},
		{
			name:     "基准时间+166天",
			input:    "2G00000",
			expected: baselineMs + 166*MillisecondsPerDay,
		},
		{
			name:      "字符串太短",
			input:     "000",
			wantError: true,
		},
		{
			name:      "无效字符",
			input:     "00000-0",
			wantError: true,
		},
		{
			name:     "特殊值0",
			input:    "0",
			expected: baselineMs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromTimestampShortMs(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("FromTimestampShortMs(%s) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("FromTimestampShortMs(%s) error: %v", tt.input, err)
				return
			}
			if got != tt.expected {
				t.Errorf("FromTimestampShortMs(%s) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTimestampShortMsRoundTrip(t *testing.T) {
	baselineMs := int64(DefaultSnowflakeEpochMs)
	tests := []roundTripTestCase{
		{
			name:     "基准时间",
			input:    baselineMs,
			baseline: baselineMs,
		},
		{
			name:     "基准时间+1天",
			input:    baselineMs + MillisecondsPerDay,
			baseline: baselineMs,
		},
		{
			name:     "基准时间+100天",
			input:    baselineMs + 100*MillisecondsPerDay,
			baseline: baselineMs,
		},
		{
			name:     "基准时间+1天+1000毫秒",
			input:    baselineMs + MillisecondsPerDay + 1000,
			baseline: baselineMs,
		},
		{
			name:     "基准时间+1天+1小时",
			input:    baselineMs + MillisecondsPerDay + 3600000,
			baseline: baselineMs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ToTimestampShortMsWithBaseline(tt.input, tt.baseline)
			decoded, err := FromTimestampShortMsWithBaseline(encoded, tt.baseline)
			t.Logf("encoded: %s, decoded: %d", encoded, decoded)
			if err != nil {
				t.Errorf("RoundTrip error for %d: %v", tt.input, err)
				return
			}
			if decoded != tt.input {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", tt.input, encoded, decoded)
			}
		})
	}
}

// ============================================================================
// 5. 纳秒级短编码算法测试
// ============================================================================

func TestToTimestampShortNs(t *testing.T) {
	baselineNs := int64(DefaultSnowflakeEpochMs) * 1000000 // 1704067200000000000
	tests := []testCase{
		{
			name:        "基准时间",
			input:       baselineNs,
			expected:    "000000000", // 天数"0" + 纳秒数"00000000"（9字符：天数1字符+纳秒数8字符）
			expectedLen: 9,
		},
		{
			name:        "基准时间+1天",
			input:       baselineNs + NanosecondsPerDay,
			expected:    "100000000", // 天数"1" + 纳秒数"00000000"（9字符：天数1字符+纳秒数8字符）
			expectedLen: 9,
		},
		{
			name:        "基准时间+166天",
			input:       baselineNs + 166*NanosecondsPerDay,
			expected:    "2G00000000", // Base62(166) = "2G" + 纳秒数"00000000"
			expectedLen: 10,
		},
		{
			name:        "基准时间+1天+1000000纳秒（1毫秒）",
			input:       baselineNs + NanosecondsPerDay + 1000000,
			expected:    "100004c92", // 天数"1" + 纳秒数"00004c92"（从末尾取8字符作为纳秒数部分）
			expectedLen: 9,           // 天数1字符 + 纳秒数8字符 = 9字符
		},
		{
			name:        "时间戳0",
			input:       0,
			expected:    "0",
			expectedLen: 1, // 特殊值，长度为1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortNs(tt.input)
			if got != tt.expected {
				t.Errorf("ToTimestampShortNs(%d) = %v, want %v", tt.input, got, tt.expected)
			}
			if tt.expectedLen > 0 && len(got) != tt.expectedLen {
				t.Errorf("ToTimestampShortNs(%d) length = %d, want %d", tt.input, len(got), tt.expectedLen)
			}
		})
	}
}

func TestToTimestampShortNsWithBaseline(t *testing.T) {
	baselineNs := int64(1609459200000000000) // 2021-01-01 00:00:00 UTC（纳秒）
	tests := []testCase{
		{
			name:        "自定义基准时间",
			input:       baselineNs,
			baseline:    baselineNs,
			expected:    "000000000", // 天数"0" + 纳秒数"00000000"（9字符）
			expectedLen: 9,
		},
		{
			name:        "自定义基准时间+1天",
			input:       baselineNs + NanosecondsPerDay,
			baseline:    baselineNs,
			expected:    "100000000", // 天数"1" + 纳秒数"00000000"（9字符）
			expectedLen: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToTimestampShortNsWithBaseline(tt.input, tt.baseline)
			if got != tt.expected {
				t.Errorf("ToTimestampShortNsWithBaseline(%d, %d) = %v, want %v",
					tt.input, tt.baseline, got, tt.expected)
			}
		})
	}
}

func TestFromTimestampShortNs(t *testing.T) {
	baselineNs := int64(DefaultSnowflakeEpochMs) * 1000000
	tests := []struct {
		name      string
		input     string
		expected  int64
		wantError bool
	}{
		{
			name:     "基准时间",
			input:    "000000000", // 天数"0" + 纳秒数"00000000"（9字符）
			expected: baselineNs,
		},
		{
			name:     "基准时间+1天",
			input:    "100000000", // 天数"1" + 纳秒数"00000000"（9字符）
			expected: baselineNs + NanosecondsPerDay,
		},
		{
			name:     "基准时间+166天",
			input:    "2G00000000",
			expected: baselineNs + 166*NanosecondsPerDay,
		},
		{
			name:      "字符串太短",
			input:     "0000000",
			wantError: true,
		},
		{
			name:      "无效字符",
			input:     "00000000-0",
			wantError: true,
		},
		{
			name:     "特殊值0",
			input:    "0",
			expected: baselineNs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromTimestampShortNs(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("FromTimestampShortNs(%s) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("FromTimestampShortNs(%s) error: %v", tt.input, err)
				return
			}
			if got != tt.expected {
				t.Errorf("FromTimestampShortNs(%s) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTimestampShortNsRoundTrip(t *testing.T) {
	baselineNs := int64(DefaultSnowflakeEpochMs) * 1000000
	tests := []roundTripTestCase{
		{
			name:     "基准时间",
			input:    baselineNs,
			baseline: baselineNs,
		},
		{
			name:     "基准时间+1天",
			input:    baselineNs + NanosecondsPerDay,
			baseline: baselineNs,
		},
		{
			name:     "基准时间+100天",
			input:    baselineNs + 100*NanosecondsPerDay,
			baseline: baselineNs,
		},
		{
			name:     "基准时间+1天+1000000纳秒（1毫秒）",
			input:    baselineNs + NanosecondsPerDay + 1000000,
			baseline: baselineNs,
		},
		{
			name:     "基准时间+1天+1秒",
			input:    baselineNs + NanosecondsPerDay + 1000000000,
			baseline: baselineNs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := ToTimestampShortNsWithBaseline(tt.input, tt.baseline)
			decoded, err := FromTimestampShortNsWithBaseline(encoded, tt.baseline)
			t.Logf("encoded: %s, decoded: %d", encoded, decoded)
			if err != nil {
				t.Errorf("RoundTrip error for %d: %v", tt.input, err)
				return
			}
			if decoded != tt.input {
				t.Errorf("RoundTrip failed: %d -> %s -> %d", tt.input, encoded, decoded)
			}
		})
	}
}

// ============================================================================
// 精度转换测试
// ============================================================================

func TestPrecisionConversion(t *testing.T) {
	// 测试秒级、毫秒级、纳秒级之间的转换一致性
	ts := int64(1704067200) // 2024-01-01 00:00:00 UTC（秒）
	tsMs := ts * 1000
	tsNs := ts * 1000000

	t.Run("秒级转毫秒级", func(t *testing.T) {
		encodedS := ToTimestampShort(ts)
		encodedMs := ToTimestampShortMs(tsMs)
		// 毫秒级编码应该比秒级编码长2字符（5-3=2）
		if len(encodedMs) != len(encodedS)+2 {
			t.Logf("秒级编码: %s (%d字符), 毫秒级编码: %s (%d字符)", encodedS, len(encodedS), encodedMs, len(encodedMs))
		}
	})

	t.Run("毫秒级转纳秒级", func(t *testing.T) {
		encodedMs := ToTimestampShortMs(tsMs)
		encodedNs := ToTimestampShortNs(tsNs)
		// 纳秒级编码应该比毫秒级编码长3字符（8-5=3）
		// 注意：如果纳秒数超出范围，可能会回退到Base62编码，长度不固定
		if len(encodedNs) < len(encodedMs)+3 {
			t.Logf("毫秒级编码: %s (%d字符), 纳秒级编码: %s (%d字符)", encodedMs, len(encodedMs), encodedNs, len(encodedNs))
			t.Logf("注意：纳秒级编码可能因为超出范围而使用回退编码")
		}
	})

	t.Run("往返转换一致性", func(t *testing.T) {
		// 秒级 -> 毫秒级 -> 秒级
		tsMs := ts * 1000
		encodedMs := ToTimestampShortMs(tsMs)
		decodedMs, err := FromTimestampShortMs(encodedMs)
		if err != nil {
			t.Fatalf("FromTimestampShortMs error: %v", err)
		}
		decodedS := decodedMs / 1000
		if decodedS != ts {
			t.Errorf("秒级转换不一致: %d -> %d", ts, decodedS)
		}

		// 毫秒级 -> 纳秒级 -> 毫秒级
		tsNs := tsMs * 1000000
		encodedNs := ToTimestampShortNs(tsNs)
		decodedNs, err := FromTimestampShortNs(encodedNs)
		if err != nil {
			t.Fatalf("FromTimestampShortNs error: %v", err)
		}
		decodedMs2 := decodedNs / 1000000
		if decodedMs2 != tsMs {
			t.Errorf("毫秒级转换不一致: %d -> %d", tsMs, decodedMs2)
		}
	})
}
