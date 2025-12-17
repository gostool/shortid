package shortid

// BusinessType 业务类型定义
type BusinessType byte

const (
	// BusinessMarketing 营销活动
	BusinessMarketing BusinessType = '0'

	// BusinessPromotion 促销活动
	BusinessPromotion BusinessType = '1'

	// BusinessUser 用户相关
	BusinessUser BusinessType = '2'

	// BusinessOrder 订单系统
	BusinessOrder BusinessType = '3'

	// BusinessPayment 支付流水
	BusinessPayment BusinessType = '4'

	// BusinessProduct 商品管理
	BusinessProduct BusinessType = '5'

	// BusinessInventory 库存管理
	BusinessInventory BusinessType = '6'

	// BusinessLogistics 物流跟踪
	BusinessLogistics BusinessType = '7'

	// BusinessSession 会话管理
	BusinessSession BusinessType = '8'

	// BusinessCache 缓存键
	BusinessCache BusinessType = '9'

	// BusinessShare 分享链接
	BusinessShare BusinessType = 'A'

	// BusinessInvite 邀请码
	BusinessInvite BusinessType = 'B'

	// BusinessCoupon 优惠券
	BusinessCoupon BusinessType = 'C'

	// BusinessActivity 活动码
	BusinessActivity BusinessType = 'D'

	// BusinessReservedZ 系统预留
	BusinessReservedZ BusinessType = 'Z'
)

// String 返回业务类型的字符串表示
func (b BusinessType) String() string {
	return string(b)
}
