package shortid

import "errors"

// ============================================================================
// Generator 相关错误定义
// ============================================================================

var (
	// ErrOverTimeLimit 时间戳溢出错误
	// 当生成ID时，如果时间戳超过42位范围（约139年），返回此错误
	ErrOverTimeLimit = errors.New("over the time limit")

	// ErrInvalidMachineID 无效机器ID错误
	// 当机器ID超出有效范围（0-63）时，返回此错误
	ErrInvalidMachineID = errors.New("invalid machine id")

	// ErrInvalidSequence 无效序列号错误
	// 当序列号超出有效范围（0-127）时，返回此错误
	ErrInvalidSequence = errors.New("invalid sequence number")

	// ErrInvalidBusinessType 无效业务类型错误
	// 当业务类型超出有效范围（0-255）时，返回此错误
	ErrInvalidBusinessType = errors.New("invalid business type")
)

