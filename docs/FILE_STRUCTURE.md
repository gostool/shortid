# 文件结构规划

## 📁 文件组织

### 核心文件结构

```
shortid/
├── provider.go                    # ✅ 接口定义（MachineIDProvider, SequenceProvider）
│
├── machineid/                     # 机器ID提供者模块
│   ├── memory.go                  # ✅ MemoryMachineIDProvider 实现（测试用）
│   ├── redis.go                   # ✅ RedisMachineIDProvider 实现（生产用）
│   ├── memory_test.go             # ✅ 机器ID提供者测试
│   └── redis_test.go              # ✅ Redis机器ID提供者测试
│
├── sequence/                      # 序列号提供者模块
│   ├── memory.go                  # ✅ MemorySequenceProvider 实现（测试用）
│   ├── redis.go                   # ✅ RedisSequenceProvider 实现（生产用）
│   ├── memory_test.go             # ✅ 序列号提供者测试
│   └── redis_test.go              # ✅ Redis序列号提供者测试
│
├── generator.go                   # ✅ Generator 核心实现（ID时间/序列主流程）
├── generator_machine_runtime.go   # ✅ 机器ID运行时（固定/Provider/租约）
├── generator_test.go              # ✅ Generator 单元测试（配置/生命周期/并发）
│
├── http_server.go                 # ✅ HTTP服务器实现（提供ID生成API）
│
├── errors.go                      # ✅ 错误定义
├── base.go                        # ✅ Base62编码
├── timestamp.go                   # ✅ 时间戳压缩
├── const.go                       # ✅ 常量定义
├── logic_enum.go                 # ✅ 业务类型定义
│
├── base_test.go                   # ✅ Base62编码测试
├── timestamp_test.go              # ✅ 时间戳压缩测试
│
├── sdk_*.go                       # ✅ SDK集成测试文件
│   ├── sdk_single_mem_test.go     # 单机内存模式测试
│   ├── sdk_single_redis_test.go   # 单机Redis模式测试
│   ├── sdk_serverless_redis_test.go  # Serverless Redis模式测试
│   └── sdk_http_redis_test.go     # HTTP服务测试
│
├── example_http/                  # ✅ HTTP服务示例
│   └── main.go                    # HTTP服务器启动示例
│
├── docs/                          # 文档目录
│   ├── FILE_STRUCTURE.md          # 文件结构文档（本文件）
│   ├── MINIMAL_VALIDATION.md      # 最小验证手册
│   ├── PERFORMANCE_TEST.md        # 性能测试报告
│   ├── step3:go唯一id.md          # 实现文档
│   └── ...                        # 其他文档
│
├── test.sh                        # 测试脚本
├── test_http.sh                   # HTTP测试脚本
├── benchmark_http.sh              # HTTP性能测试脚本
├── Makefile                       # 构建脚本
├── go.mod                         # Go模块定义
└── LICENSE                        # 许可证
```

## 📝 文件说明

### 1. `provider.go`
**职责**：定义 `MachineIDProvider` 和 `SequenceProvider` 接口

**位置**：根目录

**内容**：
- `MachineIDProvider` 接口定义
  - `GetMachineID(ctx context.Context) (uint16, error)`
  - `SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error`
  - `HealthCheck(ctx context.Context) error`
  - `Close() error`
- `SequenceProvider` 接口定义
  - `GetSequence(ctx context.Context, key string) (uint16, error)`
  - `SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error`
  - `HealthCheck(ctx context.Context) error`
  - `Close() error`

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

### 4. `sequence/memory.go`
**职责**：内存实现的序列号提供者（测试用）

**内容**：
- `MemorySequenceProvider` 结构体
- `NewMemorySequenceProvider` 构造函数
- 使用 `map[string]int64` 存储每个时间单位的序列号
- 实现 `SequenceProvider` 接口的所有方法

### 5. `sequence/redis.go`
**职责**：Redis实现的序列号提供者（生产用）

**内容**：
- `RedisSequenceProvider` 结构体
- `NewRedisSequenceProvider` 构造函数
- `NewRedisSequenceProviderWithOptions` 构造函数（支持自定义选项）
- 实现 `SequenceProvider` 接口的所有方法

