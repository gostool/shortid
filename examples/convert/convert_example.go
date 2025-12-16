package main

import (
	"fmt"
	"github.com/gostool/shortid"
)

func main() {
	fmt.Println("=== Base62 编码示例 ===")
	fmt.Println("使用字符集: 0-9, a-z, A-Z")
	fmt.Println()

	// 示例1：基本转换
	fmt.Println("1. 基本转换:")
	fmt.Printf("  输入 1  -> 输出 %s\n", shortid.ToBase64(1))
	fmt.Printf("  输入 62 -> 输出 %s\n", shortid.ToBase64(62))
	fmt.Printf("  输入 61 -> 输出 %s\n", shortid.ToBase64(61))
	fmt.Println()

	// 示例2：大数字变短
	fmt.Println("2. 大数字变短:")
	largeNumbers := []int64{
		1000000,
		999999999,
		2147483647,          // int32 max
		9223372036854775807, // int64 max
	}

	for _, num := range largeNumbers {
		encoded := shortid.ToBase64(num)
		fmt.Printf("  %20d -> %s (长度: %d)\n", num, encoded, len(encoded))
	}
	fmt.Println()

	// 示例3：实际应用场景
	fmt.Println("3. 实际应用场景:")

	// 用户ID转换
	userID := int64(123456789)
	userShortID := shortid.ToBase64(userID)
	fmt.Printf("  用户ID: %d -> 短ID: %s\n", userID, userShortID)

	// 订单号转换
	orderID := int64(20240101001)
	orderShortID := shortid.ToBase64(orderID)
	fmt.Printf("  订单号: %d -> 短订单号: %s\n", orderID, orderShortID)

	// URL中的使用
	url := fmt.Sprintf("https://example.com/video/%s", userShortID)
	fmt.Printf("  视频URL: %s\n", url)
	fmt.Println()

	// 示例4：解码验证
	fmt.Println("4. 解码验证:")
	testCases := []struct {
		original int64
	}{
		{1},
		{62},
		{1000},
		{999999},
		{-123},
	}

	for _, tc := range testCases {
		encoded := shortid.ToBase64(tc.original)
		decoded, err := shortid.FromBase64(encoded)
		if err != nil {
			fmt.Printf("  错误: %v\n", err)
		} else if decoded == tc.original {
			fmt.Printf("  ✓ %d -> %s -> %d\n", tc.original, encoded, decoded)
		} else {
			fmt.Printf("  ✗ %d -> %s -> %d (不匹配)\n", tc.original, encoded, decoded)
		}
	}
	fmt.Println()

	// 示例5：长度对比
	fmt.Println("5. 字符串长度对比:")
	fmt.Println("  数字        | 原始长度 | Base62长度 | 压缩比")
	fmt.Println("  ------------|---------|-----------|--------")
	for _, num := range []int64{10, 100, 1000, 100000, 100000000} {
		original := fmt.Sprintf("%d", num)
		encoded := shortid.ToBase64(num)
		ratio := float64(len(encoded)) / float64(len(original))
		fmt.Printf("  %-11d | %-7d | %-9d | %.2f\n", num, len(original), len(encoded), ratio)
	}
}
