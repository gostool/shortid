package shortid

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// ============================================================================
// 短编码算法配置（内部使用）
// ============================================================================

// shortEncodingConfig 短编码算法配置
type shortEncodingConfig struct {
	unitsPerDay    int64                       // 一天的单位数（秒/毫秒/纳秒）
	timePartWidth  int                         // 时间部分的固定宽度（3/5/8字符）
	minLength      int                         // 最小长度（4/6/9字符）
	maxTimeInDay   int64                       // 一天内的最大时间单位数
	decodeTimePart func(string) (int64, error) // 解码时间部分的函数
	encodeTimePart func(int64) string          // 编码时间部分的函数
}

// 预定义的精度配置
var (
	secondConfig = shortEncodingConfig{
		unitsPerDay:    int64(SecondsPerDay),
		timePartWidth:  3,
		minLength:      4,
		maxTimeInDay:   MaxSecondsInDay,
		decodeTimePart: decodeWithFixedWidth3,
		encodeTimePart: encodeWithFixedWidth3,
	}

	millisecondConfig = shortEncodingConfig{
		unitsPerDay:    int64(MillisecondsPerDay),
		timePartWidth:  5,
		minLength:      6,
		maxTimeInDay:   MaxMillisecondsInDay,
		decodeTimePart: decodeWithFixedWidth5,
		encodeTimePart: encodeWithFixedWidth5,
	}

	nanosecondConfig = shortEncodingConfig{
		unitsPerDay:    int64(NanosecondsPerDay),
		timePartWidth:  8,
		minLength:      9,
		maxTimeInDay:   MaxNanosecondsInDay,
		decodeTimePart: decodeWithFixedWidth8,
		encodeTimePart: encodeWithFixedWidth8,
	}
)

// ============================================================================
// 1. 短编码算法（Short Encoding）- 通用实现
// ============================================================================

// toTimestampShortInternal 通用的短编码实现
func toTimestampShortInternal(ts int64, baseline int64, config shortEncodingConfig) string {
	if ts == 0 {
		return "0"
	}

	diff := ts - baseline
	days := diff / config.unitsPerDay
	timeUnits := diff % config.unitsPerDay

	// 处理负数时间单位（当diff为负数时）
	if timeUnits < 0 {
		days--
		timeUnits += config.unitsPerDay
	}

	// 天数部分：可变宽度 Base62 编码（1-4字符）
	// 时间部分：固定宽度 Base62 编码
	// days 可能是负数，需要取绝对值
	daysUint := uint64(days)
	if days < 0 {
		daysUint = uint64(-days)
	}
	return EncodeBase62Int(daysUint) + config.encodeTimePart(timeUnits)
}

// fromTimestampShortInternal 通用的短编码解码实现
func fromTimestampShortInternal(s string, baseline int64, config shortEncodingConfig) (int64, error) {
	if s == "0" {
		return baseline, nil
	}

	if len(s) < config.minLength {
		return 0, fmt.Errorf("invalid format: %s (too short, need at least %d characters)", s, config.minLength)
	}

	// 时间部分：固定宽度（从末尾取）
	// 天数部分：剩余部分（1-4个字符）
	splitIdx := len(s) - config.timePartWidth
	timePart := s[splitIdx:]
	daysPart := s[:splitIdx]

	// 解码时间部分
	timeUnits, err := config.decodeTimePart(timePart)
	if err != nil {
		return 0, fmt.Errorf("invalid time part %s: %w", timePart, err)
	}

	// 解码天数（可变宽度）
	daysUint, err := DecodeBase62(daysPart)
	if err != nil {
		return 0, fmt.Errorf("invalid days part %s: %w", daysPart, err)
	}
	// 转换为 int64（days 应该是正数，但需要检查溢出）
	if daysUint > math.MaxInt64 {
		return 0, fmt.Errorf("days value too large: %d", daysUint)
	}
	days := int64(daysUint)

	return baseline + days*config.unitsPerDay + timeUnits, nil
}

// ============================================================================
// 1.1 秒级短编码算法（Second Short Encoding）
// ============================================================================

