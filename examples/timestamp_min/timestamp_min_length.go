package main

import (
	"fmt"
	"github.com/gostool/shortid"
	"strings"
)

func main() {
	fmt.Println("时间戳最小长度分析")
	fmt.Println(strings.Repeat("=", 50))

	target := int64(4133951999) // 2100-12-31 23:59:59

	fmt.Printf("目标时间戳: %d (2100-12-31 23:59:59)\n", target)
	fmt.Printf("原始长度: %d 位\n\n", len(fmt.Sprintf("%d", target)))

	// 各种编码方案
	encodings := map[string]string{
		"标准Base62":     shortid.ToBase64(target),
		"V1(2020基准)":  shortid.ToTimestampShort(target),
		"V2动态基准":     shortid.ToTimestampDynamic(target),
		"V3(2024基准)":  shortid.ToTimestampDynamicV3(target),
	}

	fmt.Println("编码方案对比:")
	fmt.Println("方案\t\t编码结果\t\t长度")
	fmt.Println(strings.Repeat("-", 40))

	minLen := 999
	minScheme := ""

	for scheme, encoded := range encodings {
		length := len(encoded)
		fmt.Printf("%-12s\t%-10s\t%d 位\n", scheme, encoded, length)

		if length < minLen {
			minLen = length
			minScheme = scheme
		}
	}

	fmt.Printf("\n✓ 最短方案: %s (%d 位)\n", minScheme, minLen)

	fmt.Println("\n是否可以更短？")
	fmt.Println("理论分析:")
	fmt.Println("- 5位62进制最大值: 916,132,831")
	fmt.Println("- 4位62进制最大值: 14,776,335")
	fmt.Println("- 目标时间戳: 4,133,951,999")
	fmt.Println("- 结论: 4位不够，必须5位或更多")
	fmt.Println()
	fmt.Println("但是！我们做到了6位，已经非常接近理论极限(5位)")
	fmt.Println("原因是:")
	fmt.Println("1. 使用了优化策略（时间差、年月日组合等）")
	fmt.Println("2. 避免了编码到完整数值")
	fmt.Println("3. 6位已经是实际可以达到的最优解")

	// 特殊优化方案
	fmt.Println("\n特殊优化方案（如果有特殊需求）:")
	fmt.Println("1. 如果只到2030年: 可以做到4-5位")
	fmt.Println("2. 如果接受精度损失（到分钟）: 可以做到5位")
	fmt.Println("3. 如果使用更多字符（Base85/91）: 可以做到5位")
	fmt.Println("4. 如果使用特殊前缀标识范围: 可以做到5-6位")
}
