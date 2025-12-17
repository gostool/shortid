package main

import (
	"fmt"
	"time"

	"github.com/gostool/shortid"
)

func main() {
	fmt.Println("Base62 编码解码示例")
	fmt.Println("================")

	// 示例1：编码时间戳
	timestamp := uint64(time.Now().Unix())
	fmt.Printf("\n当前时间戳: %d\n", timestamp)

	encoded := shortid.EncodeBase62(timestamp)
	fmt.Printf("Base62编码: %s\n", encoded)

	decoded, err := shortid.DecodeBase62ToUint(encoded)
	if err != nil {
		fmt.Printf("解码错误: %v\n", err)
	} else {
		fmt.Printf("解码结果: %d\n", decoded)
	}

	// 示例2：编码一系列数字
	fmt.Println("\n数字编码示例:")
	numbers := []uint64{0, 1, 10, 100, 1000, 12345, 999999, 1234567890}
	for _, num := range numbers {
		encoded := shortid.EncodeBase62(num)
		decoded, _ := shortid.DecodeBase62ToUint(encoded)
		fmt.Printf("%12d -> %s -> %d\n", num, encoded, decoded)
	}

	// 示例3：验证Base62字符串
	fmt.Println("\nBase62字符串验证:")
	testStrings := []string{"abcXYZ123", "hello", "abc#123", "", "123"}
	for _, s := range testStrings {
		valid := shortid.IsValidBase62(s)
		fmt.Printf("%s -> %v\n", s, valid)
	}

	// 示例4：计算编码长度
	fmt.Println("\n编码长度示例:")
	for i := uint64(0); i < 100000; i *= 10 {
		if i == 0 {
			i = 1
		}
		length := shortid.Base62Length(i)
		fmt.Printf("%d -> 长度: %d\n", i, length)
	}

	// 示例5：使用其他进制
	fmt.Println("\n其他进制编码示例:")
	num := uint64(123456789)

	base36 := shortid.EncodeBase36(num)
	fmt.Printf("Base36编码: %s\n", base36)

	base58 := shortid.EncodeBase58(num)
	fmt.Printf("Base58编码: %s\n", base58)

	// 使用自定义字符集
	customCharset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	customEncoded := shortid.EncodeWithBase(num, customCharset)
	fmt.Printf("自定义进制编码: %s\n", customEncoded)
}