// ToTimestampShort 将时间戳编码为短格式字符串。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）将时间戳分解为天数和秒数两部分，
// 分别进行 Base62 编码后直接拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(秒数，固定3字符)
// 总长度：4-7 字符
//
// 参数：
//   - ts: 要编码的时间戳（Unix 时间戳，秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+秒数"（无分隔符）
//     特殊值：如果 ts 为 0，返回 "0"
//
// 示例：
//
//	ToTimestampShort(1704067200) // 返回 "0000"（基准时间）
//	ToTimestampShort(1704153600) // 返回 "1000"（基准时间+1天）
func ToTimestampShort(ts int64) string {
	return ToTimestampShortWithBaseline(ts, DefaultBaseline)
}

// ToTimestampShortWithBaseline 使用自定义基准时间将时间戳编码为短格式字符串。
//
// 将时间戳相对于指定基准时间分解为天数和秒数两部分，分别进行 Base62 编码后拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(秒数，固定3字符)
// 总长度：4-7 字符
//
// 参数：
//   - ts: 要编码的时间戳（Unix 时间戳，秒）
//   - baseline: 基准时间戳（Unix 时间戳，秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+秒数"（无分隔符）
//     特殊值：如果 ts 为 0，返回 "0"；基准时间返回 "0000"（4字符）
func ToTimestampShortWithBaseline(ts int64, baseline int64) string {
	return toTimestampShortInternal(ts, baseline, secondConfig)
}

// FromTimestampShort 解码时间戳短编码字符串为时间戳。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）解码短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(秒数，固定3字符)
// 解析规则：从字符串末尾取 3 个字符作为秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 4 字符
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，秒）
//   - error: 如果字符串格式无效或长度不足，返回错误
//
// 示例：
//
//	FromTimestampShort("0000") // 返回 1704067200, nil（基准时间）
//	FromTimestampShort("1000") // 返回 1704153600, nil（基准时间+1天）
func FromTimestampShort(s string) (int64, error) {
	return FromTimestampShortWithBaseline(s, DefaultBaseline)
}

// FromTimestampShortWithBaseline 使用自定义基准时间解码时间戳短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(秒数，固定3字符)
// 解析规则：从字符串末尾取 3 个字符作为秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 4 字符
//   - baseline: 基准时间戳（Unix 时间戳，秒），必须与编码时使用的基准时间一致
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，秒）
//     特殊值：如果 s 为 "0"，返回基准时间
//   - error: 如果字符串格式无效、长度不足或解码失败，返回错误
func FromTimestampShortWithBaseline(s string, baseline int64) (int64, error) {
	return fromTimestampShortInternal(s, baseline, secondConfig)
}

// ============================================================================
// 2. 动态编码算法（Dynamic Encoding）
// ============================================================================

// ToTimestampDynamic 将时间戳编码为动态相对时间格式字符串。
//
// 相对于当前时间编码，使用时间单位（分钟/小时/天/月/年）表示时间差。
// 格式：[符号] + 单位 + Base62(值)
// 总长度：2-4 字符
//
// 单位说明：
//   - "m": 分钟（< 1小时）
//   - "h": 小时（< 1天）
//   - "d": 天（< 1月）
//   - "M": 月（< 1年）
//   - "y": 年（>= 1年）
//
// 参数：
//   - ts: 要编码的时间戳（Unix 时间戳，秒）
//
// 返回：
//   - string: 动态时间编码字符串
//     特殊值：如果 ts 为 0，返回 "0"；如果等于当前时间，返回 "now"
//
// 示例：
//
//	ToTimestampDynamic(time.Now().Unix() + 300)  // 返回 "m5"（5分钟后）
//	ToTimestampDynamic(time.Now().Unix() - 3600) // 返回 "-h1"（1小时前）
func ToTimestampDynamic(ts int64) string {
	return ToTimestampDynamicWithNow(ts, time.Now().UTC().Unix())
}

// ToTimestampDynamicWithNow 使用指定的当前时间将时间戳编码为动态相对时间格式。
//
// 相对于指定时间编码，使用时间单位（分钟/小时/天/月/年）表示时间差。
// 格式：[符号] + 单位 + Base62(值)
//
// 参数：
//   - ts: 要编码的时间戳（Unix 时间戳，秒）
//   - now: 作为基准的当前时间（Unix 时间戳，秒）
//
// 返回：
//   - string: 动态时间编码字符串
//     特殊值：如果 ts 为 0，返回 "0"；如果等于 now，返回 "now"
func ToTimestampDynamicWithNow(ts int64, now int64) string {
	if ts == 0 {
		return "0"
	}

	diff := ts - now
	if diff == 0 {
		return "now"
	}

	absDiff := diff
	if absDiff < 0 {
		absDiff = -absDiff
	}

	return encodeDynamicTime(absDiff, diff)
}

