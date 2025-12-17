# Go分布式唯一ID生成方案调研

## 📋 目录

1. [方案概述](#方案概述)
2. [技术选型](#技术选型)
3. [分布式状态存储协议](#分布式状态存储协议)
4. [Serverless ID 生成器设计](#serverless-id-生成器设计)
5. [机器ID分配策略](#机器id分配策略)
6. [部署模式](#部署模式)
7. [实现进度](#实现进度)
8. [使用示例](#使用示例)

---

## 方案概述

本方案旨在实现一个支持 **Serverless 部署**的分布式唯一ID生成器，基于 **Sonyflake** 算法改造，结合 Redis 实现分布式状态管理，并利用时间戳压缩和 Base62 编码技术生成短小精悍的ID。

### 核心目标

- ✅ **支持 Serverless 部署**：无需固定机器ID，动态分配
- ✅ **全局唯一性**：基于 Sonyflake 算法保证分布式环境下的唯一性
- ✅ **短ID生成**：结合时间戳压缩和 Base62 编码，生成 4-10 字符的短ID
- ✅ **业务语义**：支持业务类型标识，便于业务区分和管理
- ✅ **高可用性**：支持 ECS、K8s、Serverless 多种部署模式

### 技术栈

- **基础算法**：Sonyflake（分布式唯一ID生成，**毫秒级时间戳**）
- **状态存储**：Redis（分布式状态管理）
- **编码技术**：Base62 编码（已有实现）
- **时间压缩**：时间戳压缩算法（**需要扩展支持毫秒级**）

### 关键问题

⚠️ **时间戳精度不匹配**：
- Sonyflake 生成的是**毫秒级时间戳**（42位，毫秒精度）
- 现有时间戳压缩算法只支持**秒级时间戳**
- **需要扩展时间戳压缩算法支持毫秒级**，才能与 Sonyflake 整合

---

## 技术选型

### 分布式雪花ID方案对比

| 方案 | 项目地址 | 特点 | 适用场景 |
|------|---------|------|---------|
| **Snowflake** | https://github.com/bwmarrin/snowflake | 需要手动配置机器ID | 传统服务器部署 |
| **Sonyflake** | https://github.com/sony/sonyflake | 自动生成机器ID，但依赖私有IP | Serverless 不友好 |
| **本方案** | - | 基于 Sonyflake + Redis 动态分配机器ID | **Serverless 友好** ✅ |

### 选择 Sonyflake 的原因

1. **算法成熟**：Sonyflake 是 Snowflake 的改进版，算法稳定可靠
2. **自动机器ID**：相比 Snowflake 需要手动配置，Sonyflake 可以自动生成
3. **易于改造**：通过 Redis 替换机器ID生成逻辑，即可支持 Serverless

---

## 分布式状态存储协议

### 接口定义

```go
package shortid

import (
	"context"
	"time"
)

// DistributedStateProvider 分布式状态提供者协议
// 定义了ID生成器所需的分布式状态操作能力
// 支持多种实现：Redis、内存（测试）、其他存储系统
type DistributedStateProvider interface {
	// IncrementCounter 原子递增计数器，返回新值
	// 用于生成序列号、机器ID等需要原子递增的场景
	IncrementCounter(ctx context.Context, key string) (int64, error)
	
	// IncrementCounterBy 原子增加计数器指定值，返回新值
	// 用于批量分配或指定步长的递增操作
	IncrementCounterBy(ctx context.Context, key string, value int64) (int64, error)

	// IncrementCounterRand 原子增加计数器随机值，返回新值
	// 用于生成随机序列号，避免序列号可预测
	// min: 最小值（包含），max: 最大值（不包含）
	IncrementCounterRand(ctx context.Context, key string, min int64, max int64) (int64, error)
	
	// SetExpiration 设置键的过期时间
	// 用于设置机器ID等临时资源的过期时间，支持自动回收
	SetExpiration(ctx context.Context, key string, expiration time.Duration) error
	
	// HealthCheck 健康检查，用于验证连接是否可用
	// 在ID生成前检查存储系统是否正常，避免生成失败
	HealthCheck(ctx context.Context) error
	
	// Close 关闭连接，释放资源
	// 在程序退出或资源释放时调用
	Close() error
}
```

### 接口设计说明

1. **原子操作保证**：所有计数器操作都是原子的，确保并发安全
2. **过期机制**：支持设置键的过期时间，适合 Serverless 环境自动回收资源
3. **健康检查**：提供健康检查接口，便于监控和故障处理
4. **资源管理**：提供 Close 接口，支持优雅关闭

### 实现要求

- **Redis 实现**：用于生产环境，支持分布式部署
- **内存实现**：用于测试环境，无需外部依赖
- **其他实现**：可扩展支持其他存储系统（如 etcd、Consul 等）

---

## Serverless ID 生成器设计

### 架构设计

```
┌─────────────────────────────────────────┐
│      Serverless ID Generator            │
│  (基于 Sonyflake + 时间戳压缩)          │
│                                         │
│  - 业务类型标识                         │
│  - 时间戳压缩编码                       │
│  - Base62 编码输出                      │
└──────────────┬──────────────────────────┘
               │
               │ 使用
               ▼
┌─────────────────────────────────────────┐
│   DistributedStateProvider (接口)        │
│  - IncrementCounter (获取机器ID)        │
│  - IncrementCounterRand (生成序列号)    │
│  - SetExpiration (设置过期时间)         │
└──────────────┬──────────────────────────┘
               │
       ┌───────┴───────┐
       │               │
       ▼               ▼
┌─────────────┐  ┌─────────────┐
│ Redis实现    │  │ 内存实现     │
│ (生产环境)   │  │ (测试环境)   │
└─────────────┘  └─────────────┘
```

### ID 格式设计

基于 Sonyflake 算法的位分配：

```
64位 = 1位符号(0) + 42位时间戳(毫秒) + 8位业务类型 + 6位机器ID + 7位序列号
```

**各部分说明**：

| 部分 | 位数 | 范围 | 说明 |
|------|------|------|------|
| **符号位** | 1 | 0 | 固定为0，保证为正数 |
| **时间戳** | 42 | 0-4398046511103 | 毫秒级时间戳，支持约139年 |
| **业务类型** | 8 | 0-255 | 业务类型标识，支持256种业务 |
| **机器ID** | 6 | 0-63 | 机器标识，支持64台机器 |
| **序列号** | 7 | 0-127 | 同一毫秒内的序列号，支持128个/毫秒 |

### 短ID生成流程

1. **生成数字ID**：使用 Sonyflake 算法生成64位数字ID（毫秒级时间戳）
2. **时间戳压缩**：提取毫秒级时间戳部分，使用**毫秒级时间戳压缩算法**压缩
3. **Base62编码**：将压缩后的各部分进行 Base62 编码
4. **拼接输出**：拼接业务类型 + 压缩时间戳 + 机器ID + 序列号

**示例**：
```
数字ID: 1734365478123456
↓ 解析（64位数字）
业务类型: 3 (订单)
时间戳: 1704067200000 (毫秒，2024-01-01 00:00:00.000)
机器ID: 1
序列号: 123
↓ 时间戳压缩（毫秒级）
压缩时间戳: "+5E00000" (天数+毫秒数，6-9字符)
↓ Base62编码
短ID: "3+5E00000123" (业务类型+压缩时间戳+机器ID+序列号)
```

**注意**：当前时间戳压缩算法只支持秒级，需要扩展支持毫秒级压缩。

---

## 机器ID分配策略

### Serverless 模式

在 Serverless 环境（如 AWS Lambda、阿里云函数计算）中，每次函数启动时动态分配机器ID：

#### 分配流程

1. **获取机器ID**：
   ```go
   // 使用 Redis INCR 原子递增获取唯一ID
   machineID := redis.IncrementCounter("shortid:machine:id") % 64
   ```

2. **设置过期时间**：
   ```go
   // 设置机器ID的过期时间为函数最大运行时间 + 缓冲时间
   // 例如：15分钟（Lambda最大运行时间）+ 5分钟缓冲 = 20分钟
   redis.SetExpiration("shortid:machine:id:" + machineID, 20 * time.Minute)
   ```

3. **心跳续期**（可选）：
   ```go
   // 如果函数运行时间较长，定期续期机器ID
   // 避免机器ID被回收导致冲突
   ```

#### 并发处理

- **原子操作**：使用 Redis INCR 保证原子性，避免并发冲突
- **取模运算**：`machineID % 64` 确保机器ID在有效范围内（0-63）
- **过期回收**：通过过期时间自动回收不再使用的机器ID

#### 冲突处理

- **机器ID冲突**：虽然概率极低，但通过序列号和时间戳可以保证唯一性
- **Redis 故障**：提供降级方案，使用随机机器ID或固定机器ID

### 传统部署模式

在 ECS、K8s 等传统部署环境中，可以使用固定机器ID：

#### ECS 部署

```go
// 通过环境变量配置机器ID
machineID := os.Getenv("SHORTID_MACHINE_ID")
```

#### K8s 部署

```go
// 通过 StatefulSet 的序号作为机器ID
// 或通过 ConfigMap 配置
machineID := getMachineIDFromK8s()
```

---

## 部署模式

### 1. ECS 部署

**特点**：
- 固定机器ID，通过环境变量配置
- 无需 Redis，使用本地状态
- 适合传统服务器部署

**配置示例**：
```bash
export SHORTID_MACHINE_ID=1
export SHORTID_BUSINESS_TYPE=3
```

### 2. K8s 部署

**特点**：
- 通过 StatefulSet 序号或 ConfigMap 配置机器ID
- 支持水平扩展，每 Pod 一个机器ID
- 可选使用 Redis 进行状态同步

**配置示例**：
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: shortid-config
data:
  machine-id: "1"
  business-type: "3"
```

### 3. Serverless 部署

**特点**：
- 动态分配机器ID，无需手动配置
- 依赖 Redis 进行状态管理
- 支持自动扩缩容

**配置示例**：
```go
// Lambda 函数初始化
provider := NewRedisProvider(redisURL)
generator := NewServerlessIDGenerator(ServerlessConfig{
    Provider: provider,
    BusinessType: BusinessOrder,
})
```

---

## 实现进度

### ✅ 已完成

- [x] Base62/Base36/Base58 编码解码（`base.go`）
- [x] 时间戳压缩算法（`timestamp.go`）- **秒级精度**
  - [x] 短编码算法（Short Encoding）- 秒级
  - [x] 动态编码算法（Dynamic Encoding）- 秒级
  - [x] 紧凑编码算法（Compact Encoding）- 秒级
- [x] 业务类型定义（`logic_enum.go`）
- [x] Snowflake 常量定义（`const.go`）- 毫秒级常量已定义
- [x] Sonyflake 依赖引入（`go.mod`）

### 🚧 待实现

#### 核心功能
- [ ] **毫秒级时间戳压缩算法** ⚠️ **关键缺失**
  - [ ] 短编码算法（Short Encoding）- 毫秒级版本
  - [ ] 紧凑编码算法（Compact Encoding）- 毫秒级版本
  - [ ] 与秒级算法的兼容性设计
- [ ] `DistributedStateProvider` 接口定义（代码中）
- [ ] Redis 实现 `DistributedStateProvider`
- [ ] 内存实现 `DistributedStateProvider`（测试用）
- [ ] Serverless ID 生成器实现
  - [ ] 毫秒级时间戳提取和压缩
  - [ ] 机器ID分配逻辑
  - [ ] 序列号生成逻辑

#### 辅助功能
- [ ] 错误处理和降级方案
- [ ] 单元测试和集成测试
- [ ] 性能基准测试

---

## 使用示例

### 示例1：Serverless 环境（Lambda）

```go
package main

import (
    "context"
    "github.com/gostool/shortid"
)

func handler(ctx context.Context) (string, error) {
    // 初始化 Redis Provider
    provider, err := shortid.NewRedisProvider("redis://localhost:6379")
    if err != nil {
        return "", err
    }
    defer provider.Close()
    
    // 创建 Serverless ID 生成器
    generator := shortid.NewServerlessIDGenerator(shortid.ServerlessConfig{
        Provider:     provider,
        BusinessType: shortid.BusinessOrder,
    })
    
    // 生成订单ID
    id, err := generator.Generate(ctx)
    if err != nil {
        return "", err
    }
    
    return id, nil
}
```

### 示例2：传统部署（ECS/K8s）

```go
package main

import (
    "github.com/gostool/shortid"
)

func main() {
    // 使用固定机器ID
    generator := shortid.NewIDGenerator(shortid.IDGeneratorConfig{
        MachineID:    1,
        BusinessType: shortid.BusinessOrder,
    })
    
    // 生成订单ID
    id := generator.Generate()
    fmt.Println("订单ID:", id) // 输出: "3+5E3d7"
}
```

### 示例3：测试环境（内存实现）

```go
package main

import (
    "github.com/gostool/shortid"
)

func TestIDGeneration(t *testing.T) {
    // 使用内存 Provider（无需 Redis）
    provider := shortid.NewMemoryProvider()
    defer provider.Close()
    
    generator := shortid.NewServerlessIDGenerator(shortid.ServerlessConfig{
        Provider:     provider,
        BusinessType: shortid.BusinessOrder,
    })
    
    id, err := generator.Generate(context.Background())
    assert.NoError(t, err)
    assert.NotEmpty(t, id)
}
```

---

## 相关文档

- [进制转换实现](./step1:进制转换.md) - Base62/Base36/Base58 编码实现
- [时间戳压缩方案](./step2:时间戳压缩方案.md) - 三种时间戳压缩算法详解（**包含毫秒级精度设计说明**）
- [核心算法分析](./CORE_ALGO.md) - 完整的算法分析和性能测试
- [需求SDK文档](./需求sdk.md) - SDK 功能和使用说明

## 整合思路总结

### 核心流程

```
Sonyflake 生成64位数字ID（毫秒级时间戳）
    ↓
提取各部分：时间戳(毫秒) + 业务类型 + 机器ID + 序列号
    ↓
时间戳压缩（毫秒级）→ 短字符串（6-9字符）
    ↓
Base62 编码各部分
    ↓
拼接：业务类型 + 压缩时间戳 + 机器ID + 序列号
    ↓
输出短ID（总长度约 8-12 字符）
```

### 关键依赖

1. **毫秒级时间戳压缩**：必须实现，否则无法整合
2. **DistributedStateProvider**：用于 Serverless 模式获取机器ID
3. **Sonyflake 算法**：已有依赖，需要封装使用

### 实现优先级

1. **P0（阻塞）**：毫秒级时间戳压缩算法
2. **P1（核心）**：DistributedStateProvider 接口和 Redis 实现
3. **P2（功能）**：Serverless ID 生成器实现
4. **P3（完善）**：测试、文档、性能优化

---

## 技术难点与解决方案

### 1. 时间戳精度匹配 ⚠️ **关键问题**

**问题**：
- Sonyflake 使用毫秒级时间戳（42位）
- 现有时间戳压缩算法只支持秒级

**解决方案**：
- 扩展时间戳压缩算法，支持毫秒级精度
- 参考 `step2:时间戳压缩方案.md` 中的毫秒级精度设计
- 毫秒级短编码：天数（1-4字符）+ 毫秒数（5字符）= 6-9字符
- 保持与秒级算法的接口一致性，通过参数区分精度

**实现要点**：
```go
// 秒级版本（已有）
ToTimestampShort(ts int64) string  // ts 为秒级时间戳

// 毫秒级版本（待实现）
ToTimestampShortMs(tsMs int64) string  // tsMs 为毫秒级时间戳
```

### 2. Redis 连接管理

- **连接池**：使用连接池管理 Redis 连接，避免频繁创建连接
- **超时设置**：设置合理的超时时间，避免阻塞
- **重试机制**：实现重试逻辑，提高可用性

### 3. 机器ID冲突

- **概率极低**：通过时间戳和序列号可以保证唯一性
- **监控告警**：监控机器ID分配情况，及时发现异常
- **降级方案**：Redis 故障时使用随机机器ID或固定机器ID

### 4. 时钟同步

- **NTP 同步**：确保服务器时钟同步，避免时间戳回退
- **时钟回退处理**：实现时钟回退检测和处理逻辑（已有常量 `MaxClockBackwardMs`）
- **时间戳验证**：生成ID时验证时间戳的合理性

---

**最后更新**：2024-12-16
