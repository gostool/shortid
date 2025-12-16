package shortid

import (
	"fmt"
	"sync"
	"time"
)

// 常量定义
const (
	// DefaultDateBase 默认日期基准（用于日期编码）
	DefaultDateBase = "2024-12-31"

	// DateFormat 日期格式
	DateFormat = "2006-01-02"

	// TimestampPartLength 标准时间戳编码的假设长度
	TimestampPartLength = 6

	// DatePartLength 日期编码的最大长度
	DatePartLength = 3

	// InviteCodeSuffixLength 邀请码使用的用户ID后缀位数
	InviteCodeSuffixLength = 1000000 // 10^6，取用户ID后6位

	// MaxIDLength ID的最大长度限制
	MaxIDLength = 50

	// MinIDLength ID的最小长度限制
	MinIDLength = 2
)

// IDGenerator ID生成器接口
type IDGenerator interface {
	Generate() string
	GenerateWithTimestamp(ts int64) string
	Parse(id string) (*IDInfo, error)
}

// IDInfo ID解析信息
type IDInfo struct {
	Type      string    `json:"type"`      // ID类型
	Business  string    `json:"business"`  // 业务类型
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Sequence  int64     `json:"sequence"`  // 序列号
	Data      string    `json:"data"`      // 附加数据
}

// BusinessType 业务类型定义
type BusinessType byte

const (
	BusinessMarketing BusinessType = '0' // 营销活动
	BusinessPromotion BusinessType = '1' // 促销活动
	BusinessUser      BusinessType = '2' // 用户相关
	BusinessOrder     BusinessType = '3' // 订单系统
	BusinessPayment   BusinessType = '4' // 支付流水
	BusinessProduct   BusinessType = '5' // 商品管理
	BusinessInventory BusinessType = '6' // 库存管理
	BusinessLogistics BusinessType = '7' // 物流跟踪
	BusinessSession   BusinessType = '8' // 会话管理
	BusinessCache     BusinessType = '9' // 缓存键
	BusinessShare     BusinessType = 'A' // 分享链接
	BusinessInvite    BusinessType = 'B' // 邀请码
	BusinessCoupon    BusinessType = 'C' // 优惠券
	BusinessActivity  BusinessType = 'D' // 活动码
	BusinessReservedZ BusinessType = 'Z' // 系统预留
)

// String 返回业务类型的字符串表示
func (b BusinessType) String() string {
	return string(b)
}

// IDGeneratorConfig ID生成器配置
type IDGeneratorConfig struct {
	BusinessType BusinessType `json:"business_type"` // 业务类型
	MachineID    int64        `json:"machine_id"`    // 机器ID
	EnableDate   bool         `json:"enable_date"`   // 是否启用日期编码
	DateBase     string       `json:"date_base"`     // 日期基准 (格式: 2006-01-02)
	Prefix       string       `json:"prefix"`        // 前缀
}

// DefaultIDGenerator 默认ID生成器
type DefaultIDGenerator struct {
	config     IDGeneratorConfig
	mu         sync.RWMutex
	lastTime   int64
	sequence   int64
	dateBase   time.Time
	currentDay string
}

// NewIDGenerator 创建新的ID生成器
func NewIDGenerator(config IDGeneratorConfig) *DefaultIDGenerator {
	gen := &DefaultIDGenerator{
		config:   config,
		lastTime: 0,
		sequence: 0,
	}

	// 设置日期基准
	if config.DateBase != "" {
		if t, err := time.Parse(DateFormat, config.DateBase); err == nil {
			gen.dateBase = t
		} else {
			// 默认使用DefaultDateBase
			if t, err := time.Parse(DateFormat, DefaultDateBase); err == nil {
				gen.dateBase = t
			}
		}
	} else {
		// 使用默认日期基准
		if t, err := time.Parse(DateFormat, DefaultDateBase); err == nil {
			gen.dateBase = t
		}
	}

	return gen
}

// Generate 生成新ID
func (g *DefaultIDGenerator) Generate() string {
	return g.GenerateWithTimestamp(time.Now().Unix())
}