// encodeDynamicTime 根据时间差编码动态时间。
//
// 根据时间差的绝对值选择合适的单位（分钟/小时/天/月/年）进行编码。
//
// 参数：
//   - absDiff: 时间差的绝对值（秒）
//   - diff: 时间差（秒），负数表示过去，正数表示未来
//
// 返回：
//   - string: 动态时间编码字符串，格式为"[符号]单位值"
func encodeDynamicTime(absDiff, diff int64) string {
	switch {
	case absDiff < UnitHour:
		return encodeWithSign(diff, "m", absDiff/UnitMinute)
	case absDiff < int64(SecondsPerDay):
		return encodeWithSign(diff, "h", absDiff/UnitHour)
	case absDiff < UnitMonth:
		return encodeWithSign(diff, "d", absDiff/int64(SecondsPerDay))
	case absDiff < UnitYear:
		return encodeWithSign(diff, "M", absDiff/UnitMonth)
	default:
		return encodeWithSign(diff, "y", absDiff/UnitYear)
	}
}

// FromTimestampDynamic 解码动态时间编码字符串为时间戳。
//
// 需要提供编码时的当前时间作为基准，才能正确解码。
//
// 输入格式：[符号] + 单位 + Base62(值)
// 单位：m(分钟)、h(小时)、d(天)、M(月)、y(年)
//
// 参数：
//   - s: 动态时间编码字符串，长度 2-4 字符
//   - now: 编码时的当前时间（Unix 时间戳，秒），必须与编码时使用的基准时间一致
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，秒）
//     特殊值：如果 s 为 "0"，返回 0；如果 s 为 "now"，返回 now
//   - error: 如果字符串格式无效、单位无效或解码失败，返回错误
//
// 示例：
//
//	now := time.Now().Unix()
//	encoded := ToTimestampDynamic(now + 300)  // "m5"
//	decoded, _ := FromTimestampDynamic(encoded, now) // 返回 now + 300
func FromTimestampDynamic(s string, now int64) (int64, error) {
	if s == "0" {
		return 0, nil
	}

	if s == "now" {
		return now, nil
	}

	if len(s) < 2 {
		return 0, fmt.Errorf("invalid format: %s (too short)", s)
	}

	// 解析符号和单位
	sign := int64(1)
	startIdx := 0
	if s[0] == '-' {
		sign = -1
		startIdx = 1
	}

	unit := string(s[startIdx])
	valueStr := s[startIdx+1:]

	// 解码数值
	valueUint, err := DecodeBase62(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid value part %s: %w", valueStr, err)
	}
	// 转换为 int64（需要检查溢出）
	if valueUint > math.MaxInt64 {
		return 0, fmt.Errorf("value too large: %d", valueUint)
	}
	value := int64(valueUint)

	// 根据单位计算时间差
	diff := getUnitSeconds(unit) * value
	if diff == 0 {
		return 0, fmt.Errorf("invalid unit: %s", unit)
	}

	return now + sign*diff, nil
}

// getUnitSeconds 获取时间单位对应的秒数。
//
// 参数：
//   - unit: 时间单位字符，支持 "m"(分钟)、"h"(小时)、"d"(天)、"M"(月)、"y"(年)
//
// 返回：
//   - int64: 该单位对应的秒数，如果单位无效返回 0
func getUnitSeconds(unit string) int64 {
	switch unit {
	case "m":
		return UnitMinute
	case "h":
		return UnitHour
	case "d":
		return int64(SecondsPerDay)
	case "M":
		return UnitMonth
	case "y":
		return UnitYear
	default:
		return 0
	}
}

// ============================================================================
// 3. 紧凑编码算法（Compact Encoding）
// ============================================================================

