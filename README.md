# ShortID - 分布式唯一ID生成器

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

一个高性能的分布式唯一ID生成器，基于 Sonyflake 算法实现，支持 Serverless 部署和短ID生成。

## ✨ 特性

- 🚀 **高性能**: QPS 达到 11,479（100并发），平均响应时间 8.7ms
- 🔒 **分布式**: 支持 Serverless 模式，动态分配机器ID
- 📦 **短ID**: Base62 编码，生成 8-12 字符的短ID
- 🌐 **HTTP API**: 提供 RESTful API，支持健康检查和统计信息
- 🔧 **灵活配置**: 支持固定机器ID、Redis分布式、本地序列号等多种模式
- 📊 **监控**: 内置性能统计和健康检查

## 📦 安装

### 私有仓库配置

由于这是私有仓库，需要先配置环境：

```bash
# 设置 GOPRIVATE
export GOPRIVATE=github.com/gostool

# 配置 Git SSH（推荐）
git config --global url."git@github.com:".insteadOf "https://github.com/"

# 永久设置（添加到 ~/.zshrc）
echo 'export GOPRIVATE=github.com/gostool' >> ~/.zshrc
source ~/.zshrc
```

### 获取包

```bash
go get github.com/gostool/shortid@v1.0.0

# 或获取最新版本
go get -u github.com/gostool/shortid@latest
```

详细配置说明请参考 [私有包配置指南](docs/PRIVATE_SETUP.md)。

## 🚀 快速开始

### 1. 单机模式（固定机器ID）

```go
package main

import (
    "fmt"
    "github.com/gostool/shortid"
)

func main() {
    // 创建生成器
    generator, err := shortid.NewGenerator(shortid.Config{
        MachineID:    1,                    // 固定机器ID
        BusinessType: shortid.BusinessOrder, // 业务类型
    })
    if err != nil {
        panic(err)
    }

    // 生成短ID
    id, err := generator.Generate()
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Short ID:", id) // 输出: Short ID: Pby0X2Cy19
}
```

### 2. Serverless 模式（Redis）

```go
package main

import (
    "context"
    "fmt"
    "github.com/gostool/shortid"
)

func main() {
    // 创建 Redis 机器ID提供者
    machineProvider, err := shortid.NewRedisMachineIDProvider("localhost:6379")
    if err != nil {
        panic(err)
    }
    defer machineProvider.Close()

    // 创建生成器
    generator, err := shortid.NewGenerator(shortid.Config{
        MachineIDProvider: machineProvider,
        BusinessType:      shortid.BusinessOrder,
    })
    if err != nil {
        panic(err)
    }

    // 生成ID
    ctx := context.Background()
    id, err := generator.GenerateWithContext(ctx)
    if err != nil {
        panic(err)
    }
    
    fmt.Println("ID:", id)
}
```

### 3. HTTP 服务

```go
package main

import (
    "log"
    "github.com/gostool/shortid"
)

func main() {
    // 创建 HTTP 服务器
    server, err := shortid.NewHTTPServer(
        ":8080",                    // 监听地址
        "localhost:6379",           // Redis 地址
        shortid.BusinessOrder,      // 业务类型
    )
    if err != nil {
        log.Fatal(err)
    }

    // 启动服务器
    log.Println("Server starting on :8080")
    if err := server.Start(); err != nil {
        log.Fatal(err)
    }
}
```

启动后访问：
- `GET /nextid` - 生成ID
- `GET /health` - 健康检查和统计信息

## 📖 API 文档

### Generator

#### NewGenerator

创建ID生成器。

```go
func NewGenerator(config Config) (*Generator, error)
```

**配置选项**：
- `MachineID`: 固定机器ID（0-63）
- `MachineIDProvider`: 机器ID提供者（Serverless模式）
- `SequenceProvider`: 序列号提供者（分布式序列号）
- `BusinessType`: 业务类型
- `ReturnRawID`: 是否返回原始数字ID（默认false，返回短ID）

#### Generate / GenerateWithContext

生成ID。

```go
func (g *Generator) Generate() (string, error)
func (g *Generator) GenerateWithContext(ctx context.Context) (string, error)
```

#### NextID

生成原始数字ID（uint64）。

```go
func (g *Generator) NextID(ctx context.Context) (uint64, error)
```

### HTTP Server

#### NewHTTPServer

创建HTTP服务器。