// GenerateWithTimestamp 使用指定时间戳生成ID
func (g *DefaultIDGenerator) GenerateWithTimestamp(ts int64) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Unix(ts, 0).UTC()

	// 处理序列号
	if now.Unix() == g.lastTime {
		g.sequence++
	} else {
		g.lastTime = now.Unix()
		g.sequence = 0
	}

	// 构建ID
	var id string

	// 添加前缀
	if g.config.Prefix != "" {
		id += g.config.Prefix
	}

	// 添加业务类型
	id += g.config.BusinessType.String()

	// 添加时间戳部分
	if g.config.EnableDate {
		// 使用日期编码
		datePart := EncodeDateFromTime(now, g.dateBase)
		id += datePart
	} else {
		// 使用标准Base62时间戳
		timePart := ToBase64(now.Unix())
		id += timePart
	}

	// 添加机器ID（如果有）
	if g.config.MachineID > 0 {
		id += ToBase64(g.config.MachineID)
	}

	// 添加序列号
	if g.sequence > 0 {
		id += ToBase64(g.sequence)
	}

	return id
}

// Parse 解析ID
func (g *DefaultIDGenerator) Parse(id string) (*IDInfo, error) {
	if len(id) < MinIDLength {
		return nil, fmt.Errorf("invalid ID length: must be at least %d characters", MinIDLength)
	}

	info := &IDInfo{
		Type:     "business_id",
		Business: string(id[0]),
	}

	// 解析时间戳部分
	var timePart string
	startIdx := 1

	// 如果有前缀
	if g.config.Prefix != "" && len(id) > len(g.config.Prefix) {
		if id[:len(g.config.Prefix)] == g.config.Prefix {
			startIdx += len(g.config.Prefix)
			info.Business = string(id[startIdx-1])
		}
	}

	// 提取时间部分
	if g.config.EnableDate {
		// 日期编码通常是1-DatePartLength位
		endIdx := startIdx + DatePartLength
		if endIdx > len(id) {
			endIdx = len(id)
		}
		timePart = id[startIdx:endIdx]

		// 解码日期
		t, err := DecodeDateToDate(timePart, g.dateBase)
		if err == nil {
			info.Timestamp = t
		}
	} else {
		// 标准时间戳编码，需要确定长度
		// 假设时间戳部分是TimestampPartLength位
		endIdx := startIdx + TimestampPartLength
		if endIdx > len(id) {
			endIdx = len(id)
		}
		timePart = id[startIdx:endIdx]

		if ts, err := FromBase64(timePart); err == nil {
			info.Timestamp = time.Unix(ts, 0)
		}
	}

	return info, nil
}

// GenerateShareCode 生成分享码
func GenerateShareCode(business BusinessType, expireIn time.Duration, data int64) string {
	expireTime := time.Now().Add(expireIn)
	year, month, day := expireTime.Date()

	// 使用日期编码
	datePart := EncodeDate(year, int(month), day)
	dataPart := ToBase64(data)

	return fmt.Sprintf("%c%s%s", business, datePart, dataPart)
}

// GenerateInviteCode 生成邀请码
func GenerateInviteCode(userID int64) string {
	// 取用户ID的后缀位进行编码，保持简短
	suffix := userID % InviteCodeSuffixLength
	return fmt.Sprintf("%c%s", BusinessInvite, ToBase64(suffix))
}

// GenerateCacheKey 生成缓存键
func GenerateCacheKey(prefix string, id int64) string {
	now := time.Now()
	datePart := EncodeDate(now.Year(), int(now.Month()), now.Day())
	idPart := ToBase64(id)

	return fmt.Sprintf("%s:%s:%s", prefix, datePart, idPart)
}

// GenerateOrderID 生成订单号
func GenerateOrderID(seq int64) string {
	now := time.Now()
	datePart := EncodeDate(now.Year(), int(now.Month()), now.Day())
	seqPart := ToBase64(seq)

	return fmt.Sprintf("%c%s%s", BusinessOrder, datePart, seqPart)
}

// GenerateSessionID 生成会话ID
func GenerateSessionID(userID int64, duration time.Duration) string {
	expireTime := time.Now().Add(duration)
	timePart := ToTimestampDynamic(expireTime.Unix())
	userPart := ToBase64(userID)

	return fmt.Sprintf("%c%s_%s", BusinessSession, timePart, userPart)
}

// BatchGenerate 批量生成ID
func BatchGenerate(gen IDGenerator, count int) []string {
	ids := make([]string, count)
	for i := 0; i < count; i++ {
		ids[i] = gen.Generate()
	}
	return ids
}
