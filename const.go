package shortid

import "time"

// ============================================================================
// ID 相关常量
// ============================================================================

const (
	// DefaultDateBase 默认日期基准（用于日期编码）
	// 统一基准时间：2024-01-01 00:00:00 UTC
	DefaultDateBase = "2024-01-01"

	// DateFormat 日期格式
	DateFormat = "2006-01-02"

	// TimestampPartLength 标准时间戳编码的假设长度
	TimestampPartLength = 6

	// DatePartLength 日期编码的固定长度（固定3字符）
	DatePartLength = 3

	// MaxSequencePerSecond 每秒最大序列号（支持5000 QPS）
	// Base62编码：62^3 = 238328，足够支持5000/s
	MaxSequencePerSecond = 5000

	// SequencePartLength 序列号编码的固定长度（固定3字符）
	SequencePartLength = 3

	// InviteCodeSuffixLength 邀请码使用的用户ID后缀位数
	InviteCodeSuffixLength = 1000000 // 10^6，取用户ID后6位

	// MaxIDLength ID的最大长度限制
	MaxIDLength = 50

	// MinIDLength ID的最小长度限制
	MinIDLength = 2
)

// ============================================================================
// 时间戳相关常量
// ============================================================================

const (
	// DefaultBaseline 默认基准时间：2024-01-01 00:00:00 UTC
	// 统一基准时间，用于所有时间戳和日期编码
	DefaultBaseline = 1704067200

	// SecondsPerDay 一天的秒数
	SecondsPerDay = 86400

	// Base62Base Base62进制基数
	Base62Base = 62

	// BaseYear 紧凑编码的基准年份（用于年份偏移计算）
	BaseYear = 2000

	// SingleDayRange 日期编码的便捷范围（小于62天的日期可以用单个字符表示）
	SingleDayRange = 62

	// UnitMinute 分钟单位（秒数）
	UnitMinute = 60

	// UnitHour 小时单位（秒数）
	UnitHour = 60 * UnitMinute

	// UnitDay 天单位（秒数）
	UnitDay = 24 * UnitHour

	// UnitMonth 月单位（秒数，按30天计算）
	UnitMonth = 30 * UnitDay

	// UnitYear 年单位（秒数，按365天计算）
	UnitYear = 365 * UnitDay

	// MaxSecondsInDay 一天内的最大秒数（0-86399）
	MaxSecondsInDay = 86399

	// MillisecondsPerDay 一天的毫秒数
	MillisecondsPerDay = 86400000

	// MaxMillisecondsInDay 一天内的最大毫秒数（0-86399999）
	MaxMillisecondsInDay = 86399999

	// NanosecondsPerDay 一天的纳秒数
	NanosecondsPerDay = 86400000000000

	// MaxNanosecondsInDay 一天内的最大纳秒数（0-86399999999999）
	MaxNanosecondsInDay = 86399999999999

	// MaxDaysInYear 一年内的最大天数（1-366）
	MaxDaysInYear = 366

	// MaxYearOffset 最大年份偏移（支持约100年范围）
	MaxYearOffset = 100

	// TimestampPrecision 时间戳精度枚举
	// 用于配置时间戳压缩的精度级别
)

// TimestampPrecision 时间戳精度类型
type TimestampPrecision int

const (
	// PrecisionSecond 秒级精度（默认）
	PrecisionSecond TimestampPrecision = iota
	// PrecisionMillisecond 毫秒级精度
	PrecisionMillisecond
	// PrecisionNanosecond 纳秒级精度
	PrecisionNanosecond
)

// ============================================================================
// 编码相关常量
// ============================================================================

const ()

// ============================================================================
// 辅助函数
// ============================================================================

// DefaultDateBaseline 返回默认日期基准（2024-01-01）
// 统一基准时间，基于 DefaultBaseline 常量
func DefaultDateBaseline() time.Time {
	return time.Unix(DefaultBaseline, 0).UTC()
}

// ============================================================================
// Snowflake 相关常量
// ============================================================================

const (
	// Snowflake位分配
	SnowflakeTimestampBits = 42 // 时间戳位数（毫秒精度）
	SnowflakeBusinessBits  = 8  // 业务类型位数
	SnowflakeMachineBits   = 6  // 机器ID位数
	SnowflakeSequenceBits  = 7  // 序列号位数

	// Snowflake位移
	SnowflakeTimestampShift = SnowflakeBusinessBits + SnowflakeMachineBits + SnowflakeSequenceBits // 21
	SnowflakeBusinessShift  = SnowflakeMachineBits + SnowflakeSequenceBits                         // 13
	SnowflakeMachineShift   = SnowflakeSequenceBits                                                // 7

	// Snowflake最大值
	SnowflakeMaxSequence = (1 << SnowflakeSequenceBits) - 1 // 127
	SnowflakeMaxMachine  = (1 << SnowflakeMachineBits) - 1  // 63
	SnowflakeMaxBusiness = (1 << SnowflakeBusinessBits) - 1 // 255

	// Snowflake时间相关
	// DefaultSnowflakeEpochMs 默认Snowflake纪元（毫秒）：2024-01-01 00:00:00 UTC
	DefaultSnowflakeEpochMs = DefaultBaseline * 1000

	// MaxClockBackwardMs 最大可接受的时钟回退（毫秒）
	MaxClockBackwardMs = 1000 // 1秒

	// MaxBatchCount 批量生成的最大数量限制
	// 防止一次性生成过多ID导致内存溢出或性能问题
	MaxBatchCount = 10000
)

// GetSnowflakeTimestampShift 获取时间戳位移
func GetSnowflakeTimestampShift() int {
	return SnowflakeTimestampShift
}

// GetSnowflakeBusinessShift 获取业务类型位移
func GetSnowflakeBusinessShift() int {
	return SnowflakeBusinessShift
}

// GetSnowflakeMachineShift 获取机器ID位移
func GetSnowflakeMachineShift() int {
	return SnowflakeMachineShift
}
