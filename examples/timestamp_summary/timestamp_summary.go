package main

import (
	"fmt"
	"github.com/gostool/shortid"
	"strings"
)

func main() {
	fmt.Println("时间戳编码方案总结")
	fmt.Println(strings.Repeat("=", 60))

	// 测试目标：2025年6月15日
	target := fmt.Sprintf("%d", 1749907200) // 2025-06-15 00:00:00 UTC

	fmt.Println("\n目标时间: 2025-06-15")
	fmt.Println("时间戳:", target)
	fmt.Println("原始长度:", len(target), "位")
	fmt.Println()

	fmt.Println("各种编码方案对比:")
	fmt.Println(strings.Repeat("-", 60))

	// 1. 标准Base62
	standard := shortid.ToBase64(1749907200)
	fmt.Printf("标准Base62:   %s\t\t(%d位) - 基础方案\n", standard, len(standard))

	// 2. V3时间差方案
	v3 := shortid.ToTimestampShort(1749907200)
	fmt.Printf("时间差方案:   %s\t\t(%d位) - 2020年基准\n", v3, len(v3))

	// 3. 动态基准方案
	dynamic := shortid.ToTimestampDynamicV3(1749907200)
	fmt.Printf("动态基准:     %s\t\t(%d位) - 2024年基准\n", dynamic, len(dynamic))

	// 4. 日期编码方案
	date := shortid.EncodeDate(2025, 6, 15)
	fmt.Printf("日期编码:     %s\t\t(%d位) - 最优方案!\n", date, len(date))

	fmt.Println()
	fmt.Println("结论:")
	fmt.Println("1. 使用动态基准（如2024年）可以将时间戳压缩到1-3位")
	fmt.Println("2. 对于2025年的日期，最短只需要1位")
	fmt.Println("3. 适合场景:")
	fmt.Println("   - 临时链接/分享码")
	fmt.Println("   - 近期事件标识")
	fmt.Println("   - 缓存键名")
	fmt.Println("   - 相对时间表达")

	fmt.Println("\n使用示例:")
	fmt.Println("// 编码")
	fmt.Printf(`encoded := shortid.EncodeDate(2025, 6, 15)  // 结果: %s`, date)
	fmt.Println()
	fmt.Println("// 解码（需要自定义解码函数）")
	fmt.Println("date := decodeToDate(encoded)")
}
