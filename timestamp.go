package shortid

import (
	"fmt"
	"strings"
	"time"
)

const (
	// 默认基准时间：2024-12-31 23:59:59 UTC
	DefaultBaseline = 1735689599

	// 时间相关常量
	SecondsPerDay = 86400 // 一天的秒数
	Base62Base    = 62    // Base62进制基数

	// 基准年份
	BaseYear = 2000 // 紧凑编码的基准年份

	// 短编码基准时间：2020-01-01 00:00:00 UTC
	ShortBaseline = 1577836800

	// 日期编码的便捷范围
	SingleDayRange = 62 // 小于62天的日期可以用单个字符表示

	// 动态时间编码单位
	UnitMinute = 60
	UnitHour   = 60 * UnitMinute
	UnitDay    = 24 * UnitHour
	UnitMonth  = 30 * UnitDay
	UnitYear   = 365 * UnitDay
)

// 全局日期基准变量
var (
	// DefaultDateBaseline 默认日期基准（2024-12-31）
	DefaultDateBaseline = time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
)

// TimestampEncoder 时间戳编码器接口
type TimestampEncoder interface {
	Encode(ts int64) string
	Decode(s string) (int64, error)
}

// TimestampConfig 时间戳编码配置
type TimestampConfig struct {
	Baseline    int64  // 基准时间戳
	UseDays     bool   // 是否使用天数
	CompactMode bool   // 是否使用紧凑模式
	Precision   string // 精度：s秒/m分钟/h小时/d天
}

// DefaultTimestampEncoder 默认时间戳编码器
type DefaultTimestampEncoder struct {
	config TimestampConfig
}

// NewTimestampEncoder 创建时间戳编码器
func NewTimestampEncoder(config TimestampConfig) *DefaultTimestampEncoder {
	if config.Baseline == 0 {
		config.Baseline = DefaultBaseline
	}
	return &DefaultTimestampEncoder{config: config}
}

// Encode 编码时间戳
func (e *DefaultTimestampEncoder) Encode(ts int64) string {
	if ts == 0 {
		return "0"
	}

	if e.config.UseDays {
		return e.encodeWithDays(ts)
	}

	if e.config.CompactMode {
		return e.encodeCompact(ts)
	}

	return ToBase64(ts)
}

// Decode 解码时间戳
func (e *DefaultTimestampEncoder) Decode(s string) (int64, error) {
	if s == "0" {
		return 0, nil
	}

	if e.config.UseDays {
		return e.decodeWithDays(s)
	}

	if e.config.CompactMode {
		return e.decodeCompact(s)
	}

	return FromBase64(s)
}

// encodeWithDays 使用天数编码
func (e *DefaultTimestampEncoder) encodeWithDays(ts int64) string {
	diff := ts - e.config.Baseline
	days := diff / SecondsPerDay
	seconds := diff % SecondsPerDay

	if days == 0 && seconds == 0 {
		return "0"
	}

	daysStr := ToBase64(days)
	secondsStr := ToBase64(seconds)

	return daysStr + "." + secondsStr
}

// decodeWithDays 解码天数编码
func (e *DefaultTimestampEncoder) decodeWithDays(s string) (int64, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format: %s", s)
	}

	days, err := FromBase64(parts[0])
	if err != nil {
		return 0, err
	}

	seconds, err := FromBase64(parts[1])
	if err != nil {
		return 0, err
	}

	return e.config.Baseline + days*SecondsPerDay + seconds, nil
}

// encodeCompact 紧凑编码
func (e *DefaultTimestampEncoder) encodeCompact(ts int64) string {
	t := time.Unix(ts, 0).UTC()

	// 年份偏移（BaseYear基准）
	yearOffset := t.Year() - BaseYear
	if yearOffset < 0 || yearOffset > 100 {
		return ToBase64(ts)
	}

	// 年月日时分组合
	dayOfYear := getDayOfYear(t)
	timeCode := t.Hour()*60 + t.Minute()

	// 使用62进制编码
	yearStr := encodeWithFixedWidth(int64(yearOffset), Base62Base)
	dayStr := encodeWithFixedWidth(int64(dayOfYear), UnitDay)
	timeStr := encodeWithFixedWidth(int64(timeCode), UnitDay)

	return yearStr + dayStr + timeStr
}

// decodeCompact 解码紧凑编码
func (e *DefaultTimestampEncoder) decodeCompact(s string) (int64, error) {
	if len(s) < 6 {
		return 0, fmt.Errorf("invalid compact format")
	}

	// 解析各部分
	yearPart := s[:2]
	dayPart := s[2:4]
	timePart := s[4:6]

	yearOffset, err := decodeWithFixedWidth(yearPart, Base62Base)
	if err != nil {
		return 0, err
	}

	dayOfYear, err := decodeWithFixedWidth(dayPart, UnitDay)
	if err != nil {
		return 0, err
	}

	timeCode, err := decodeWithFixedWidth(timePart, UnitDay)
	if err != nil {
		return 0, err
	}

	// 构建时间
	year := BaseYear + int(yearOffset)
	date := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	date = date.AddDate(0, 0, int(dayOfYear)-1)
	date = date.Add(time.Duration(timeCode) * time.Minute)

	return date.Unix(), nil
}

// ToTimestampShort 时间戳短编码（V3版本）
// 使用2020年为基准，格式：天数.秒数
func ToTimestampShort(ts int64) string {
	if ts == 0 {
		return "0"
	}

	// 特殊处理2020-01-01
	if ts == ShortBaseline {
		return "1.0"
	}

	baseline := int64(ShortBaseline)
	secondsPerDay := int64(SecondsPerDay)

	diff := ts - baseline
	days := diff / secondsPerDay
	seconds := diff % secondsPerDay

	daysStr := ToBase64(days)
	secondsStr := ToBase64(seconds)

	return daysStr + "." + secondsStr
}