### 6. `generator.go`
**职责**：Generator 核心实现（时间推进、序列处理、ID组装）

**内容**：
- `Generator` 结构体
- `Config` 配置结构
- `NewGenerator` 构造函数
- `Generate` / `GenerateWithContext` 方法
- `nextID` 方法（Sonyflake算法实现）
- `toShortID` 方法（短ID转换）

### 7. `generator_machine_runtime.go`
**职责**：机器ID运行时管理（从 `nextID` 中解耦）

**内容**：
- `ensureMachineIdentity` 统一入口
- `ensureMachineLease` 租约获取与续租
- `ensureMachineProvider` 兼容旧 `MachineIDProvider` 延迟初始化

### 8. `http_server.go`
**职责**：HTTP服务器实现，提供ID生成API

**内容**：
- `HTTPServer` 结构体
- `NewHTTPServer` 构造函数
- `Start`, `StartTLS`, `Shutdown` 方法
- `handleNextID` - 处理 `/nextid` 端点
- `handleHealth` - 处理 `/health` 端点（返回统计信息）
- `ServerStats` - 服务器统计信息
- `HealthResponse` - 健康检查响应结构
- `IDResponse` - ID生成响应结构
- Redis提供者实现（`redisMachineIDProviderImpl`, `redisSequenceProviderImpl`）

### 9. `errors.go`
**职责**：错误定义

**内容**：
- `ErrOverTimeLimit` - 时间戳溢出错误
- `ErrInvalidBusinessType` - 无效业务类型错误
- `ErrInvalidMachineID` - 无效机器ID错误
- `ErrInvalidSequence` - 无效序列号错误

### 10. 测试文件

#### 单元测试
- `generator_test.go` - Generator 核心功能测试
- `base_test.go` - Base62编码测试
- `timestamp_test.go` - 时间戳压缩测试
- `machineid/memory_test.go` - 内存机器ID提供者测试
- `machineid/redis_test.go` - Redis机器ID提供者测试
- `sequence/memory_test.go` - 内存序列号提供者测试
- `sequence/redis_test.go` - Redis序列号提供者测试

#### SDK集成测试
- `sdk_single_mem_test.go` - 单机内存模式SDK测试
  - `TestSDK_SingleMemory_ShortID` - 测试短ID生成
  - `TestSDK_SingleMemory_UID` - 测试原始数字ID生成
  - `TestSDK_SingleMemory_NextID` - 测试NextID方法
- `sdk_single_redis_test.go` - 单机Redis模式SDK测试
- `sdk_serverless_redis_test.go` - Serverless Redis模式SDK测试
  - `TestSDK_ServerlessRedis_ShortID` - 完整Serverless模式（Redis机器ID+Redis序列号）
  - `TestSDK_ServerlessRedis_NextID` - 完整Serverless模式生成原始ID
  - `TestSDK_ServerlessRedis_Simplified` - 简化模式（Redis机器ID+本地序列号）
  - `TestSDK_ServerlessRedis_Concurrent` - 并发测试
- `sdk_http_redis_test.go` - HTTP服务测试
  - `TestSDK_HTTPRedis_NextID` - HTTP API生成ID测试
  - `TestSDK_HTTPRedis_Health` - 健康检查端点测试
  - `TestSDK_HTTPRedis_Concurrent` - 并发HTTP请求测试
  - `TestSDK_HTTPRedis_MethodNotAllowed` - 方法限制测试

### 11. 示例和脚本

- `example_http/main.go` - HTTP服务器启动示例
- `test.sh` - 单元测试脚本
- `test_http.sh` - HTTP服务测试脚本
- `benchmark_http.sh` - HTTP性能基准测试脚本
- `Makefile` - 构建和测试命令

## ✅ 实现状态

### 已完成的功能

