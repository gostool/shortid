package main

import (
	"fmt"
	"github.com/gostool/shortid"
	"strings"
)

func main() {
	fmt.Println("62进制与10进制对应关系表")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Println()

	// 字符集
	const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	fmt.Printf("字符集: %s\n\n", chars)

	// 1-62的对应关系
	fmt.Println("1位62进制:")
	fmt.Println("62进制 | 10进制")
	fmt.Println("-------|--------")
	for i := 0; i < 62; i++ {
		fmt.Printf("  %c    | %2d\n", chars[i], i)
	}

	fmt.Println("\n2位62进制（重要数值）:")
	fmt.Println("62进制 | 10进制              | 说明")
	fmt.Println("-------|--------------------|--------------------")
	fmt.Println("  00   | 0                  | 最小值")
	fmt.Println("  01   | 1                  | ")
	fmt.Println("  10   | 62                 | 1*62^1")
	fmt.Println("  11   | 63                 | 1*62^1 + 1")
	fmt.Println("  1Z   | 123                | 1*62 + 61")
	fmt.Println("  20   | 124                | 2*62^1 + 0")
	fmt.Println("  99   | 619                | 9*62 + 61")
	fmt.Println("  Z0   | 3782               | 61*62 + 0")
	fmt.Println("  ZZ   | 3843               | 61*62 + 61")
	fmt.Println(" 100   | 3844               | 1*62^2 + 0 + 0")

	fmt.Println("\n3位62进制（重要数值）:")
	fmt.Println("62进制  | 10进制         | 说明")
	fmt.Println("--------|---------------|-------------------")
	fmt.Println("   000  | 0             | 最小值")
	fmt.Println("   001  | 1             | ")
	fmt.Println("   010  | 62            | ")
	fmt.Println("   100  | 3844          | 1*62^2")
	fmt.Println("   ZZZ  | 238327        | 最大3位数 (62^3 - 1)")
	fmt.Println("  1000  | 238328        | 1*62^3")

	fmt.Println("\n4位62进制（重要数值）:")
	fmt.Println("62进制   | 10进制        | 说明")
	fmt.Println("---------|---------------|-------------------")
	fmt.Println("    0000 | 0             | 最小值")
	fmt.Println("    0001 | 1             | ")
	fmt.Println("    1000 | 3844          | ")
	fmt.Println("   10000 | 14776336      | 1*62^4")
	fmt.Println("   ZZZZ  | 14776335      | 最大4位数 (62^4 - 1)")
	fmt.Println("  10000  | 14776336      | 1*62^4")

	fmt.Println("\n各位数范围总结:")
	fmt.Println("位数 | 最小值     | 最大值        | 可表示范围")
	fmt.Println("-----|------------|---------------|-------------------")
	fmt.Println(" 1   | 0          | 61            | 62 个值")
	fmt.Println(" 2   | 0          | 3843          | 3,844 个值")
	fmt.Println(" 3   | 0          | 238327        | 238,328 个值")
	fmt.Println(" 4   | 0          | 14776335      | 14,776,336 个值")
	fmt.Println(" 5   | 0          | 916132831     | 916,132,832 个值")
	fmt.Println(" 6   | 0          | 56800235583   | 56,800,235,584 个值")
	fmt.Println(" 7   | 0          | 3521614606143 | 3,521,614,606,144 个值")

	fmt.Println("\n实际应用示例:")
	fmt.Println("- 用户ID (1亿): 需要 4-5 位")
	fmt.Println("- 订单号 (万亿): 需要 6-7 位")
	
	// 使用库函数进行编码
	timestamp2100 := int64(4133951999)
	encoded := shortid.ToBase64(timestamp2100)
	fmt.Printf("- 时间戳 (2100年): %d -> %s (%d 位)\n", timestamp2100, encoded, len(encoded))
}

