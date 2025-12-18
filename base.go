package shortid

import (
	"errors"
	"strings"
)

// ============================================================================
// 错误定义
// ============================================================================

var (
	// ErrInvalidInput 无效输入错误
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidNumber 无效数字错误
	ErrInvalidNumber = errors.New("invalid number")

	// ErrInvalidBase62Char 无效Base62字符错误
	ErrInvalidBase62Char = errors.New("invalid base62 character")

	// ErrInvalidChar 无效字符错误
	ErrInvalidChar = errors.New("invalid character")
)

// ============================================================================
// 预定义的其他进制字符集
// ============================================================================

const (
	// 基础字符集组件
	digits    = "0123456789"
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

const (
	// Base36字符集（数字+小写字母）
	Base36Chars = digits + lowercase

	// Base58字符集（Bitcoin风格，去除容易混淆的字符：0, O, I, l）
	// 数字部分：移除 0
	base58Digits = "123456789"
	// 大写字母部分：移除 O, I
	base58Upper = "ABCDEFGHJKLMNPQRSTUVWXYZ"
	// 小写字母部分：移除 l
	base58Lower = "abcdefghijkmnopqrstuvwxyz"
	Base58Chars = base58Digits + base58Upper + base58Lower

	// Base62字符集（数字+小写字母+大写字母）
	base62Chars = digits + lowercase + uppercase
)

// ============================================================================
// 扩展功能：支持任意进制
// ============================================================================

// EncodeWithBase 使用指定字符集进行编码
func EncodeWithBase(num uint64, charset string) string {
	if num == 0 {
		return string(charset[0])
	}

	base := uint64(len(charset))
	var result strings.Builder

	for num > 0 {
		remainder := num % base
		result.WriteByte(charset[remainder])
		num = num / base
	}

	// 反转字符串
	encoded := result.String()
	runes := []rune(encoded)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// DecodeWithBase 使用指定字符集进行解码
func DecodeWithBase(s string, charset string) (uint64, error) {
	if s == "" {
		return 0, ErrInvalidInput
	}

	base := uint64(len(charset))
	var result uint64

	for _, char := range s {
		index := strings.IndexRune(charset, char)
		if index == -1 {
			return 0, ErrInvalidChar
		}

		result = result*base + uint64(index)
	}

	return result, nil
}

// ============================================================================
// Base62 编码解码
// ============================================================================

// EncodeBase62 将无符号整数编码为Base62字符串
func EncodeBase62(num uint64) string {
	return EncodeWithBase(num, base62Chars)
}

// EncodeBase62Int 将uint64数字编码为Base62字符串
func EncodeBase62Int(num uint64) string {
	return EncodeBase62(num)
}

// DecodeBase62 将Base62字符串解码为数字
func DecodeBase62(s string) (uint64, error) {
	return DecodeBase62ToUint(s)
}

// DecodeBase62ToUint 将Base62字符串解码为无符号整数
func DecodeBase62ToUint(s string) (uint64, error) {
	return DecodeWithBase(s, base62Chars)
}

// ============================================================================
// 工具函数
// ============================================================================

// IsValidBase62 检查字符串是否为有效的Base62编码
func IsValidBase62(s string) bool {
	if s == "" {
		return false
	}

	for _, char := range s {
		if !strings.ContainsRune(base62Chars, char) {
			return false
		}
	}

	return true
}

// Base62Length 计算数字编码为Base62后的长度
func Base62Length(num uint64) int {
	if num == 0 {
		return 1
	}

	base := uint64(len(base62Chars))
	length := 0

	for num > 0 {
		num = num / base
		length++
	}

	return length
}

// ============================================================================
// 预定义的其他进制编码函数
// ============================================================================

// EncodeBase36 将数字编码为Base36字符串
func EncodeBase36(num uint64) string {
	return EncodeWithBase(num, Base36Chars)
}

// DecodeBase36 将Base36字符串解码为数字
func DecodeBase36(s string) (uint64, error) {
	return DecodeWithBase(s, Base36Chars)
}

// EncodeBase58 将数字编码为Base58字符串
func EncodeBase58(num uint64) string {
	return EncodeWithBase(num, Base58Chars)
}

// DecodeBase58 将Base58字符串解码为数字
func DecodeBase58(s string) (uint64, error) {
	return DecodeWithBase(s, Base58Chars)
}
