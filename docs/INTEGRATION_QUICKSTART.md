# shortid 快速对接手册

本文档给新接入方一条最短可落地路径，默认目标是“先跑通，再生产化”。

## 1. 安装

```bash
go get github.com/gostool/shortid
```

## 2. 5分钟接入（推荐起步）

```go
package main

import (
	"context"
	"fmt"

	"github.com/gostool/shortid"
)

func main() {
	g, err := shortid.New(1, shortid.BusinessOrder)
	if err != nil {
		panic(err)
	}

	rawID, err := g.NextID(context.Background())
	if err != nil {
		panic(err)
	}
	shortID, err := g.GenerateWithContext(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("raw:", rawID)
	fmt.Println("short:", shortID)
}
```

高级模式（自定义 Provider / 基准时间）再切换到 `Config + NewGenerator`。

## 3. 推荐集成方式：通过 Endpoint 对接业务层

`Endpoint` 是传输无关入口，方便统一接 HTTP/gRPC/MQ/任务系统。

```go
ep := shortid.NewEndpoint(g)

id, err := ep.NextID(ctx)
if err != nil {
	return err
}
if err := ep.Health(ctx); err != nil {
	return err
}
_ = id
```

## 4. 部署模式选择

| 场景 | 配置方式 | 说明 |
|---|---|---|
| 单机/固定节点 | `MachineID` | 依赖最少，适合起步 |
| Serverless/弹性实例 | `MachineIDLeaseProvider` | 推荐：租约模式（分布式锁级安全） |
| 强一致分布式序列 | `MachineIDLeaseProvider + SequenceProvider` | 多实例并发下统一序列来源 |

注意：
- `MachineID` 与 `MachineIDProvider` 互斥。
- `MachineID`、`MachineIDProvider`、`MachineIDLeaseProvider` 两两互斥。
- `MachineID` 范围 `0~63`。

## 5. HTTP 示例

仓库内提供示例启动文件：

```bash
go run ./example_http
```

默认接口：
- `GET/POST /nextid`
- `GET /health`

可用环境变量：
- `REDIS_ADDR`，默认 `localhost:6379`

## 6. GoFrame 可选适配

如业务项目使用 GoFrame，可使用 `adapter/gfhttp`（build tag: `goframe`）：

```go
//go:build goframe

gfhttp.BindRoutes(server, g)
// 或
gfhttp.BindRoutesWithEndpoint(server, ep)
```

## 7. 生产前验证清单

```bash
go test ./...
go test -race ./...
SHORTID_REQUIRE_REDIS_TESTS=1 go test ./...
```

或使用 Makefile：

```bash
make test
make test-redis
```

## 8. 常见问题

- `invalid machine id`
  `MachineID` 超范围（必须 `0~63`）。
- `invalid business type`
  `BusinessType` 非法。
- `invalid generator config`
  同时设置了 `MachineID` 与 `MachineIDProvider`。

## 9. 相关文档

- `README.md`
- `docs/API_CONTRACT.md`
- `docs/GOFRAME_INTEGRATION.md`
- `docs/MINIMAL_VALIDATION.md`
- `docs/UPGRADING.md`