// FromTimestampShort 解码时间戳短编码
func FromTimestampShort(s string) (int64, error) {
	if s == "0" {
		return 0, nil
	}

	// 特殊处理2020-01-01
	if s == "1.0" {
		return ShortBaseline, nil
	}

	baseline := int64(ShortBaseline)
	secondsPerDay := int64(SecondsPerDay)

	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid format: %s", s)
	}

	days, err := FromBase64(parts[0])
	if err != nil {
		return 0, err
	}

	seconds, err := FromBase64(parts[1])
	if err != nil {
		return 0, err
	}

	return baseline + days*secondsPerDay + seconds, nil
}

// EncodeDate 日期编码（相对2024-12-31）
func EncodeDate(year, month, day int) string {
	baseline := DefaultDateBaseline
	target := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	// 计算天数差
	days := int(target.Sub(baseline).Hours() / 24)

	if days == 0 {
		return "0"
	}

	if days > 0 {
		if days < SingleDayRange {
			return string(base62Chars[days])
		}
		return "+" + ToBase64(int64(days))
	} else {
		return "-" + ToBase64(int64(-days))
	}
}

// EncodeDateFromTime 从时间对象编码日期
func EncodeDateFromTime(t time.Time, baseline time.Time) string {
	days := int(t.Sub(baseline).Hours() / 24)

	if days == 0 {
		return "0"
	}

	if days > 0 {
		if days < SingleDayRange {
			return string(base62Chars[days])
		}
		return "+" + ToBase64(int64(days))
	} else {
		return "-" + ToBase64(int64(-days))
	}
}

// DecodeDateToDate 解码日期为时间对象
func DecodeDateToDate(s string, baseline time.Time) (time.Time, error) {
	if s == "0" {
		return baseline, nil
	}

	var days int

	switch s[0] {
	case '+':
		dayInt, decodeErr := FromBase64(s[1:])
		if decodeErr != nil {
			return time.Time{}, decodeErr
		}
		days = int(dayInt)
	case '-':
		dayInt, decodeErr := FromBase64(s[1:])
		if decodeErr != nil {
			return time.Time{}, decodeErr
		}
		days = -int(dayInt)
	default:
		dayInt, decodeErr := FromBase64(s)
		if decodeErr != nil {
			return time.Time{}, decodeErr
		}
		days = int(dayInt)
	}

	return baseline.AddDate(0, 0, days), nil
}

// ToTimestampDynamic 动态基准时间编码
func ToTimestampDynamic(ts int64) string {
	if ts == 0 {
		return "0"
	}

	// 获取当前时间作为基准
	now := time.Now().UTC()
	baseline := now.Unix()
	diff := ts - baseline

	// 使用预定义的时间单位常量
	const (
		Hour  = 60 * UnitMinute
		Day   = 24 * Hour
		Month = 30 * Day
		Year  = 365 * Day
	)

	if diff == 0 {
		return "now"
	}

	absDiff := diff
	if absDiff < 0 {
		absDiff = -absDiff
	}

	switch {
	case absDiff < Hour:
		minutes := absDiff / UnitMinute
		return encodeWithSign(diff, "m", minutes)
	case absDiff < Day:
		hours := absDiff / Hour
		return encodeWithSign(diff, "h", hours)
	case absDiff < Month:
		days := absDiff / Day
		return encodeWithSign(diff, "d", days)
	case absDiff < Year:
		months := absDiff / Month
		return encodeWithSign(diff, "M", months)
	default:
		years := absDiff / Year
		return encodeWithSign(diff, "y", years)
	}
}

// ToTimestampDynamicV3 以2024年为基础的极简编码
func ToTimestampDynamicV3(ts int64) string {
	if ts == 0 {
		return "0"
	}

	// 2024年基准时间戳
	const V3Baseline = 1704067200 // 2024-01-01 00:00:00 UTC
	days := (ts - V3Baseline) / SecondsPerDay

	if days >= 0 && days < SingleDayRange {
		return string(base62Chars[days])
	}

	return ToBase64(ts)
}

// 辅助函数

func getDayOfYear(t time.Time) int {
	year := t.Year()
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	return int(t.Sub(start).Hours()/24) + 1
}

func encodeWithSign(diff int64, unit string, value int64) string {
	var sign string
	if diff < 0 {
		sign = "-"
	}

	encoded := ToBase64(value)
	if value < 10 {
		encoded = fmt.Sprintf("%d", value)
	}

	return sign + unit + encoded
}

func encodeWithFixedWidth(num, max int64) string {
	if num == 0 {
		return "00"
	}

	base := int64(len(base62Chars))
	width := 2

	var result []byte
	for i := 0; i < width; i++ {
		remainder := num % base
		result = append([]byte{base62Chars[remainder]}, result...)
		num = num / base
	}

	for len(result) < width {
		result = append([]byte{base62Chars[0]}, result...)
	}

	return string(result)
}

func decodeWithFixedWidth(s string, max int64) (int64, error) {
	if len(s) != 2 {
		return 0, fmt.Errorf("invalid width: %s", s)
	}

	charToIndex := make(map[byte]int64)
	for i, c := range base62Chars {
		charToIndex[byte(c)] = int64(i)
	}

	base := int64(len(base62Chars))
	result := int64(charToIndex[s[0]])*base + int64(charToIndex[s[1]])

	if result >= max {
		return 0, fmt.Errorf("value out of range: %d", result)
	}

	return result, nil
}
