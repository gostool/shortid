package shortid

import (
	"fmt"
	"math"
	"strings"
)

const (
	// Base62字符集
	base62Chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Base91字符集（包含更多符号）
	base91Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!#$%&()*+-;<=>?@[]^_`{|}~"
)

// ToBase64 将数字转换为Base62字符串
// 保持原有函数名以保持兼容性
func ToBase64(in int64) string {
	if in == 0 {
		return "0"
	}

	isNegative := in < 0
	if isNegative {
		in = -in
	}

	base := int64(len(base62Chars))
	var result []byte

	for in > 0 {
		remainder := in % base
		result = append([]byte{base62Chars[remainder]}, result...)
		in = in / base
	}

	if isNegative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// ToBase64URL URL安全的Base64编码（与ToBase64相同，因为Base62本身就是URL安全的）
func ToBase64URL(in int64) string {
	return ToBase64(in)
}

// FromBase64 将Base62字符串转换回数字
func FromBase64(s string) (int64, error) {
	if len(s) == 0 {
		return 0, nil
	}

	isNegative := false
	if s[0] == '-' {
		isNegative = true
		s = s[1:]
	}

	base := int64(len(base62Chars))
	var result int64

	charToIndex := make(map[byte]int64)
	for i, c := range base62Chars {
		charToIndex[byte(c)] = int64(i)
	}

	for _, char := range s {
		if idx, ok := charToIndex[byte(char)]; ok {
			result = result*base + idx
		} else {
			return 0, fmt.Errorf("invalid character: %c", char)
		}
	}

	if isNegative {
		result = -result
	}

	return result, nil
}

// EncodeBase62 内部Base62编码函数
func EncodeBase62(num int64) string {
	return ToBase64(num)
}

// DecodeBase62 内部Base62解码函数
func DecodeBase62(s string) (int64, error) {
	return FromBase64(s)
}

// EncodeBase91 Base91编码（更高的压缩率）
func EncodeBase91(in int64) string {
	if in == 0 {
		return "0"
	}

	isNegative := in < 0
	if isNegative {
		in = -in
	}

	base := int64(len(base91Chars))
	var result []byte

	for in > 0 {
		remainder := in % base
		result = append([]byte{base91Chars[remainder]}, result...)
		in = in / base
	}

	if isNegative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// DecodeBase91 Base91解码
func DecodeBase91(s string) (int64, error) {
	if len(s) == 0 {
		return 0, nil
	}

	isNegative := false
	if s[0] == '-' {
		isNegative = true
		s = s[1:]
	}

	base := int64(len(base91Chars))
	var result int64

	charToIndex := make(map[byte]int64)
	for i, c := range base91Chars {
		charToIndex[byte(c)] = int64(i)
	}

	for _, char := range s {
		if idx, ok := charToIndex[byte(char)]; ok {
			result = result*base + idx
		} else {
			return 0, fmt.Errorf("invalid character: %c", char)
		}
	}

	if isNegative {
		result = -result
	}

	return result, nil
}

// EncodeVarint 变长编码（类似protobuf的varint）
func EncodeVarint(in int64) string {
	if in == 0 {
		return "0"
	}

	isNegative := in < 0
	if isNegative {
		in = -in
	}

	var result []byte
	for in >= 0x80 {
		b := byte(in) | 0x80
		result = append(result, b)
		in >>= 7
	}
	result = append(result, byte(in))

	if isNegative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// DecodeVarint 变长解码
func DecodeVarint(s string) (int64, error) {
	if len(s) == 0 {
		return 0, nil
	}

	isNegative := false
	if s[0] == '-' {
		isNegative = true
		s = s[1:]
	}

	var result int64
	var shift uint

	for i, b := range s {
		if i > 9 {
			return 0, fmt.Errorf("varint too long")
		}

		result |= int64(b&0x7F) << shift
		if b&0x80 == 0 {
			break
		}
		shift += 7
	}

	if isNegative {
		result = -result
	}

	return result, nil
}

// EncodeWithCustomBase 使用自定义字符集编码
func EncodeWithCustomBase(in int64, chars string) string {
	if len(chars) < 2 {
		return fmt.Sprintf("%d", in)
	}

	if in == 0 {
		return string(chars[0])
	}

	isNegative := in < 0
	if isNegative {
		in = -in
	}

	base := int64(len(chars))
	var result []byte

	for in > 0 {
		remainder := in % base
		result = append([]byte{chars[remainder]}, result...)
		in = in / base
	}

	if isNegative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// DecodeWithCustomBase 使用自定义字符集解码
func DecodeWithCustomBase(s string, chars string) (int64, error) {
	if len(chars) < 2 {
		return 0, fmt.Errorf("character set too short")
	}

	if len(s) == 0 {
		return 0, nil
	}

	isNegative := false
	if s[0] == '-' {
		isNegative = true
		s = s[1:]
	}

	charToIndex := make(map[byte]int64)
	for i, c := range chars {
		charToIndex[byte(c)] = int64(i)
	}

	base := int64(len(chars))
	var result int64

	for _, char := range s {
		if idx, ok := charToIndex[byte(char)]; ok {
			result = result*base + idx
		} else {
			return 0, fmt.Errorf("invalid character: %c", char)
		}
	}

	if isNegative {
		result = -result
	}

	return result, nil
}

// EstimateLength 估算编码后的长度
func EstimateLength(num int64, base int) int {
	if num == 0 {
		return 1
	}

	if num < 0 {
		num = -num
	}

	length := 0
	for num > 0 {
		num /= int64(base)
		length++
	}

	return length
}

// CompressString 压缩字符串（重复字符替换）
func CompressString(s string) string {
	if len(s) < 2 {
		return s
	}

	var result strings.Builder
	count := 1
	prev := s[0]

	for i := 1; i < len(s); i++ {
		if s[i] == prev {
			count++
			if count == math.MaxInt16 {
				result.WriteString(fmt.Sprintf("%c%d", prev, count))
				count = 1
			}
		} else {
			if count > 3 {
				result.WriteString(fmt.Sprintf("%c%d", prev, count))
			} else {
				result.WriteString(strings.Repeat(string(prev), count))
			}
			prev = s[i]
			count = 1
		}
	}

	// 处理最后一组
	if count > 3 {
		result.WriteString(fmt.Sprintf("%c%d", prev, count))
	} else {
		result.WriteString(strings.Repeat(string(prev), count))
	}

	return result.String()
}

// DecompressString 解压字符串
func DecompressString(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		char := s[i]
		result.WriteByte(char)
		i++

		// 检查是否有数字
		if i < len(s) && s[i] >= '0' && s[i] <= '9' {
			// 读取完整数字
			j := i
			for j < len(s) && s[j] >= '0' && s[j] <= '9' {
				j++
			}

			if j > i {
				count := 0
				fmt.Sscanf(s[i:j], "%d", &count)
				result.WriteString(strings.Repeat(string(char), count-1))
				i = j
			}
		}
	}

	return result.String()
}