// ToTimestampCompact 将时间戳编码为紧凑格式字符串。
//
// 将时间戳分解为年份偏移、年内天数、当天秒数三部分，使用固定宽度 Base62 编码。
// 格式：YY + DD + SSS (固定7字符)
//   - YY: 年份偏移（固定2字符，Base62编码，相对于2000年）
//   - DD: 年内天数（固定2字符，Base62编码，1-366）
//   - SSS: 当天秒数（固定3字符，Base62编码，0-86399）
//
// 年份范围：2000-2100年（年份偏移 0-100）
// 如果超出范围，回退到 Base62 编码（长度可变）
//
// 参数：
//   - ts: 要编码的时间戳（Unix 时间戳，秒）
//
// 返回：
//   - string: 紧凑编码字符串，正常情况固定7字符，超出范围时长度可变
//
// 示例：
//
//	ToTimestampCompact(1704067200) // 返回 "00001000"（2024-01-01 00:00:00 UTC）
func ToTimestampCompact(ts int64) string {
	t := time.Unix(ts, 0).UTC()

	// 年份偏移（相对于基准年份）
	yearOffset := t.Year() - BaseYear

	// 允许年份偏移 0-100（支持约100年范围）
	if yearOffset < 0 || yearOffset > MaxYearOffset {
		// 超出范围，回退到 Base62 编码
		// ts 可能是负数，需要取绝对值
		tsUint := uint64(ts)
		if ts < 0 {
			tsUint = uint64(-ts)
		}
		return EncodeBase62Int(tsUint)
	}

	// 年内天数（1-366）
	dayOfYear := getDayOfYear(t)

	// 当天秒数（0-86399）
	timeCode := int64(t.Hour()*3600 + t.Minute()*60 + t.Second())

	// 使用固定宽度编码
	yearStr := encodeWithFixedWidth2(int64(yearOffset))
	dayStr := encodeWithFixedWidth2(int64(dayOfYear))
	timeStr := encodeWithFixedWidth3(timeCode)

	return yearStr + dayStr + timeStr
}

// FromTimestampCompact 解码紧凑编码字符串为时间戳。
//
// 输入格式：YY + DD + SSS (固定7字符)
//   - YY: 年份偏移（前2字符，Base62编码）
//   - DD: 年内天数（中间2字符，Base62编码）
//   - SSS: 当天秒数（后3字符，Base62编码）
//
// 参数：
//   - s: 紧凑编码字符串，必须恰好7字符
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，秒）
//   - error: 如果字符串长度不是7字符、格式无效或值超出范围，返回错误
//
// 示例：
//
//	FromTimestampCompact("00001000") // 返回 1704067200, nil（2024-01-01 00:00:00 UTC）
func FromTimestampCompact(s string) (int64, error) {
	// 格式：2字符年份 + 2字符天数 + 3字符时间 = 7字符
	if len(s) != 7 {
		return 0, fmt.Errorf("invalid compact format: expected 7 characters, got %d", len(s))
	}

	// 解析各部分
	yearPart := s[:2]
	dayPart := s[2:4]
	timePart := s[4:7]

	// 年份偏移范围是 0-100
	yearOffset, err := decodeWithFixedWidth2(yearPart, MaxYearOffset+1)
	if err != nil {
		return 0, fmt.Errorf("invalid year part %s: %w", yearPart, err)
	}

	dayOfYear, err := decodeWithFixedWidth2(dayPart, MaxDaysInYear+1)
	if err != nil {
		return 0, fmt.Errorf("invalid day part %s: %w", dayPart, err)
	}

	timeCode, err := decodeWithFixedWidth3(timePart)
	if err != nil {
		return 0, fmt.Errorf("invalid time part %s: %w", timePart, err)
	}

	// 构建时间
	year := BaseYear + int(yearOffset)
	date := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	date = date.AddDate(0, 0, int(dayOfYear)-1)
	date = date.Add(time.Duration(timeCode) * time.Second)

	return date.Unix(), nil
}

// ============================================================================
// 辅助函数
// ============================================================================

// getDayOfYear 获取时间在一年中的第几天。
//
// 参数：
//   - t: 时间对象
//
// 返回：
//   - int: 一年中的第几天，范围 1-366（1月1日为1，闰年最多366天）
func getDayOfYear(t time.Time) int {
	return t.YearDay()
}

// encodeWithSign 编码带符号的时间值。
//
// 根据时间差的符号添加"-"前缀，然后拼接单位字符和Base62编码的值。
//
// 参数：
//   - diff: 时间差（秒），负数表示过去，正数表示未来
//   - unit: 时间单位字符（"m"、"h"、"d"、"M"、"y"）
//   - value: 时间值（Base62编码前的数值）
//
// 返回：
//   - string: 编码后的字符串，格式为"[符号]单位Base62值"
//     如果 diff < 0，返回 "-单位值"；否则返回 "单位值"
func encodeWithSign(diff int64, unit string, value int64) string {
	if diff < 0 {
		// value 可能是负数，需要取绝对值
		if value < 0 {
			value = -value
		}
		return "-" + unit + EncodeBase62Int(uint64(value))
	}
	// value 应该是正数，但为了安全也检查一下
	if value < 0 {
		value = -value
	}
	return unit + EncodeBase62Int(uint64(value))
}

