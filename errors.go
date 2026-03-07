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

	// ErrInvalidConfig 无效配置错误
	// 当生成器配置字段组合非法时返回此错误
	ErrInvalidConfig = errors.New("invalid generator config")

	// ErrMachineIDLeaseUnavailable 机器ID租约不可用
	// 当所有机器ID都被占用或无法申请时返回此错误
	ErrMachineIDLeaseUnavailable = errors.New("machine id lease unavailable")

	// ErrMachineIDLeaseLost 机器ID租约已失效
	// 当续租失败或租约已不属于当前实例时返回此错误
	ErrMachineIDLeaseLost = errors.New("machine id lease lost")
)
