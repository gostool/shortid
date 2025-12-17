# 文件结构规划

## 📁 文件组织

### 核心文件结构

```
shortid/
├── machineid/                     # 机器ID提供者模块
│   ├── provider.go                # MachineIDProvider 接口定义
│   ├── memory.go                  # MemoryMachineIDProvider 实现（测试用）
│   ├── redis.go                   # RedisMachineIDProvider 实现（生产用）
│   └── provider_test.go           # 机器ID提供者测试
│
├── sequence/                      # 序列号提供者模块
│   ├── provider.go                # SequenceProvider 接口定义
│   ├── memory.go                  # MemorySequenceProvider 实现（测试用）
│   ├── redis.go                   # RedisSequenceProvider 实现（生产用）
│   └── provider_test.go           # 序列号提供者测试
│
├── generator.go                   # Generator 核心实现
├── generator_test.go              # Generator 测试
│
├── errors.go                      # 错误定义（新增）
│
├── base.go                        # ✅ 已有：Base62编码
├── timestamp.go                   # ✅ 已有：时间戳压缩
├── const.go                       # ✅ 已有：常量定义
├── logic_enum.go                 # ✅ 已有：业务类型定义
│
└── docs/                          # 文档目录
    └── step3:go唯一id.md          # 实现文档
```

## 📝 文件说明

### 1. `machineid/provider.go`
**职责**：定义 `MachineIDProvider` 接口

**内容**：
- `MachineIDProvider` 接口定义
- 接口方法：`GetMachineID`, `SetMachineIDExpiration`, `HealthCheck`, `Close`

### 2. `machineid/memory.go`
**职责**：内存实现的机器ID提供者（测试用）

**内容**：
- `MemoryMachineIDProvider` 结构体
- `NewMemoryMachineIDProvider` 构造函数
- 实现 `MachineIDProvider` 接口的所有方法

### 3. `machineid/redis.go`
**职责**：Redis实现的机器ID提供者（生产用）

**内容**：
- `RedisMachineIDProvider` 结构体
- `NewRedisMachineIDProvider` 构造函数
- `NewRedisMachineIDProviderWithOptions` 构造函数（支持自定义选项）
- 实现 `MachineIDProvider` 接口的所有方法

### 4. `sequence/provider.go`
**职责**：定义 `SequenceProvider` 接口

**内容**：
- `SequenceProvider` 接口定义
- 接口方法：`GetSequence`, `SetSequenceExpiration`, `HealthCheck`, `Close`

### 5. `sequence/memory.go`
**职责**：内存实现的序列号提供者（测试用）

**内容**：
- `MemorySequenceProvider` 结构体
- `NewMemorySequenceProvider` 构造函数
- 使用 `map[string]int64` 存储每个时间单位的序列号
- 实现 `SequenceProvider` 接口的所有方法

### 6. `sequence/redis.go`
**职责**：Redis实现的序列号提供者（生产用）

**内容**：
- `RedisSequenceProvider` 结构体
- `NewRedisSequenceProvider` 构造函数
- `NewRedisSequenceProviderWithOptions` 构造函数（支持自定义选项）
- 实现 `SequenceProvider` 接口的所有方法

### 7. `generator.go`
**职责**：Generator 核心实现

**内容**：
- `Generator` 结构体
- `Config` 配置结构
- `NewGenerator` 构造函数
- `Generate` / `GenerateWithContext` 方法
- `nextID` 方法（Sonyflake算法实现）
- `toShortID` 方法（短ID转换）

### 8. `errors.go`
**职责**：错误定义

**内容**：
- `ErrOverTimeLimit` 时间戳溢出错误
- 其他相关错误定义

### 9. 测试文件

- `machineid/provider_test.go` - 机器ID提供者测试
- `sequence/provider_test.go` - 序列号提供者测试
- `generator_test.go` - Generator 测试

## 🔄 迁移计划

### 阶段1：创建新文件结构
1. 创建 `machineid/` 目录和文件
2. 创建 `sequence/` 目录和文件
3. 创建 `generator.go`
4. 创建 `errors.go`

### 阶段2：实现接口
1. 实现 `MachineIDProvider` 接口（memory + redis）
2. 实现 `SequenceProvider` 接口（memory + redis）
3. 实现 `Generator` 核心逻辑

### 阶段3：测试
1. 编写单元测试
2. 编写集成测试

### 阶段4：清理
1. 删除旧的 `provider.go` 和 `provider_redis.go`
2. 更新文档

## 📦 包结构

所有文件都在 `package shortid` 下，保持统一的包名。

## 🎯 命名规范

- **接口**：`MachineIDProvider`, `SequenceProvider`
- **内存实现**：`MemoryMachineIDProvider`, `MemorySequenceProvider`
- **Redis实现**：`RedisMachineIDProvider`, `RedisSequenceProvider`
- **构造函数**：`NewMemoryMachineIDProvider`, `NewRedisMachineIDProvider` 等

