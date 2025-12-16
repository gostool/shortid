# ShortID - Go 语言 ID 和时间戳压缩库

[![Go Version](https://img.shields.io/badge/go-1.22+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](LICENSE)

一个高性能的 ID 生成和时间戳压缩库，专为互联网应用设计。提供 Base62 编码、时间戳压缩、业务 ID 生成等功能。

## ✨ 特性

- 🚀 **高性能**：单次编码操作 < 150ns
- 📦 **高压缩率**：Base62 编码可节省 30-50% 的存储空间
- 🔒 **URL 安全**：所有编码结果可直接用于 URL
- 🎯 **业务友好**：内置多种业务类型预定义（订单、支付、分享等）
- ⚡ **零依赖**：纯 Go 实现，无外部依赖
- 🔧 **灵活配置**：支持自定义业务类型、日期基准、机器ID等

## 📦 安装

```bash
go get github.com/gostool/shortid
```

## 🚀 快速开始

### 导入

```go
import "github.com/gostool/shortid"
```

### 基础使用

#### 1. 数字转短字符串

```go
// ID 编码
encoded := shortid.QuickID(123456789)        // "8m0Kx"
decoded, err := shortid.FromBase64(encoded) // 123456789, nil

// 批量生成订单ID
batchIDs := shortid.BatchOrderIDs(100)
```

#### 2. 时间戳压缩

```go
import "time"

// 短时间戳编码（以2020年为基准）
shortTS := shortid.QuickTimestamp(time.Now().Unix()) // "z6.4gw"

// 动态基准编码（相对当前时间）
dynamicTS := shortid.ToTimestampDynamic(
    time.Now().Add(2 * time.Hour).Unix(), // "h2"
)

// 日期编码（相对2024-12-31）
dateCode := shortid.QuickDate(2025, 6, 15) // "+2G"
```

#### 3. 专用 ID 生成

```go
// 订单ID
orderID := shortid.QuickOrderID(1001) // "3+5E3d7"

// 分享码（7天有效）
shareCode := shortid.QuickShareCode(12345) // "A+5L3d7"

// 邀请码
inviteCode := shortid.QuickInviteCode(987654321) // "B2Kdz"

// 缓存键
cacheKey := shortid.QuickCacheKey("user", 12345) // "user:+5E:3d7"
```

## 📚 核心组件

### 1. Base62 编码

将数字转换为 62 进制字符串，提供高压缩率。

```go
// 编码
shortid.ToBase64(12345)      // "3d7"
shortid.ToBase64URL(12345)   // "3d7" (URL安全)

// 解码
shortid.FromBase64("3d7")    // 12345, nil
```

**特性：**
- ✅ 支持正负数
- ✅ 压缩率约 30-50%
- ✅ URL 安全
- ✅ 支持 Base91 编码（更高压缩率）

### 2. 时间戳编码器

提供多种时间戳压缩方案。

#### 短编码（推荐）

```go
// 以2020年为基准
encoded := shortid.ToTimestampShort(1749907200)  // "hG.0"
decoded, err := shortid.FromTimestampShort("hG.0") // 1749907200, nil
```

#### 动态编码

```go
// 相对当前时间，适合短期场景
dynamicTS := shortid.ToTimestampDynamic(
    time.Now().Add(time.Hour).Unix(), // "h1"
)
```

#### 日期编码

```go
// 相对2024-12-31，适合日期相关场景
shortid.EncodeDate(2025, 1, 1)      // "1"
shortid.EncodeDate(2024, 12, 30)    // "-1"
shortid.EncodeDate(2025, 6, 15)     // "+2G"
```

### 3. ID 生成器

灵活的 ID 生成器，支持业务前缀、日期、机器ID等配置。

#### 基础配置

```go
config := shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessOrder,  // 业务类型
    EnableDate:   true,                    // 启用日期
    DateBase:     "2024-01-01",           // 日期基准
    MachineID:    1,                       // 机器ID
}

gen := shortid.NewIDGenerator(config)
id := gen.Generate()  // "3+5E1"

// 解析ID
info, err := gen.Parse(id)
```

#### 业务类型预定义

```go
const (
    BusinessMarketing   shortid.BusinessType = '0'  // 营销活动
    BusinessPromotion   shortid.BusinessType = '1'  // 促销活动
    BusinessUser        shortid.BusinessType = '2'  // 用户相关
    BusinessOrder       shortid.BusinessType = '3'  // 订单系统
    BusinessPayment     shortid.BusinessType = '4'  // 支付流水
    BusinessProduct     shortid.BusinessType = '5'  // 商品管理
    BusinessInventory   shortid.BusinessType = '6'  // 库存管理
    BusinessLogistics   shortid.BusinessType = '7'  // 物流跟踪
    BusinessSession     shortid.BusinessType = '8'  // 会话管理
    BusinessCache       shortid.BusinessType = '9'  // 缓存键
    BusinessShare       shortid.BusinessType = 'A'  // 分享链接
    BusinessInvite      shortid.BusinessType = 'B'  // 邀请码
    BusinessCoupon      shortid.BusinessType = 'C'  // 优惠券
    BusinessActivity    shortid.BusinessType = 'D'  // 活动码
    BusinessReservedZ   shortid.BusinessType = 'Z'  // 系统预留
)
```

## 🎯 便捷 API

提供简单易用的快速 API。

```go
// 快速编码
shortid.QuickID(12345)                // Base62编码
shortid.QuickTimestamp(now.Unix())     // 时间戳压缩
shortid.QuickDate(2025, 6, 15)        // 日期编码

// 快速生成
shortid.QuickOrderID(1001)            // 订单ID
shortid.QuickShareCode(12345)         // 分享码
shortid.QuickInviteCode(987654321)    // 邀请码
shortid.QuickCacheKey("user", 12345)  // 缓存键

// 批量操作
shortid.BatchOrderIDs(100)            // 批量订单ID
shortid.BatchUserIDs(50)              // 批量用户ID
```

## 💡 应用场景

### 1. 电商系统

#### 订单管理
```go
// 生成订单ID（自动包含日期和序列号）
orderID := shortid.QuickOrderID(1001)  // "3+5E3d7"
// 格式说明：3(订单) + 5E(日期) + 3d7(序列号)

// 批量生成订单ID
orderIDs := shortid.BatchOrderIDs(100)
// 适用于：批量导入、数据迁移等场景
```

#### 支付流水
```go
// 支付流水号（包含日期和机器ID）
paymentGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessPayment,  // '4'
    EnableDate:   true,
    DateBase:     "2024-01-01",
    MachineID:    100,  // 支付网关ID
})
paymentID := paymentGen.Generate()  // "4+5E1"
```

#### 物流跟踪
```go
// 物流单号（7天有效期）
logisticsID := shortid.GenerateShareCode(
    shortid.BusinessLogistics,  // '7'
    7*24*time.Hour,              // 7天有效期
    98765,                       // 订单ID
)
// 结果：7+5LpGZ（7天有效，过期自动失效）
```

**使用场景：**
- ✅ 订单号生成：短小精悍，易于分享和记忆
- ✅ 支付流水：包含机器ID，便于追踪和排查
- ✅ 物流单号：带有效期，自动过期，安全性高

### 2. 社交平台

#### 用户会话管理
```go
// 会话ID（2小时有效期）
sessionID := shortid.GenerateSessionID(12345, 2*time.Hour)
// 格式：8h2_3d7（8=会话类型，h2=2小时，3d7=用户ID）

// 缓存键生成
cacheKey := shortid.QuickCacheKey("user", 12345)
// 结果：user:+5E:3d7（前缀 + 日期 + ID）
```

#### 内容发布
```go
// 动态/帖子ID
postGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: 'P',  // Post
    EnableDate:   true,
})
postID := postGen.Generate()  // "P+5E"

// 评论ID（带前缀）
commentGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: 'c',  // Comment
    Prefix:       "cmt",
})
commentID := commentGen.Generate()  // "cmtc1VvqU4"
```

#### 分享链接
```go
// 分享码（7天有效）
shareCode := shortid.QuickShareCode(12345)  // "A+5L3d7"
// 格式：A(分享) + 5L(7天后过期) + 3d7(内容ID)

// 生成分享URL
shareURL := fmt.Sprintf("https://example.com/share/%s", shareCode)
```

**使用场景：**
- ✅ 会话管理：自动过期，提高安全性
- ✅ 内容ID：短小易分享，适合社交媒体
- ✅ 分享链接：带有效期，防止链接泄露风险

### 3. 营销系统

#### 优惠券系统
```go
// 优惠券码（30天有效）
couponCode := shortid.GenerateShareCode(
    shortid.BusinessCoupon,  // 'C'
    30*24*time.Hour,         // 30天有效期
    520,                     // 优惠券ID
)
// 结果：C+688o

// 带前缀的优惠券
couponGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessCoupon,
    Prefix:       "CPN",
    EnableDate:   true,
})
customCoupon := couponGen.Generate()  // "CPNC+5E"
```

#### 活动推广
```go
// 活动报名链接（30天有效）
activityLink := shortid.GenerateShareCode(
    shortid.BusinessActivity,  // 'D'
    30*24*time.Hour,
    20250615,  // 活动ID
)
// 结果：D+681mY6P

// 促销码（带自定义前缀）
promoGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessMarketing,  // '0'
    Prefix:       "PROMO",
})
promoCode := promoGen.Generate()  // "PROMO01VvqU4"
```

**使用场景：**
- ✅ 优惠券：短码易输入，带有效期自动失效
- ✅ 活动推广：批量生成，便于追踪转化
- ✅ 促销码：自定义前缀，提升品牌识别度

### 4. 短链接服务

#### 基础使用
```go
// 生成短链接（7天有效）
productLink := shortid.GenerateShareCode(
    shortid.BusinessShare,
    7*24*time.Hour,
    888666,
)
// 结果：A+5L3Jbk
// URL: https://short.ly/A+5L3Jbk

// 批量生成
productIDs := []int64{1001, 1002, 1003}
for _, pid := range productIDs {
    link := shortid.GenerateShareCode(shortid.BusinessShare, 7*24*time.Hour, pid)
    fmt.Printf("商品%d: %s\n", pid, link)
}
```

**使用场景：**
- ✅ URL 缩短：将长链接压缩为短码，节省空间
- ✅ 链接追踪：通过渠道ID追踪来源
- ✅ 批量生成：为大量内容生成短链接

> 💡 **完整实现示例**：查看 [examples/shortlink_service/](examples/shortlink_service/) 了解如何构建完整的短链接服务，包括高并发优化和性能测试。

### 5. 缓存系统

#### 缓存键生成
```go
// 用户缓存键
userKey := shortid.QuickCacheKey("user", 12345)
// 结果：user:+5E:3d7（前缀 + 日期 + ID）

// 会话缓存键
sessionKey := shortid.QuickCacheKey("session", 987654321)
// 结果：session:+5E:3d7

// 商品缓存键
productKey := shortid.QuickCacheKey("product", 5555)
// 结果：product:+5E:3d7
```

**优势：**
- ✅ 自动包含日期，便于按日期清理缓存
- ✅ 短小精悍，节省内存空间
- ✅ 易于识别和管理

### 6. 邀请码系统

#### 用户邀请码
```go
// 生成邀请码（固定2个字符）
inviteCode := shortid.QuickInviteCode(987654321)
// 结果：Bp（B=邀请类型，p=用户ID编码）

// 特点：
// - 长度固定为2个字符，易于分享
// - 支持最多62个不同用户（0-61）
// - 适合小规模邀请场景
```

**使用场景：**
- ✅ 内测邀请：短码易输入，提升用户体验
- ✅ 推荐奖励：易于分享和追踪
- ✅ 活动邀请：固定长度，便于管理

### 7. 时间戳压缩

#### 会话过期时间
```go
// 动态时间编码（相对当前时间）
expireTime := time.Now().Add(2 * time.Hour).Unix()
dynamicTS := shortid.ToTimestampDynamic(expireTime)
// 结果：h2（2小时后）

// 缓存过期时间
cacheExpire := shortid.ToTimestampDynamic(
    time.Now().Add(30 * time.Minute).Unix(),
)
// 结果：m30（30分钟后）
```

#### 事件调度
```go
// 活动开始时间
eventTime := time.Date(2025, 6, 18, 0, 0, 0, 0, time.UTC)
eventCode := shortid.QuickDate(2025, 6, 18)
// 结果：+2G（相对2024-12-31）

// 节日提醒
holiday := shortid.QuickDate(2024, 2, 10)  // 春节
// 结果：-5f
```

**使用场景：**
- ✅ 会话管理：压缩过期时间，节省存储
- ✅ 事件调度：日期编码，便于排序和查询
- ✅ 缓存管理：动态时间，自动过期

## ⚡ 性能基准

基于 Apple M1 Pro 的测试结果：

```
BenchmarkPerformance/base64_encode-8          9361320    121.3 ns/op
BenchmarkPerformance/timestamp_encode-8       8678692    140.4 ns/op
BenchmarkPerformance/id_generate-8            4306465    287.9 ns/op

BenchmarkTimestampEncoders/dynamic-8         13585750     88.51 ns/op  // 最快
BenchmarkTimestampEncoders/short-8           8578692    140.6 ns/op
BenchmarkTimestampEncoders/base62-8           7167398    162.1 ns/op
```

## 🎯 快速选择指南

根据你的使用场景，快速选择合适的 API：

| 场景 | 推荐 API | 示例 | 特点 |
|------|---------|------|------|
| **订单ID** | `QuickOrderID()` | `"3+5E3d7"` | 自动包含日期，短小易读 |
| **用户ID编码** | `QuickID()` | `"8m0Kx"` | 纯数字转短字符串 |
| **分享链接** | `QuickShareCode()` | `"A+5L3d7"` | 7天有效期，自动过期 |
| **邀请码** | `QuickInviteCode()` | `"Bp"` | 固定2字符，易于分享 |
| **缓存键** | `QuickCacheKey()` | `"user:+5E:3d7"` | 包含日期，便于管理 |
| **时间戳压缩** | `QuickTimestamp()` | `"z6.8q8"` | 高压缩率，适合存储 |
| **动态时间** | `ToTimestampDynamic()` | `"h2"` | 相对当前时间，适合短期 |
| **日期编码** | `QuickDate()` | `"+2G"` | 日期压缩，适合调度 |

### 场景选择建议

**需要长期存储？**
- ✅ 使用 `QuickOrderID()` 或自定义 `IDGenerator`
- ✅ 包含日期和序列号，保证唯一性

**需要自动过期？**
- ✅ 使用 `QuickShareCode()` 或 `GenerateShareCode()`
- ✅ 设置有效期，过期自动失效

**需要批量生成？**
- ✅ 使用 `BatchOrderIDs()` 或循环调用生成器
- ✅ 复用生成器实例，提高性能

**需要自定义格式？**
- ✅ 使用 `NewIDGenerator()` 配置自定义参数
- ✅ 支持前缀、机器ID、日期基准等

## 📖 最佳实践

### 1. 选择合适的编码方案

- **长期存储ID**：使用业务码 + 时间戳 + 序列号
- **临时ID**：使用业务码 + 日期编码（1-3位）
- **缓存键**：使用前缀 + 日期 + ID
- **分享码**：使用业务码 + 过期时间

### 2. ID 格式建议

```
格式: [业务码(1)][日期(1-3)][时间戳(4-6)][序列号(1-3)]
示例: 3+5E3d7
        ^   ^ ^   ^
        |   | |   |
        |   | |   +-- 序列号
        |   | +------ 时间戳
        |   +---------- 日期(相对基准)
        +-------------- 业务码
```

### 3. 性能优化

- ✅ 使用批量生成减少开销
- ✅ 缓存日期编码结果
- ✅ 复用 ID 生成器实例
- ✅ 对于高频场景，考虑使用 `DefaultOrderGenerator` 等预设实例

## 📝 完整示例

示例程序位于 `examples/` 目录：

- `examples/base62_table/base62_table.go` - Base62 进制对应关系表
- `examples/conv_sdk/conv_sdk_example.go` - SDK 完整使用示例（推荐）
- `examples/convert/convert_example.go` - Base62 编码示例
- `examples/timestamp_min/timestamp_min_length.go` - 时间戳最小长度分析
- `examples/timestamp_summary/timestamp_summary.go` - 时间戳编码方案总结
- `examples/shortlink_service/shortlink_service.go` - **短链接服务完整实现**（推荐）⭐

运行示例：

```bash
# 短链接服务示例（推荐，包含性能测试）
cd examples/shortlink_service
go run shortlink_service.go

# SDK 完整使用示例
cd examples/conv_sdk
go run conv_sdk_example.go

# Base62 进制对应关系表
cd examples/base62_table
go run base62_table.go

# Base62 编码示例
cd examples/convert
go run convert_example.go

# 时间戳最小长度分析
cd examples/timestamp_min
go run timestamp_min_length.go

# 时间戳编码方案总结
cd examples/timestamp_summary
go run timestamp_summary.go
```

### 短链接服务示例亮点

`examples/shortlink_service/` 展示了如何构建高性能短链接服务：

- ✅ **完整实现**：包含生成、解析、批量操作
- ✅ **高并发支持**：多生成器实例设计，减少锁竞争
- ✅ **性能测试**：内置性能测试，验证每天1亿写入可行性
- ✅ **读写比 1:10**：支持高并发读取场景

**性能验证**：✅ 每天1亿写入、读写比1:10的需求**完全可行**，性能余量超过1000倍。

> 📖 详细实现和性能分析请查看：[examples/shortlink_service/README.md](examples/shortlink_service/README.md)

## 🧪 测试

运行所有测试：

```bash
go test -v ./...
```

运行特定测试：

```bash
# 基础编码测试
go test -v -run TestBasicEncoding

# 时间戳测试
go test -v -run TestTimestampEncoding

# ID生成测试
go test -v -run TestIDGeneration

# 性能基准测试
go test -v -bench=.
```

## ❓ 常见问题

### Q1: 如何选择合适的 ID 生成方式？

**A:** 根据使用场景选择：
- **订单、支付等需要长期存储**：使用 `QuickOrderID()` 或自定义 `IDGenerator`
- **分享链接、邀请码等临时场景**：使用 `QuickShareCode()` 或 `QuickInviteCode()`
- **缓存键**：使用 `QuickCacheKey()`，自动包含日期便于管理
- **纯数字编码**：使用 `QuickID()`，简单直接

### Q2: 邀请码为什么只有 2 个字符？

**A:** 邀请码设计为固定 2 个字符（B + 1个字符），支持最多 62 个不同用户（0-61）。这样设计的好处：
- ✅ 易于分享和输入
- ✅ 提升用户体验
- ✅ 适合小规模邀请场景

如果需要支持更多用户，可以使用 `QuickShareCode()` 生成更长的分享码。

### Q3: 分享码的有效期如何设置？

**A:** 使用 `GenerateShareCode()` 的第二个参数设置有效期：
```go
// 7天有效
shareCode := shortid.GenerateShareCode(
    shortid.BusinessShare,
    7*24*time.Hour,  // 有效期
    12345,
)

// 1小时有效
shortCode := shortid.GenerateShareCode(
    shortid.BusinessShare,
    1*time.Hour,
    12345,
)
```

### Q4: 如何解析生成的 ID？

**A:** 使用 `IDGenerator` 的 `Parse()` 方法：
```go
gen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessOrder,
    EnableDate:   true,
})

id := gen.Generate()
info, err := gen.Parse(id)
if err == nil {
    fmt.Printf("业务类型: %s\n", info.Business)
    fmt.Printf("时间戳: %v\n", info.Timestamp)
}
```

### Q5: Base62 编码和 Base64 有什么区别？

**A:** 
- **Base62**：使用 `0-9a-zA-Z` 共 62 个字符，URL 安全，无需编码
- **Base64**：使用 `A-Za-z0-9+/` 共 64 个字符，包含 `+` 和 `/`，需要 URL 编码

本库使用 Base62 编码，所有结果都可以直接在 URL 中使用，无需额外编码。

### Q6: 时间戳编码有哪些模式？

**A:** 三种模式：
1. **短编码** (`ToTimestampShort`)：以 2020 年为基准，格式 `天数.秒数`
2. **动态编码** (`ToTimestampDynamic`)：相对当前时间，适合短期场景
3. **日期编码** (`QuickDate`)：相对 2024-12-31，适合日期相关场景

### Q7: 如何批量生成 ID？

**A:** 使用批量 API 或循环生成：
```go
// 方式1：使用批量 API（推荐）
orderIDs := shortid.BatchOrderIDs(100)

// 方式2：循环生成（可自定义）
gen := shortid.NewIDGenerator(config)
for i := 0; i < 100; i++ {
    id := gen.Generate()
    // 使用 id
}
```

### Q8: 自定义业务类型如何使用？

**A:** 使用 `NewIDGenerator()` 配置自定义业务类型：
```go
gen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: 'P',  // 自定义：Post
    EnableDate:   true,
    Prefix:       "POST",  // 可选前缀
})
```

## ⚠️ 注意事项

1. **ID 唯一性**：结合业务码、时间戳和序列号保证全局唯一
2. **动态编码**：需要知道基准时间才能解码，不适合长期存储
3. **安全性**：编码不是加密，敏感数据需要额外保护
4. **兼容性**：Base62 编码是 URL 安全的，可以直接在 URL 中使用
5. **日期基准**：默认日期基准为 `2024-12-31`，可根据业务需求调整
6. **邀请码限制**：邀请码只支持 0-61 的用户ID范围，适合小规模场景
7. **时间戳精度**：compact_mode 包含秒数，但总长度为 7 个字符

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 联系方式

如有问题或建议，请通过以下方式联系：

- 提交 [Issue](https://github.com/gostool/shortid/issues)
- 提交 [Pull Request](https://github.com/gostool/shortid/pulls)

---

**Made with ❤️ by Go Tool Team**