// encodeWithFixedWidth 将数字编码为固定宽度的 Base62 字符串。
//
// 如果编码长度超过指定宽度，取后width个字符（低位部分）；
// 如果编码长度小于指定宽度，左侧补零到指定宽度。
//
// 参数：
//   - num: 要编码的数字，负数会被视为0
//   - width: 目标宽度（字符数）
//
// 返回：
//   - string: 固定宽度的 Base62 编码字符串
func encodeWithFixedWidth(num int64, width int) string {
	if num < 0 {
		num = 0
	}

	encoded := EncodeBase62Int(uint64(num))
	encodedLen := len(encoded)

	if encodedLen > width {
		// 取后width个字符（低位部分）
		return encoded[encodedLen-width:]
	}

	// 左侧补零到指定宽度
	if encodedLen < width {
		return strings.Repeat("0", width-encodedLen) + encoded
	}

	return encoded
}

// encodeWithFixedWidth2 将数字编码为固定宽度2字符的 Base62 字符串。
//
// 参数：
//   - num: 要编码的数字
//
// 返回：
//   - string: 固定2字符的 Base62 编码字符串
func encodeWithFixedWidth2(num int64) string {
	return encodeWithFixedWidth(num, 2)
}

// encodeWithFixedWidth3 将数字编码为固定宽度3字符的 Base62 字符串。
//
// 参数：
//   - num: 要编码的数字
//
// 返回：
//   - string: 固定3字符的 Base62 编码字符串
func encodeWithFixedWidth3(num int64) string {
	return encodeWithFixedWidth(num, 3)
}

// decodeWithFixedWidth2 解码固定宽度2字符的 Base62 字符串为数字。
//
// 参数：
//   - s: Base62 编码的字符串，必须恰好2字符
//   - max: 允许的最大值（不包含），用于范围检查
//
// 返回：
//   - int64: 解码后的数字
//   - error: 如果字符串长度不是2字符、格式无效或值超出范围，返回错误
func decodeWithFixedWidth2(s string, max int64) (int64, error) {
	if len(s) != 2 {
		return 0, fmt.Errorf("invalid width: expected 2 characters, got %d", len(s))
	}

	resultUint, err := DecodeBase62(s)
	if err != nil {
		return 0, err
	}

	// 转换为 int64（需要检查溢出）
	if resultUint > math.MaxInt64 {
		return 0, fmt.Errorf("value too large: %d", resultUint)
	}
	result := int64(resultUint)

	if result >= max {
		return 0, fmt.Errorf("value out of range: %d (max: %d)", result, max)
	}

	return result, nil
}

// decodeWithFixedWidth3 解码固定宽度3字符的 Base62 字符串为数字。
//
// 参数：
//   - s: Base62 编码的字符串，必须恰好3字符
//
// 返回：
//   - int64: 解码后的数字，范围 0-86399（一天内的最大秒数）
//   - error: 如果字符串长度不是3字符、格式无效或值超出范围，返回错误
func decodeWithFixedWidth3(s string) (int64, error) {
	return decodeWithFixedWidth(s, 3, MaxSecondsInDay)
}

// ============================================================================
// 4. 毫秒级短编码算法（Millisecond Short Encoding）
// ============================================================================

// ToTimestampShortMs 将毫秒级时间戳编码为短格式字符串。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）将毫秒级时间戳分解为天数和毫秒数两部分，
// 分别进行 Base62 编码后直接拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(毫秒数，固定5字符)
// 总长度：6-9 字符
//
// 参数：
//   - tsMs: 要编码的时间戳（Unix 时间戳，毫秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+毫秒数"（无分隔符）
//     特殊值：如果 tsMs 为 0，返回 "0"
//
// 示例：
//
//	ToTimestampShortMs(1704067200000) // 返回 "0000000"（基准时间）
//	ToTimestampShortMs(1704153600000) // 返回 "1000000"（基准时间+1天）
func ToTimestampShortMs(tsMs int64) string {
	return ToTimestampShortMsWithBaseline(tsMs, DefaultSnowflakeEpochMs)
}

