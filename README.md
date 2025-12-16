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

### 电商系统

```go
// 订单号
orderID := shortid.GenerateOrderID(12345)

// 支付流水
paymentGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessPayment,
    EnableDate:   true,
    MachineID:    100,
})
paymentID := paymentGen.Generate()

// 物流单号
logisticsID := shortid.GenerateShareCode(
    shortid.BusinessLogistics,
    7*24*time.Hour,
    98765,
)
```

### 社交平台

```go
// 会话ID
sessionID := shortid.GenerateSessionID(12345, 2*time.Hour)

// 动态ID
postGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: 'P', // Post
    EnableDate:   true,
})
postID := postGen.Generate()

// 分享链接
shareCode := shortid.GenerateShareCode(
    shortid.BusinessShare,
    7*24*time.Hour,  // 7天有效
    12345,
)
```

### 营销系统

```go
// 优惠券
couponGen := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessCoupon,
    Prefix:       "CPN",
})
couponID := couponGen.Generate()

// 活动码
activityCode := shortid.GenerateShareCode(
    shortid.BusinessActivity,
    30*24*time.Hour,
    520,
)
```

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

运行示例：

```bash
# SDK 完整使用示例（推荐）
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

## ⚠️ 注意事项

1. **ID 唯一性**：结合业务码、时间戳和序列号保证全局唯一
2. **动态编码**：需要知道基准时间才能解码，不适合长期存储
3. **安全性**：编码不是加密，敏感数据需要额外保护
4. **兼容性**：Base62 编码是 URL 安全的，可以直接在 URL 中使用
5. **日期基准**：默认日期基准为 `2024-12-31`，可根据业务需求调整

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