#### 核心功能
- ✅ `MachineIDProvider` 接口定义（`provider.go`）
- ✅ `SequenceProvider` 接口定义（`provider.go`）
- ✅ `MemoryMachineIDProvider` 实现（`machineid/memory.go`）
- ✅ `RedisMachineIDProvider` 实现（`machineid/redis.go`）
- ✅ `MemorySequenceProvider` 实现（`sequence/memory.go`）
- ✅ `RedisSequenceProvider` 实现（`sequence/redis.go`）
- ✅ `Generator` 核心实现（`generator.go` + `generator_machine_runtime.go`）
  - ✅ 支持固定机器ID模式
  - ✅ 支持Serverless模式（动态机器ID）
  - ✅ 支持本地序列号模式
  - ✅ 支持分布式序列号模式
  - ✅ 支持短ID生成（Base62编码）
  - ✅ 支持原始数字ID生成（uint64）

#### HTTP服务
- ✅ HTTP服务器实现（`http_server.go`）
  - ✅ `/nextid` 端点（生成ID）
  - ✅ `/health` 端点（健康检查和统计信息）
  - ✅ 服务器启动时自动获取机器ID
  - ✅ 统计信息收集（QPS、响应时间、成功率等）

#### 测试
- ✅ Generator单元测试
- ✅ Base62编码测试
- ✅ 时间戳压缩测试
- ✅ 内存Provider测试
- ✅ Redis Provider测试
- ✅ SDK集成测试（单机内存、Serverless Redis、HTTP服务）

#### 文档
- ✅ 文件结构文档（本文件）
- ✅ 最小验证手册（`MINIMAL_VALIDATION.md`）
- ✅ 性能测试报告（`PERFORMANCE_TEST.md`）
- ✅ 实现文档（`step3:go唯一id.md`）

### 待实现的功能

- ⏳ HTTP服务的更多测试场景

## 📦 包结构

所有文件都在 `package shortid` 下，保持统一的包名。

**注意**：SDK测试文件（`sdk_*.go`）也使用 `package shortid`，可以直接访问包内导出的类型和函数。

## 🎯 命名规范

- **接口**：`MachineIDProvider`, `SequenceProvider`
- **内存实现**：`MemoryMachineIDProvider`, `MemorySequenceProvider`
- **Redis实现**：`RedisMachineIDProvider`, `RedisSequenceProvider`
- **构造函数**：`NewMemoryMachineIDProvider`, `NewRedisMachineIDProvider` 等
- **HTTP服务**：`HTTPServer`, `NewHTTPServer`
- **响应结构**：`IDResponse`, `HealthResponse`

## 🔍 关键设计决策

### 1. 接口位置
- 接口定义在根目录的 `provider.go` 中，而不是子目录
- 原因：方便 `generator.go` 直接使用，避免循环依赖

### 2. 测试文件组织
- 单元测试：与源文件同目录或根目录
- SDK集成测试：根目录的 `sdk_*.go` 文件
- 原因：Go测试约定，便于运行和管理

### 3. HTTP服务器实现
- `http_server.go` 包含Redis提供者的内部实现
- 原因：避免在测试文件中重复实现，统一管理

### 4. 统计信息
- HTTP服务器内置统计信息收集
- `/health` 端点返回详细的性能指标
- 原因：便于监控和调试

## 📊 性能指标

根据 `PERFORMANCE_TEST.md`：
- **QPS**: 11,479（100并发）
- **平均响应时间**: 8.711ms（100并发）
- **99%响应时间**: 15ms
- **成功率**: 100%

## 🚀 使用示例

### 1. 单机内存模式
```go
generator, _ := shortid.NewGenerator(shortid.Config{
    MachineID:    1,
    BusinessType: shortid.BusinessOrder,
})
id, _ := generator.Generate()
```

### 2. Serverless模式（Redis）
```go
machineProvider := shortid.NewRedisMachineIDProvider("localhost:6379")
generator, _ := shortid.NewGenerator(shortid.Config{
    MachineIDProvider: machineProvider,
    BusinessType:      shortid.BusinessOrder,
})
id, _ := generator.GenerateWithContext(ctx)
```

### 3. HTTP服务
```go
server, _ := shortid.NewHTTPServer(":8080", "localhost:6379", shortid.BusinessOrder)
server.Start()
```