// ToTimestampShortMsWithBaseline 使用自定义基准时间将毫秒级时间戳编码为短格式字符串。
//
// 将毫秒级时间戳相对于指定基准时间分解为天数和毫秒数两部分，分别进行 Base62 编码后拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(毫秒数，固定5字符)
// 总长度：6-9 字符
//
// 参数：
//   - tsMs: 要编码的时间戳（Unix 时间戳，毫秒）
//   - baselineMs: 基准时间戳（Unix 时间戳，毫秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+毫秒数"（无分隔符）
//     特殊值：如果 tsMs 为 0，返回 "0"；基准时间返回 "0000000"（7字符）
func ToTimestampShortMsWithBaseline(tsMs int64, baselineMs int64) string {
	return toTimestampShortInternal(tsMs, baselineMs, millisecondConfig)
}

// FromTimestampShortMs 解码毫秒级时间戳短编码字符串为时间戳。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）解码毫秒级短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(毫秒数，固定5字符)
// 解析规则：从字符串末尾取 5 个字符作为毫秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 6 字符
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，毫秒）
//   - error: 如果字符串格式无效或长度不足，返回错误
//
// 示例：
//
//	FromTimestampShortMs("0000000") // 返回 1704067200000, nil（基准时间）
//	FromTimestampShortMs("1000000") // 返回 1704153600000, nil（基准时间+1天）
func FromTimestampShortMs(s string) (int64, error) {
	return FromTimestampShortMsWithBaseline(s, DefaultSnowflakeEpochMs)
}

// FromTimestampShortMsWithBaseline 使用自定义基准时间解码毫秒级时间戳短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(毫秒数，固定5字符)
// 解析规则：从字符串末尾取 5 个字符作为毫秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 6 字符
//   - baselineMs: 基准时间戳（Unix 时间戳，毫秒），必须与编码时使用的基准时间一致
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，毫秒）
//     特殊值：如果 s 为 "0"，返回基准时间
//   - error: 如果字符串格式无效、长度不足或解码失败，返回错误
func FromTimestampShortMsWithBaseline(s string, baselineMs int64) (int64, error) {
	return fromTimestampShortInternal(s, baselineMs, millisecondConfig)
}

// ============================================================================
// 5. 纳秒级短编码算法（Nanosecond Short Encoding）
// ============================================================================

// ToTimestampShortNs 将纳秒级时间戳编码为短格式字符串。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）将纳秒级时间戳分解为天数和纳秒数两部分，
// 分别进行 Base62 编码后直接拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(纳秒数，固定8字符)
// 总长度：9-12 字符
//
// 参数：
//   - tsNs: 要编码的时间戳（Unix 时间戳，纳秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+纳秒数"（无分隔符）
//     特殊值：如果 tsNs 为 0，返回 "0"
//
// 示例：
//
//	ToTimestampShortNs(1704067200000000000) // 返回 "0000000000"（基准时间）
//	ToTimestampShortNs(1704153600000000000) // 返回 "1000000000"（基准时间+1天）
func ToTimestampShortNs(tsNs int64) string {
	baselineNs := int64(DefaultSnowflakeEpochMs) * 1000000 // 转换为纳秒
	return ToTimestampShortNsWithBaseline(tsNs, baselineNs)
}

// ToTimestampShortNsWithBaseline 使用自定义基准时间将纳秒级时间戳编码为短格式字符串。
//
// 将纳秒级时间戳相对于指定基准时间分解为天数和纳秒数两部分，分别进行 Base62 编码后拼接。
//
// 编码格式：Base62(天数，可变宽度) + Base62(纳秒数，固定8字符)
// 总长度：9-12 字符
//
// 参数：
//   - tsNs: 要编码的时间戳（Unix 时间戳，纳秒）
//   - baselineNs: 基准时间戳（Unix 时间戳，纳秒）
//
// 返回：
//   - string: Base62 编码的字符串，格式为"天数+纳秒数"（无分隔符）
//     特殊值：如果 tsNs 为 0，返回 "0"；基准时间返回 "0000000000"（10字符）
func ToTimestampShortNsWithBaseline(tsNs int64, baselineNs int64) string {
	return toTimestampShortInternal(tsNs, baselineNs, nanosecondConfig)
}

