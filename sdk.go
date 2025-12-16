package shortid

import (
	"time"
)

// SDK 主要入口，提供简化的API

// SimpleEncoder 简单编码器
type SimpleEncoder struct{}

// NewSimpleEncoder 创建简单编码器
func NewSimpleEncoder() *SimpleEncoder {
	return &SimpleEncoder{}
}

// EncodeID 编码ID
func (e *SimpleEncoder) EncodeID(id int64) string {
	return ToBase64(id)
}

// DecodeID 解码ID
func (e *SimpleEncoder) DecodeID(s string) (int64, error) {
	return FromBase64(s)
}

// EncodeTimestamp 编码时间戳
func (e *SimpleEncoder) EncodeTimestamp(ts int64) string {
	return ToTimestampShort(ts)
}

// DecodeTimestamp 解码时间戳
func (e *SimpleEncoder) DecodeTimestamp(s string) (int64, error) {
	return FromTimestampShort(s)
}

// EncodeDate 编码日期
func (e *SimpleEncoder) EncodeDate(year, month, day int) string {
	return EncodeDate(year, month, day)
}

// QuickID 快速ID编码
func QuickID(id int64) string {
	return ToBase64(id)
}

// QuickTimestamp 快速时间戳编码
func QuickTimestamp(ts int64) string {
	return ToTimestampShort(ts)
}

// QuickDate 快速日期编码
func QuickDate(year, month, day int) string {
	return EncodeDate(year, month, day)
}

// QuickOrderID 快速生成订单ID
func QuickOrderID(seq int64) string {
	return GenerateOrderID(seq)
}

// QuickShareCode 快速生成分享码（7天有效）
func QuickShareCode(data int64) string {
	return GenerateShareCode(BusinessShare, 7*24*time.Hour, data)
}

// QuickInviteCode 快速生成邀请码
func QuickInviteCode(userID int64) string {
	return GenerateInviteCode(userID)
}

// QuickCacheKey 快速生成缓存键
func QuickCacheKey(prefix string, id int64) string {
	return GenerateCacheKey(prefix, id)
}

// Presets 预设配置

// DefaultOrderGenerator 默认订单生成器
var DefaultOrderGenerator = NewIDGenerator(IDGeneratorConfig{
	BusinessType: BusinessOrder,
	EnableDate:   true,
	DateBase:     "2024-01-01",
})

// DefaultUserGenerator 默认用户生成器
var DefaultUserGenerator = NewIDGenerator(IDGeneratorConfig{
	BusinessType: BusinessUser,
	MachineID:    1,
})

// DefaultShareGenerator 默认分享码生成器
var DefaultShareGenerator = NewIDGenerator(IDGeneratorConfig{
	BusinessType: BusinessShare,
	EnableDate:   true,
})

// BatchQuick 批量快速生成
func BatchQuick(count int, generator func() string) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = generator()
	}
	return ids
}

// BatchOrderIDs 批量生成订单ID
func BatchOrderIDs(count int) []string {
	return BatchQuick(count, func() string {
		return DefaultOrderGenerator.Generate()
	})
}

// BatchUserIDs 批量生成用户ID
func BatchUserIDs(count int) []string {
	return BatchQuick(count, func() string {
		return DefaultUserGenerator.Generate()
	})
}

// BatchShareCodes 批量生成分享码
func BatchShareCodes(count int, data []int64) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = GenerateShareCode(BusinessShare, 7*24*time.Hour, data[i])
	}
	return ids
}
