package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gostool/shortid"
)

func main() {
	fmt.Println("=== Conv SDK 使用示例 ===")
	fmt.Println()

	// 1. 快速编码
	fmt.Println("1. 快速编码示例:")
	userID := int64(123456789)
	shortID := shortid.QuickID(userID)
	fmt.Printf("   用户ID: %d -> %s\n", userID, shortID)

	timestamp := time.Now().Unix()
	shortTS := shortid.QuickTimestamp(timestamp)
	fmt.Printf("   时间戳: %d -> %s\n", timestamp, shortTS)

	dateCode := shortid.QuickDate(2025, 6, 15)
	fmt.Printf("   日期码: 2025-06-15 -> %s\n", dateCode)
	fmt.Println()

	// 2. 订单相关
	fmt.Println("2. 订单相关示例:")
	orderID := shortid.QuickOrderID(1001)
	fmt.Printf("   订单ID: %s\n", orderID)

	// 批量生成订单ID
	orderIDs := shortid.BatchOrderIDs(5)
	fmt.Printf("   批量订单ID: %v\n", orderIDs)
	fmt.Println()

	// 3. 社交分享
	fmt.Println("3. 社交分享示例:")
	shareCode := shortid.QuickShareCode(12345)
	fmt.Printf("   分享码: %s (7天有效)\n", shareCode)

	inviteCode := shortid.QuickInviteCode(987654321)
	fmt.Printf("   邀请码: %s\n", inviteCode)
	fmt.Println()

	// 4. 缓存键
	fmt.Println("4. 缓存键示例:")
	cacheKey := shortid.QuickCacheKey("user", 12345)
	fmt.Printf("   缓存键: %s\n", cacheKey)
	fmt.Println()

	// 5. 高级ID生成器
	fmt.Println("5. 高级ID生成器示例:")

	// 配置订单生成器
	orderGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: shortid.BusinessOrder,
		EnableDate:   true,
		DateBase:     "2024-01-01",
		MachineID:    1,
	})

	// 生成订单ID
	for i := 0; i < 3; i++ {
		id := orderGen.Generate()
		info, err := orderGen.Parse(id)
		if err != nil {
			log.Printf("解析错误: %v", err)
			continue
		}
		fmt.Printf("   订单ID: %s (业务:%s, 时间:%v)\n",
			id, info.Business, info.Timestamp)
	}
	fmt.Println()

	// 6. 时间戳编码器
	fmt.Println("6. 时间戳编码器示例:")

	// 创建自定义配置的时间戳编码器
	tsEncoder := shortid.NewTimestampEncoder(shortid.TimestampConfig{
		UseDays:     true,
		CompactMode: false,
	})

	now := time.Now().Unix()
	encodedTS := tsEncoder.Encode(now)
	decodedTS, err := tsEncoder.Decode(encodedTS)
	if err != nil {
		log.Printf("解码错误: %v", err)
	} else {
		fmt.Printf("   时间戳: %d -> %s -> %d\n", now, encodedTS, decodedTS)
	}

	// 动态时间编码
	dynamicTS := shortid.ToTimestampDynamic(time.Now().Add(2 * time.Hour).Unix())
	fmt.Printf("   动态时间: 2小时后 -> %s\n", dynamicTS)
	fmt.Println()

	// 7. 互联网应用场景示例
	fmt.Println("7. 互联网应用场景示例:")

	// 电商场景
	fmt.Println("   电商场景:")
	eCommerceOrderID := shortid.GenerateOrderID(12345)
	paymentID := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: shortid.BusinessPayment,
		EnableDate:   true,
	}).Generate()
	logisticsID := shortid.GenerateShareCode(shortid.BusinessLogistics, 7*24*time.Hour, 98765)

	fmt.Printf("     订单号: %s\n", eCommerceOrderID)
	fmt.Printf("     支付流水: %s\n", paymentID)
	fmt.Printf("     物流单号: %s\n", logisticsID)

	// 社交场景
	fmt.Println("   社交场景:")
	sessionID := shortid.GenerateSessionID(12345, 2*time.Hour)
	postID := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: 'P', // Post
		EnableDate:   true,
	}).Generate()
	commentID := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: 'c', // Comment
		Prefix:       "cmt",
	}).Generate()

	fmt.Printf("     会话ID: %s\n", sessionID)
	fmt.Printf("     动态ID: %s\n", postID)
	fmt.Printf("     评论ID: %s\n", commentID)

	// 营销场景
	fmt.Println("   营销场景:")
	couponCode := shortid.GenerateShareCode(shortid.BusinessCoupon, 30*24*time.Hour, 520)
	promoCode := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: shortid.BusinessMarketing,
		Prefix:       "PROMO",
	}).Generate()

	fmt.Printf("     优惠券码: %s\n", couponCode)
	fmt.Printf("     促销码: %s\n", promoCode)
	fmt.Println()

	// 8. 短链接生成和解析示例
	fmt.Println("8. 短链接生成和解析示例:")

	// 生成商品分享短链接
	productID := int64(888666)
	productLink := shortid.GenerateShareCode(shortid.BusinessShare, 7*24*time.Hour, productID)
	fmt.Printf("   商品分享链接: %s (7天有效)\n", productLink)

	// 生成活动报名短链接
	activityID := int64(20250615)
	activityLink := shortid.GenerateShareCode(shortid.BusinessActivity, 30*24*time.Hour, activityID)
	fmt.Printf("   活动报名链接: %s (30天有效)\n", activityLink)

	// 生成限时优惠短链接
	promotionID := int64(666888)
	promotionLink := shortid.GenerateShareCode(shortid.BusinessPromotion, 3*24*time.Hour, promotionID)
	fmt.Printf("   限时优惠链接: %s (3天有效)\n", promotionLink)

	// 解析短链接信息
	fmt.Println("\n   解析短链接信息:")

	// 解析商品分享链接
	if productInfo, err := parseShareCode(productLink); err == nil {
		fmt.Printf("   %s -> 商业类型:%c, 数据ID:%d\n",
			productLink, productInfo.Business, productInfo.DataID)
	}

	// 解析活动报名链接
	if activityInfo, err := parseShareCode(activityLink); err == nil {
		fmt.Printf("   %s -> 商业类型:%c, 数据ID:%d\n",
			activityLink, activityInfo.Business, activityInfo.DataID)
	}

	// 解析限时优惠链接
	if promotionInfo, err := parseShareCode(promotionLink); err == nil {
		fmt.Printf("   %s -> 商业类型:%c, 数据ID:%d\n",
			promotionLink, promotionInfo.Business, promotionInfo.DataID)
	}

	// 9. 批量生成短链接
	fmt.Println("\n9. 批量生成短链接示例:")

	// 为电商商品批量生成分享链接
	productIDs := []int64{1001, 1002, 1003, 1004, 1005}
	fmt.Println("   批量商品分享链接:")
	for _, pid := range productIDs {
		link := shortid.GenerateShareCode(shortid.BusinessShare, 7*24*time.Hour, pid)
		fmt.Printf("     商品%d: %s\n", pid, link)
	}

	// 为营销活动批量生成推广链接
	campaignIDs := []int64{8888, 8889, 8890}
	fmt.Println("\n   批量营销推广链接:")
	for _, cid := range campaignIDs {
		link := shortid.GenerateShareCode(shortid.BusinessMarketing, 15*24*time.Hour, cid)
		fmt.Printf("     活动%d: %s\n", cid, link)
	}

	// 10. 自定义短链接格式示例
	fmt.Println("\n10. 自定义短链接格式示例:")

	// 使用自定义前缀生成短链接
	customGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: shortid.BusinessShare,
		Prefix:       "LINK",
		EnableDate:   true,
	})

	customLink1 := customGen.Generate()
	customLink2 := customGen.Generate()
	fmt.Printf("   带前缀的短链接: %s\n", customLink1)
	fmt.Printf("   带前缀的短链接: %s\n", customLink2)

	// 生成带时间戳的追踪链接
	trackingGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
		BusinessType: 'T', // Tracking
		EnableDate:   true,
		MachineID:    100, // 渠道ID
	})

	trackingLink := trackingGen.Generate()
	fmt.Printf("   追踪链接: %s (包含渠道信息)\n", trackingLink)

	fmt.Println("\n=== 示例完成 ===")
}

// ShareCodeInfo 短链接信息结构
type ShareCodeInfo struct {
	Business byte  // 业务类型
	DataID   int64 // 数据ID
}

// parseShareCode 解析短链接信息
// 注意：这是一个简化的解析示例，实际应用中可能需要更复杂的解析逻辑
func parseShareCode(code string) (*ShareCodeInfo, error) {
	if len(code) < 2 {
		return nil, fmt.Errorf("invalid share code format")
	}

	info := &ShareCodeInfo{
		Business: code[0],
	}

	// 尝试解析剩余部分为数据ID
	dataPart := code[1:]
	if dataID, err := shortid.FromBase64(dataPart); err == nil {
		info.DataID = dataID
	} else {
		// 如果解析失败，可能是带日期格式的编码
		// 这里简化处理，实际需要根据具体的编码格式进行解析
		return nil, fmt.Errorf("unable to parse share code: %v", err)
	}

	return info, nil
}
