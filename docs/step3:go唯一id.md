# Go分布式唯一ID生成方案

## 📋 目录

1. [方案概述](#方案概述)
2. [ID格式设计](#id格式设计)
3. [实现步骤](#实现步骤)
4. [使用示例](#使用示例)

---

## 方案概述

基于 **Sonyflake 算法**实现分布式唯一ID生成器，支持 **Serverless 部署**，结合时间戳压缩和 Base62 编码生成短ID。

### 核心特性

- ✅ **短ID生成**：8-12 字符的短ID（相比64位数字更短）
- ✅ **Serverless 友好**：动态分配机器ID，无需固定配置
- ✅ **业务语义**：支持业务类型标识（8位，256种业务）
- ✅ **毫秒精度**：42位毫秒级时间戳，支持约139年

### 技术栈

- **算法**：Sonyflake（毫秒级时间戳）
- **编码**：Base62（已有实现 ✅）
- **时间压缩**：毫秒级时间戳压缩（已有实现 ✅）
- **状态存储**：Redis（分布式机器ID分配）

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

## 实现步骤

### Step 1: 定义分布式状态接口

**文件**：`provider.go`（新建）

```go
package shortid

import (
    "context"
    "time"
)

// StateProvider 分布式状态提供者（简化版）
type StateProvider interface {
    // GetMachineID 获取机器ID（原子递增，取模64）
    GetMachineID(ctx context.Context) (uint16, error)
    
    // SetMachineIDExpiration 设置机器ID过期时间
    SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error
    
    // HealthCheck 健康检查
    HealthCheck(ctx context.Context) error
    
    // Close 关闭连接
    Close() error
}
```

### Step 2: 实现内存 Provider（测试用）

**文件**：`provider.go`（追加）

```go
// MemoryProvider 内存实现（测试用）
type MemoryProvider struct {
    counter int64
    mu      sync.Mutex
}

func NewMemoryProvider() *MemoryProvider {
    return &MemoryProvider{}
}

func (m *MemoryProvider) GetMachineID(ctx context.Context) (uint16, error) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.counter++
    return uint16(m.counter % 64), nil
}

func (m *MemoryProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
    return nil // 内存实现无需过期
}

func (m *MemoryProvider) HealthCheck(ctx context.Context) error {
    return nil
}

func (m *MemoryProvider) Close() error {
    return nil
}
```

### Step 3: 实现 Redis Provider（生产用）

**文件**：`provider_redis.go`（新建）

```go
package shortid

import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisProvider struct {
    client *redis.Client
}

func NewRedisProvider(addr string) (*RedisProvider, error) {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    return &RedisProvider{client: client}, nil
}

func (r *RedisProvider) GetMachineID(ctx context.Context) (uint16, error) {
    // 原子递增，取模64
    val, err := r.client.Incr(ctx, "shortid:machine:id").Result()
    if err != nil {
        return 0, err
    }
    return uint16(val % 64), nil
}

func (r *RedisProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
    key := fmt.Sprintf("shortid:machine:id:%d", machineID)
    return r.client.Expire(ctx, key, expiration).Err()
}

func (r *RedisProvider) HealthCheck(ctx context.Context) error {
    return r.client.Ping(ctx).Err()
}

func (r *RedisProvider) Close() error {
    return r.client.Close()
}
```

### Step 4: 实现 ID 生成器（核心）

**文件**：`generator.go`（新建）

基于 Sonyflake 算法，参考 https://github.com/sony/sonyflake/blob/master/sonyflake.go

```go
package shortid

import (
    "context"
    "sync"
    "time"
)

// Generator ID生成器
type Generator struct {
    mu          sync.Mutex
    startTime   int64  // 基准时间（毫秒）
    elapsedTime int64  // 已过时间（10ms单位）
    sequence    uint16 // 序列号
    machineID   uint16 // 机器ID
    businessType uint8 // 业务类型
    
    // Serverless模式
    provider    StateProvider
    useProvider bool
}

// Config 生成器配置
type Config struct {
    // 固定机器ID（传统部署）
    MachineID uint16
    
    // 业务类型
    BusinessType BusinessType
    
    // Serverless模式（二选一）
    Provider StateProvider
    
    // 基准时间（可选，默认2024-01-01）
    StartTime time.Time
}

// NewGenerator 创建ID生成器
func NewGenerator(config Config) (*Generator, error) {
    g := &Generator{
        businessType: uint8(config.BusinessType),
        sequence:     SnowflakeMaxSequence, // 初始化为最大值，首次生成时重置为0
    }
    
    // 设置基准时间
    if config.StartTime.IsZero() {
        g.startTime = DefaultSnowflakeEpochMs / 10 // 转换为10ms单位
    } else {
        g.startTime = config.StartTime.UnixMilli() / 10
    }
    
    // 机器ID分配
    if config.Provider != nil {
        g.provider = config.Provider
        g.useProvider = true
        // 延迟获取机器ID（首次生成时获取）
    } else {
        g.machineID = config.MachineID
        g.useProvider = false
    }
    
    return g, nil
}

// Generate 生成ID（固定机器ID模式）
func (g *Generator) Generate() (string, error) {
    return g.GenerateWithContext(context.Background())
}

// GenerateWithContext 生成ID（支持Serverless模式）
func (g *Generator) GenerateWithContext(ctx context.Context) (string, error) {
    // Serverless模式：首次获取机器ID
    if g.useProvider && g.machineID == 0 {
        machineID, err := g.provider.GetMachineID(ctx)
        if err != nil {
            return "", err
        }
        g.machineID = machineID
        // 设置过期时间（20分钟）
        _ = g.provider.SetMachineIDExpiration(ctx, machineID, 20*time.Minute)
    }
    
    // 生成64位数字ID
    id, err := g.nextID()
    if err != nil {
        return "", err
    }
    
    // 转换为短ID
    return g.toShortID(id)
}

// nextID 生成64位数字ID（基于Sonyflake算法）
func (g *Generator) nextID() (uint64, error) {
    const maskSequence = uint16(1<<SnowflakeSequenceBits - 1)
    const timeUnit = 10 // 10ms单位（Sonyflake使用10ms）
    
    g.mu.Lock()
    defer g.mu.Unlock()
    
    // 计算当前已过时间（10ms单位）
    now := time.Now().UnixMilli() / timeUnit
    current := now - g.startTime
    
    if g.elapsedTime < current {
        g.elapsedTime = current
        g.sequence = 0
    } else {
        g.sequence = (g.sequence + 1) & maskSequence
        if g.sequence == 0 {
            // 序列号溢出，等待下一时间单位
            g.elapsedTime++
            overtime := g.elapsedTime - current
            time.Sleep(time.Duration(overtime*timeUnit) * time.Millisecond)
        }
    }
    
    // 检查时间溢出
    if g.elapsedTime >= 1<<SnowflakeTimestampBits {
        return 0, ErrOverTimeLimit
    }
    
    // 组装64位ID
    return uint64(g.elapsedTime)<<(SnowflakeBusinessShift+SnowflakeMachineShift+SnowflakeSequenceBits) |
        uint64(g.businessType)<<(SnowflakeMachineShift+SnowflakeSequenceBits) |
        uint64(g.machineID)<<SnowflakeSequenceBits |
        uint64(g.sequence), nil
}

// toShortID 转换为短ID
func (g *Generator) toShortID(id uint64) (string, error) {
    // 提取各部分
    elapsedTime := id >> (SnowflakeBusinessShift + SnowflakeMachineShift + SnowflakeSequenceBits)
    businessType := uint8((id >> (SnowflakeMachineShift + SnowflakeSequenceBits)) & SnowflakeMaxBusiness)
    machineID := uint16((id >> SnowflakeSequenceBits) & SnowflakeMaxMachine)
    sequence := uint16(id & SnowflakeMaxSequence)
    
    // 计算实际时间戳（毫秒）
    timestampMs := (g.startTime*10 + int64(elapsedTime)*10)
    
    // 时间戳压缩（毫秒级）
    compressedTime := ToTimestampShortMs(timestampMs)
    
    // Base62编码各部分
    businessStr := EncodeBase62(uint64(businessType))
    machineStr := EncodeBase62(uint64(machineID))
    sequenceStr := EncodeBase62(uint64(sequence))
    
    // 拼接
    return businessStr + compressedTime + machineStr + sequenceStr, nil
}
```

### Step 5: 错误定义

**文件**：`generator.go`（追加）

```go
var (
    ErrOverTimeLimit = errors.New("over the time limit")
)
```

### Step 6: 添加依赖

**文件**：`go.mod`

```bash
go get github.com/redis/go-redis/v9
```

---

## 使用示例

### 示例1：固定机器ID（传统部署）

```go
generator, _ := shortid.NewGenerator(shortid.Config{
    MachineID:    1,
    BusinessType: shortid.BusinessOrder,
})

id, _ := generator.Generate()
fmt.Println(id) // 输出: "300000011V"
```

### 示例2：Serverless模式（Redis）

```go
provider, _ := shortid.NewRedisProvider("localhost:6379")
defer provider.Close()

generator, _ := shortid.NewGenerator(shortid.Config{
    Provider:     provider,
    BusinessType: shortid.BusinessOrder,
})

id, _ := generator.GenerateWithContext(ctx)
fmt.Println(id)
```

### 示例3：测试模式（内存）

```go
provider := shortid.NewMemoryProvider()
defer provider.Close()

generator, _ := shortid.NewGenerator(shortid.Config{
    Provider:     provider,
    BusinessType: shortid.BusinessOrder,
})

id, _ := generator.GenerateWithContext(ctx)
fmt.Println(id)
```

---

## 实现检查清单

- [x] Base62编码 ✅
- [x] 毫秒级时间戳压缩 ✅
- [x] 业务类型定义 ✅
- [x] Snowflake常量定义 ✅
- [ ] StateProvider接口定义
- [ ] MemoryProvider实现
- [ ] RedisProvider实现
- [ ] Generator核心实现
- [ ] 单元测试
- [ ] 集成测试

---

**最后更新**：2024-12-16