```go
func NewHTTPServer(addr, redisAddr string, businessType BusinessType) (*HTTPServer, error)
```

#### 端点

- `GET /nextid` - 生成ID，返回 JSON: `{"id": 1234567890}`
- `GET /health` - 健康检查，返回统计信息

## 🎯 使用场景

### 场景1: 分布式唯一ID（Serverless部署）

适合 Serverless、容器化等动态部署场景：

```go
// 使用 Redis 管理机器ID和序列号
machineProvider := shortid.NewRedisMachineIDProvider("redis:6379")
sequenceProvider := shortid.NewRedisSequenceProvider("redis:6379")

generator, _ := shortid.NewGenerator(shortid.Config{
    MachineIDProvider: machineProvider,
    SequenceProvider:  sequenceProvider,
    BusinessType:      shortid.BusinessOrder,
})
```

### 场景2: 生成短ID

适合需要短ID的场景（如短链接、邀请码等）：

```go
generator, _ := shortid.NewGenerator(shortid.Config{
    MachineID:    1,
    BusinessType: shortid.BusinessOrder,
})

id, _ := generator.Generate() // 返回: "Pby0X2Cy19" (10字符)
```

## 📊 性能指标

根据性能测试报告（[PERFORMANCE_TEST.md](docs/PERFORMANCE_TEST.md)）：

| 指标 | 值 |
|------|-----|
| **QPS** | 11,479（100并发） |
| **平均响应时间** | 8.7ms |
| **99%响应时间** | 15ms |
| **成功率** | 100% |

测试环境：
- 机器ID: Redis（分布式）
- 序列号: 本地内存
- 并发: 100
- 请求数: 10,000

## 📁 项目结构

```
shortid/
├── provider.go              # 接口定义
├── generator.go              # 核心实现
├── http_server.go           # HTTP服务器
├── machineid/               # 机器ID提供者
│   ├── memory.go           # 内存实现
│   └── redis.go            # Redis实现
├── sequence/                # 序列号提供者
│   ├── memory.go           # 内存实现
│   └── redis.go            # Redis实现
└── docs/                    # 文档
    ├── FILE_STRUCTURE.md    # 文件结构
    ├── PERFORMANCE_TEST.md  # 性能测试
    └── PRIVATE_SETUP.md     # 私有包配置
```

详细结构请参考 [文件结构文档](docs/FILE_STRUCTURE.md)。

## 🔧 配置说明

### 业务类型

```go
const (
    BusinessOrder    BusinessType = 3  // 订单
    BusinessPayment  BusinessType = 4  // 支付
    BusinessUser     BusinessType = 5  // 用户
    // ... 更多业务类型
)
```

### 机器ID模式

- **固定模式**: 使用 `MachineID` 字段，适合传统部署
- **Serverless模式**: 使用 `MachineIDProvider`，适合动态部署

### 序列号模式

- **本地模式**: 不设置 `SequenceProvider`，序列号在本地内存中维护
- **分布式模式**: 设置 `SequenceProvider`，序列号在 Redis 中管理

## 📚 文档

- [文件结构](docs/FILE_STRUCTURE.md) - 项目文件结构说明
- [性能测试报告](docs/PERFORMANCE_TEST.md) - 详细的性能测试结果
- [私有包配置指南](docs/PRIVATE_SETUP.md) - 私有仓库访问配置
- [版本测试指南](docs/VERSION_TEST.md) - 版本验证方法

## 🧪 测试

```bash
# 运行所有测试
go test ./...

# 运行特定测试
go test -v -run TestSDK_SingleMemory

# 运行 HTTP 服务测试（需要 Redis）
go test -v -run TestSDK_HTTPRedis
```

## 📝 版本历史

- **v1.0.0** (2024-12-17)
  - 完整的分布式唯一ID生成器实现
  - HTTP服务器和健康检查
  - 完善的文档和测试

- **v0.4.0** (2024-12-17)
  - 添加HTTP服务器实现
  - 添加健康检查端点
  - 修复文档文件名问题

查看所有版本: `git tag -l`

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

[MIT License](LICENSE)

## 🔗 相关链接

- [Go Modules 文档](https://go.dev/ref/mod)
- [Sonyflake 算法](https://github.com/sony/sonyflake)

---

**注意**: 这是私有仓库，使用前请配置 GOPRIVATE 环境变量。详见 [私有包配置指南](docs/PRIVATE_SETUP.md)。

