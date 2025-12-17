# Go分布式唯一ID生成方案

## 📋 目录

1. [方案概述](#方案概述)
   - [核心场景](#核心场景)
2. [ID格式设计](#id格式设计)
3. [接口协议要求](#接口协议要求)
4. [使用示例](#使用示例)
5. [文件结构](#文件结构)
6. [实现检查清单](#实现检查清单)

---

## 方案概述

基于 **Sonyflake 算法**实现分布式唯一ID生成器，支持 **Serverless 部署**，结合时间戳压缩和 Base62 编码生成短ID。

### 核心场景

#### 场景1：分布式唯一ID（Serverless部署）

**问题**：在 Serverless 环境（如 AWS Lambda、阿里云函数计算）中，无法使用固定机器ID，需要动态分配。

**解决方案**：
- 使用 **Redis** 动态分配机器ID（`MachineIDProvider`）
- 使用 **Redis** 管理序列号（`SequenceProvider`，可选）
- 基于 Sonyflake 算法保证全局唯一性
- 支持毫秒级时间戳，高并发场景下性能优异

**特点**：
- ✅ **无需固定配置**：每次函数启动时动态获取机器ID
- ✅ **自动回收**：机器ID和序列号键支持过期时间，自动清理
- ✅ **高可用**：Redis 故障时可降级处理

#### 场景2：生成短ID

**问题**：64位数字ID太长（如 `1734365478123456`），不适合URL、二维码等场景。

**解决方案**：
- 使用 **时间戳压缩算法**（毫秒级）压缩时间戳部分
- 使用 **Base62 编码**编码各部分
- 最终生成 **8-12 字符**的短ID

**特点**：
- ✅ **短小精悍**：8-12 字符（相比64位数字更短）
- ✅ **可读性强**：保留业务类型、时间戳等信息
- ✅ **URL友好**：Base62编码，可直接用于URL

### 技术栈

- **算法**：Sonyflake（毫秒级时间戳）
- **编码**：Base62（已有实现 ✅）
- **时间压缩**：毫秒级时间戳压缩（已有实现 ✅）
- **状态存储**：Redis（分布式机器ID和序列号分配）

---

## ID格式设计

### 位分配

```
64位 = 1位符号(0) + 42位时间戳(毫秒) + 8位业务类型 + 6位机器ID + 7位序列号
```

| 部分 | 位数 | 范围 | 说明 |
|------|------|------|------|
| 符号位 | 1 | 0 | 固定为0 |
| 时间戳 | 42 | 0-4398046511103 | 毫秒级，支持约139年 |
| 业务类型 | 8 | 0-255 | 业务标识 |
| 机器ID | 6 | 0-63 | 支持64台机器 |
| 序列号 | 7 | 0-127 | 128个/毫秒 |

### 短ID生成流程

```
1. 生成64位数字ID（Sonyflake算法）
   ↓
2. 提取各部分：时间戳(毫秒) + 业务类型 + 机器ID + 序列号
   ↓
3. 时间戳压缩（毫秒级）→ 6-9字符 ✅ 已实现
   ↓
4. Base62编码各部分 ✅ 已实现
   ↓
5. 拼接：业务类型 + 压缩时间戳 + 机器ID + 序列号
   ↓
6. 输出短ID（总长度约 8-12 字符）
```

**示例**：
```
数字ID: 1734365478123456
↓ 解析
业务类型: 3, 时间戳: 1704067200000(ms), 机器ID: 1, 序列号: 123
↓ 压缩
压缩时间戳: "000000" (6字符) - 基准时间附近最短
↓ Base62编码
业务类型: "3", 机器ID: "1", 序列号: "1V"
↓ 拼接
短ID: "300000011V" (10字符)
```

### 长度优化分析

**当前格式长度**：
- 业务类型：1-2字符（0-61是1字符，62-255是2字符）
- 时间戳压缩：6-9字符（毫秒级，基准时间附近6字符，越远越长）
- 机器ID：1-2字符（0-61是1字符，62-63是2字符）
- 序列号：1-2字符（0-61是1字符，62-127是2字符）
- **总长度：9-15字符**（基准时间附近最短约9-10字符）

**优化方案**：

| 方案 | 优化方法 | 最短长度 | 节省 | 缺点 |
|------|---------|---------|------|------|
| **方案A（当前）** | 分别编码各部分 | 9-10字符 | - | - |
| **方案B（合并编码）** | 合并业务+机器+序列 | 8-9字符 | 1字符 | 失去部分可读性 |
| **方案C（秒级时间戳）** | 使用秒级时间戳压缩 | 7-8字符 | 2字符 | **失去毫秒精度** |
| **方案D（整体合并）** | 所有部分合并编码 | 6-7字符 | 3-4字符 | **完全失去可读性** |

**推荐**：
- ✅ **方案A（当前）**：平衡了长度和可读性，推荐使用
- ⚠️ **方案B**：如果追求极致长度，可考虑合并编码（节省1字符）
- ❌ **方案C**：不推荐，会失去毫秒精度（Sonyflake的核心优势）
- ❌ **方案D**：不推荐，完全失去可读性，难以调试

**实际长度分布**（基于当前方案）：
- 基准时间附近（2024-2025年）：9-10字符
- 1年后：10-11字符
- 5年后：11-12字符
- 10年后：12-13字符

---

## 接口协议要求

### MachineIDProvider 接口

机器ID提供者接口，用于 Serverless 环境动态分配机器ID。

**接口定义**：

```go
type MachineIDProvider interface {
    // GetMachineID 获取机器ID（原子递增，取模64）
    // 要求：必须保证原子性，返回 0-63 范围内的机器ID
    // 参数：
    //   - ctx: 上下文
    // 返回：
    //   - uint16: 机器ID（0-63）
    //   - error: 如果操作失败，返回错误
    GetMachineID(ctx context.Context) (uint16, error)
    
    // SetMachineIDExpiration 设置机器ID过期时间
    // 要求：用于 Serverless 环境，设置机器ID的过期时间，支持自动回收
    // 参数：
    //   - ctx: 上下文
    //   - machineID: 机器ID
    //   - expiration: 过期时间，建议设置为函数最大运行时间 + 缓冲时间（例如：20分钟）
    // 返回：
    //   - error: 如果操作失败，返回错误
    SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error
    
    // HealthCheck 健康检查
    // 要求：验证连接是否可用，在ID生成前检查存储系统是否正常
    // 参数：
    //   - ctx: 上下文
    // 返回：
    //   - error: 如果连接不可用，返回错误
    HealthCheck(ctx context.Context) error
    
    // Close 关闭连接，释放资源
    // 要求：在程序退出或资源释放时调用，释放相关资源
    // 返回：
    //   - error: 如果关闭失败，返回错误
    Close() error
}
```

**实现要求**：

1. **MemoryMachineIDProvider**（测试用）：
   - 使用内存存储，无需外部依赖
   - 使用 `sync.Mutex` 保证并发安全
   - 原子递增计数器，取模64

2. **RedisMachineIDProvider**（生产用）：
   - 使用 Redis INCR 命令实现原子递增
   - 创建时自动测试连接（Ping）
   - 支持自定义 Redis 选项
   - 使用 `shortid:machine:id` 键，原子递增，取模64
   - 错误处理使用 `fmt.Errorf` 和 `%w` 包装

### SequenceProvider 接口

序列号提供者接口，用于分布式环境生成序列号。

**接口定义**：

```go
type SequenceProvider interface {
    // GetSequence 获取序列号（原子递增，取模128）
    // 要求：必须保证原子性，返回 0-127 范围内的序列号
    // 参数：
    //   - ctx: 上下文
    //   - key: 序列号键名，通常基于时间戳（10ms单位）生成唯一键
    // 返回：
    //   - uint16: 序列号（0-127）
    //   - error: 如果操作失败，返回错误
    // 说明：
    //   - 序列号在同一时间单位（10ms）内递增
    //   - 不同时间单位使用不同的 key，序列号从0开始
    //   - 如果序列号溢出（达到128），调用方需要等待下一时间单位
    GetSequence(ctx context.Context, key string) (uint16, error)
    
    // SetSequenceExpiration 设置序列号键的过期时间
    // 要求：用于清理过期的序列号键，避免内存/存储泄漏
    // 参数：
    //   - ctx: 上下文
    //   - key: 序列号键名
    //   - expiration: 过期时间，建议设置为时间单位的2-3倍（例如：30ms）
    // 返回：
    //   - error: 如果操作失败，返回错误
    SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error
    
    // HealthCheck 健康检查
    // 要求：验证连接是否可用，在ID生成前检查存储系统是否正常
    // 参数：
    //   - ctx: 上下文
    // 返回：
    //   - error: 如果连接不可用，返回错误
    HealthCheck(ctx context.Context) error
    
    // Close 关闭连接，释放资源
    // 要求：在程序退出或资源释放时调用，释放相关资源
    // 返回：
    //   - error: 如果关闭失败，返回错误
    Close() error
}
```

**实现要求**：

1. **MemorySequenceProvider**（测试用）：
   - 使用内存存储，无需外部依赖
   - 使用 `sync.Mutex` 保证并发安全
   - 使用 `map[string]int64` 存储每个时间单位的序列号
   - 原子递增，取模128

2. **RedisSequenceProvider**（生产用）：
   - 使用 Redis INCR 命令实现原子递增
   - 创建时自动测试连接（Ping）
   - 支持自定义 Redis 选项
   - 使用 `shortid:sequence:{key}` 键，原子递增，取模128
   - 错误处理使用 `fmt.Errorf` 和 `%w` 包装

### Generator 接口

ID生成器接口，基于 Sonyflake 算法生成短ID。

**核心要求**：

1. **时间单位**：使用 10ms 单位（与 Sonyflake 一致）
2. **位分配**：42位时间戳 + 8位业务类型 + 6位机器ID + 7位序列号
3. **基准时间**：默认 2024-01-01 00:00:00 UTC（`DefaultSnowflakeEpochMs`）
4. **序列号管理**：
   - **本地模式**（默认）：同一时间单位内序列号递增，溢出时等待下一时间单位
   - **分布式模式**（可选）：使用 `SequenceProvider.GetSequence` 获取序列号
     - 序列号键名：基于时间戳（10ms单位）生成，例如：`shortid:sequence:{elapsedTime}`
     - 序列号范围：0-127（128个/10ms）
     - 如果序列号溢出（达到128），等待下一时间单位
     - 自动设置序列号键的过期时间（建议30ms）
5. **短ID转换**：
   - 提取各部分：时间戳(毫秒) + 业务类型 + 机器ID + 序列号
   - 时间戳压缩：使用 `ToTimestampShortMs`（毫秒级）
   - Base62编码：使用 `EncodeBase62` 编码各部分
   - 拼接：业务类型 + 压缩时间戳 + 机器ID + 序列号

**配置结构**：

```go
type Config struct {
    MachineID        uint16            // 固定机器ID（传统部署）
    BusinessType     BusinessType      // 业务类型
    MachineIDProvider MachineIDProvider // Serverless模式机器ID提供者（与MachineID二选一）
    SequenceProvider SequenceProvider   // 分布式序列号提供者（可选，默认使用本地序列号）
    StartTime        time.Time         // 基准时间（可选，默认2024-01-01）
}
```

**错误定义**：

- `ErrOverTimeLimit`：时间戳溢出错误（超过42位时间戳范围）

**参考实现**：

- Sonyflake 算法：https://github.com/sony/sonyflake/blob/master/sonyflake.go
- 时间戳压缩：`ToTimestampShortMs` / `FromTimestampShortMs`（已实现 ✅）
- Base62编码：`EncodeBase62` / `DecodeBase62`（已实现 ✅）

---

## 使用示例

### 场景1：分布式唯一ID（Serverless部署）

#### 示例1.1：完整Serverless模式（机器ID + 序列号都使用Redis）

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/gostool/shortid"
)

func handler(ctx context.Context) (string, error) {
    // 创建Redis机器ID提供者
    machineProvider, err := shortid.NewRedisMachineIDProvider("localhost:6379")
    if err != nil {
        return "", fmt.Errorf("failed to create machine provider: %w", err)
    }
    defer machineProvider.Close()

    // 创建Redis序列号提供者（可选，用于高并发场景）
    sequenceProvider, err := shortid.NewRedisSequenceProvider("localhost:6379")
    if err != nil {
        return "", fmt.Errorf("failed to create sequence provider: %w", err)
    }
    defer sequenceProvider.Close()

    // 创建ID生成器
    generator, err := shortid.NewGenerator(shortid.Config{
        MachineIDProvider: machineProvider,
        SequenceProvider:  sequenceProvider, // 使用分布式序列号
        BusinessType:      shortid.BusinessOrder,
    })
    if err != nil {
        return "", fmt.Errorf("failed to create generator: %w", err)
    }

    // 生成唯一ID（短ID格式）
    id, err := generator.GenerateWithContext(ctx)
    if err != nil {
        return "", fmt.Errorf("failed to generate ID: %w", err)
    }

    return id, nil // 返回短ID，如 "300000011V"
}
```

#### 示例1.2：简化Serverless模式（仅机器ID使用Redis）

```go
// 只使用Redis分配机器ID，序列号使用本地模式（适合大多数场景）
machineProvider, err := shortid.NewRedisMachineIDProvider("localhost:6379")
if err != nil {
    log.Fatal(err)
}
defer machineProvider.Close()

generator, err := shortid.NewGenerator(shortid.Config{
    MachineIDProvider: machineProvider,
    // SequenceProvider 不设置，使用默认的本地序列号
    BusinessType: shortid.BusinessOrder,
})
if err != nil {
    log.Fatal(err)
}

id, err := generator.GenerateWithContext(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(id) // 输出短ID
```

### 场景2：生成短ID

#### 示例2.1：传统部署（固定机器ID）

```go
// 传统服务器部署，使用固定机器ID
generator, err := shortid.NewGenerator(shortid.Config{
    MachineID:    1, // 固定机器ID
    BusinessType: shortid.BusinessOrder,
})
if err != nil {
    log.Fatal(err)
}

// 生成短ID
id, err := generator.Generate()
if err != nil {
    log.Fatal(err)
}
fmt.Println(id) // 输出: "300000011V"（示例，实际长度8-12字符）
```

#### 示例2.2：测试环境（内存实现）

```go
// 测试环境，使用内存实现（无需Redis）
machineProvider := shortid.NewMemoryMachineIDProvider()
defer machineProvider.Close()

generator, err := shortid.NewGenerator(shortid.Config{
    MachineIDProvider: machineProvider,
    BusinessType:     shortid.BusinessOrder,
})
if err != nil {
    log.Fatal(err)
}

id, err := generator.GenerateWithContext(ctx)
if err != nil {
    log.Fatal(err)
}
fmt.Println(id) // 输出短ID
```

---

## 文件结构规划

### 目录结构

```
shortid/
├── provider.go                    # ✅ MachineIDProvider 和 SequenceProvider 接口定义
├── generator.go                   # ✅ Generator 核心实现
├── generator_test.go              # ✅ Generator 测试
├── errors.go                      # ✅ 错误定义
│
├── machineid/                     # 机器ID提供者模块
│   ├── memory.go                  # ✅ MemoryMachineIDProvider 实现（测试用）
│   ├── redis.go                   # ✅ RedisMachineIDProvider 实现（生产用）
│   └── memory_test.go             # ✅ 机器ID提供者测试
│
├── sequence/                      # 序列号提供者模块
│   ├── memory.go                  # ✅ MemorySequenceProvider 实现（测试用）
│   ├── redis.go                   # ✅ RedisSequenceProvider 实现（生产用）
│   └── memory_test.go             # ✅ 序列号提供者测试
│
├── base.go                        # ✅ Base62编码
├── timestamp.go                   # ✅ 时间戳压缩
├── const.go                       # ✅ 常量定义
└── logic_enum.go                 # ✅ 业务类型定义
```

### 文件说明

| 文件 | 职责 | 状态 |
|------|------|------|
| `provider.go` | MachineIDProvider 和 SequenceProvider 接口定义 | ✅ 已完成 |
| `machineid/memory.go` | MemoryMachineIDProvider 实现 | ✅ 已完成 |
| `machineid/redis.go` | RedisMachineIDProvider 实现 | ✅ 已完成 |
| `sequence/memory.go` | MemorySequenceProvider 实现 | ✅ 已完成 |
| `sequence/redis.go` | RedisSequenceProvider 实现 | ✅ 已完成 |
| `generator.go` | Generator 核心实现 | ✅ 已完成 |
| `errors.go` | 错误定义 | ✅ 已完成 |
| `machineid/memory_test.go` | 机器ID提供者测试 | ✅ 已完成 |
| `sequence/memory_test.go` | 序列号提供者测试 | ✅ 已完成 |
| `generator_test.go` | Generator 测试 | ✅ 已完成 |

**注意**：旧的 `provider.go` 和 `provider_redis.go` 将在新实现完成后删除。

---

## 实现检查清单

### 基础功能（已完成 ✅）

- [x] Base62编码
- [x] 毫秒级时间戳压缩
- [x] 业务类型定义
- [x] Snowflake常量定义

### 核心功能（已完成 ✅）

- [x] MachineIDProvider接口定义 ✅
  - [x] MemoryMachineIDProvider实现（测试用）✅
  - [x] RedisMachineIDProvider实现（生产用）✅
- [x] SequenceProvider接口定义 ✅
  - [x] MemorySequenceProvider实现（测试用）✅
  - [x] RedisSequenceProvider实现（生产用）✅
- [x] Generator核心实现 ✅
  - [x] 64位ID生成（Sonyflake算法）✅
  - [x] 短ID转换（时间戳压缩 + Base62编码）✅
  - [x] 支持本地序列号和分布式序列号两种模式 ✅
- [x] 错误定义 ✅
- [x] 单元测试 ✅

### 待完善功能

- [ ] 集成测试（需要Redis环境）
- [ ] 性能基准测试
- [ ] 文档完善

---

**最后更新**：2024-12-16