// FromTimestampShortNs 解码纳秒级时间戳短编码字符串为时间戳。
//
// 使用默认基准时间（2024-01-01 00:00:00 UTC）解码纳秒级短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(纳秒数，固定8字符)
// 解析规则：从字符串末尾取 8 个字符作为纳秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 9 字符
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，纳秒）
//   - error: 如果字符串格式无效或长度不足，返回错误
//
// 示例：
//
//	FromTimestampShortNs("0000000000") // 返回 1704067200000000000, nil（基准时间）
//	FromTimestampShortNs("1000000000") // 返回 1704153600000000000, nil（基准时间+1天）
func FromTimestampShortNs(s string) (int64, error) {
	baselineNs := int64(DefaultSnowflakeEpochMs) * 1000000 // 转换为纳秒
	return FromTimestampShortNsWithBaseline(s, baselineNs)
}

// FromTimestampShortNsWithBaseline 使用自定义基准时间解码纳秒级时间戳短编码字符串。
//
// 输入格式：Base62(天数，可变宽度) + Base62(纳秒数，固定8字符)
// 解析规则：从字符串末尾取 8 个字符作为纳秒数部分，剩余部分作为天数部分
//
// 参数：
//   - s: Base62 编码的字符串，长度至少 9 字符
//   - baselineNs: 基准时间戳（Unix 时间戳，纳秒），必须与编码时使用的基准时间一致
//
// 返回：
//   - int64: 解码后的时间戳（Unix 时间戳，纳秒）
//     特殊值：如果 s 为 "0"，返回基准时间
//   - error: 如果字符串格式无效、长度不足或解码失败，返回错误
func FromTimestampShortNsWithBaseline(s string, baselineNs int64) (int64, error) {
	return fromTimestampShortInternal(s, baselineNs, nanosecondConfig)
}

// ============================================================================
// 辅助函数：固定宽度编码/解码
// ============================================================================

// encodeWithFixedWidth5 将数字编码为固定宽度5字符的 Base62 字符串。
//
// 参数：
//   - num: 要编码的数字
//
// 返回：
//   - string: 固定5字符的 Base62 编码字符串
func encodeWithFixedWidth5(num int64) string {
	return encodeWithFixedWidth(num, 5)
}

// encodeWithFixedWidth8 将数字编码为固定宽度8字符的 Base62 字符串。
//
// 参数：
//   - num: 要编码的数字
//
// 返回：
//   - string: 固定8字符的 Base62 编码字符串
func encodeWithFixedWidth8(num int64) string {
	return encodeWithFixedWidth(num, 8)
}

// decodeWithFixedWidth 通用的固定宽度解码函数
func decodeWithFixedWidth(s string, width int, maxValue int64) (int64, error) {
	if len(s) != width {
		return 0, fmt.Errorf("invalid width: expected %d characters, got %d", width, len(s))
	}

	resultUint, err := DecodeBase62(s)
	if err != nil {
		return 0, err
	}

	// 转换为 int64（需要检查溢出）
	if resultUint > math.MaxInt64 {
		return 0, fmt.Errorf("value too large: %d", resultUint)
	}
	result := int64(resultUint)

	if result > maxValue {
		return 0, fmt.Errorf("value out of range: %d (max: %d)", result, maxValue)
	}

	return result, nil
}

// decodeWithFixedWidth5 解码固定宽度5字符的 Base62 字符串为数字。
//
// 参数：
//   - s: Base62 编码的字符串，必须恰好5字符
//
// 返回：
//   - int64: 解码后的数字，范围 0-86399999（一天内的最大毫秒数）
//   - error: 如果字符串长度不是5字符、格式无效或值超出范围，返回错误
func decodeWithFixedWidth5(s string) (int64, error) {
	return decodeWithFixedWidth(s, 5, MaxMillisecondsInDay)
}

// decodeWithFixedWidth8 解码固定宽度8字符的 Base62 字符串为数字。
//
// 参数：
//   - s: Base62 编码的字符串，必须恰好8字符
//
// 返回：
//   - int64: 解码后的数字，范围 0-86399999999999（一天内的最大纳秒数）
//   - error: 如果字符串长度不是8字符、格式无效或值超出范围，返回错误
func decodeWithFixedWidth8(s string) (int64, error) {
	return decodeWithFixedWidth(s, 8, MaxNanosecondsInDay)
}
