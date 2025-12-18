package shortid

import (
	"context"
	"math"
	"testing"
)

// TestGenerateIDPrecision 测试 GenerateID 是否有精度丢失问题
func TestGenerateIDPrecision(t *testing.T) {
	generator, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}
	ctx := context.Background()

	// 测试大量ID，检查是否有精度丢失
	maxDiff := uint64(0)
	totalTests := 10000
	precisionErrors := 0

	for i := 0; i < totalTests; i++ {
		id, b62Str, err := generator.GenerateID(ctx)
		if err != nil {
			t.Fatalf("GenerateID() error = %v", err)
		}

		decoded, err := DecodeBase62ToUint(b62Str)
		if err != nil {
			t.Fatalf("DecodeBase62ToUint() error = %v", err)
		}

		diff := id - decoded
		if diff > maxDiff {
			maxDiff = diff
		}

		if id != decoded {
			precisionErrors++
			if precisionErrors <= 10 {
				t.Logf("精度丢失: id=%d, b62Str=%s, decoded=%d, diff=%d", id, b62Str, decoded, diff)
			}
		}
	}

	if precisionErrors > 0 {
		t.Errorf("发现 %d 个精度丢失问题，最大差值: %d", precisionErrors, maxDiff)
	} else {
		t.Logf("✓ 测试通过: 生成 %d 个ID，无精度丢失", totalTests)
	}
}

// TestBase62PrecisionBoundary 测试 Base62 编码/解码的边界情况
func TestBase62PrecisionBoundary(t *testing.T) {
	// 测试 uint64 的最大值
	maxUint64 := uint64(math.MaxUint64)

	encoded := EncodeBase62(maxUint64)
	decoded, err := DecodeBase62ToUint(encoded)
	if err != nil {
		t.Fatalf("DecodeBase62ToUint() error = %v", err)
	}

	if decoded != maxUint64 {
		t.Errorf("精度丢失: 原始值=%d, 编码=%s, 解码=%d, 差值=%d",
			maxUint64, encoded, decoded, maxUint64-decoded)
	} else {
		t.Logf("✓ uint64 最大值测试通过: %d -> %s -> %d", maxUint64, encoded, decoded)
	}

	// 测试一些大数值
	largeNumbers := []uint64{
		1000000000000000000,  // 10^18
		18446744073709551615, // 2^64 - 1
		18446744073709551614, // 2^64 - 2
		100000000000000000,   // 10^17
		999999999999999999,   // 接近 10^18
	}

	for _, num := range largeNumbers {
		encoded := EncodeBase62(num)
		decoded, err := DecodeBase62ToUint(encoded)
		if err != nil {
			t.Errorf("DecodeBase62ToUint() error for %d: %v", num, err)
			continue
		}

		if decoded != num {
			t.Errorf("精度丢失: %d -> %s -> %d (差值: %d)",
				num, encoded, decoded, num-decoded)
		}
	}
}

// TestBase62Overflow 测试 Base62 解码时是否会发生溢出
func TestBase62Overflow(t *testing.T) {
	// Base62 可以表示的最大值
	// 11位 Base62: 62^11 - 1 = 5520614389124367359
	// 12位 Base62: 62^12 - 1 = 342277092121710775258 (超过 uint64 最大值)

	// 测试 11 位 Base62 的最大值
	max11Digit := uint64(5520614389124367359)
	encoded11 := EncodeBase62(max11Digit)
	decoded11, err := DecodeBase62ToUint(encoded11)
	if err != nil {
		t.Fatalf("DecodeBase62ToUint() error: %v", err)
	}

	if decoded11 != max11Digit {
		t.Errorf("11位最大值精度丢失: %d -> %s -> %d", max11Digit, encoded11, decoded11)
	} else {
		t.Logf("✓ 11位 Base62 最大值测试通过: %d -> %s -> %d", max11Digit, encoded11, decoded11)
	}

	// 测试接近 uint64 最大值的数
	nearMax := uint64(math.MaxUint64)
	encoded := EncodeBase62(nearMax)
	decoded, err := DecodeBase62ToUint(encoded)
	if err != nil {
		t.Fatalf("DecodeBase62ToUint() error: %v", err)
	}

	if decoded != nearMax {
		t.Errorf("接近最大值精度丢失: %d -> %s -> %d (差值: %d)",
			nearMax, encoded, decoded, nearMax-decoded)
	} else {
		t.Logf("✓ 接近 uint64 最大值测试通过: %d -> %s -> %d", nearMax, encoded, decoded)
	}
}
