# shortid

`shortid` 是一个工程稳定优先的 Go SDK，用于生成分布式唯一 ID 和短编码字符串。

## 定位

- 核心能力：`Generator`、编码解码、`MachineIDProvider`/`SequenceProvider` 抽象。
- 非目标：HTTP/gRPC 传输层不是核心演进目标；建议在业务服务层封装。

## 快速开始

```go
package main

import (
    "context"
    "fmt"

    "github.com/gostool/shortid"
)

func main() {
    cfg := shortid.Config{
        MachineID:    1,
        BusinessType: shortid.BusinessOrder,
    }
    if err := shortid.ValidateConfig(cfg); err != nil {
        panic(err)
    }

    g, err := shortid.NewGenerator(cfg)
    if err != nil {
        panic(err)
    }

    id, err := g.GenerateWithContext(context.Background())
    if err != nil {
        panic(err)
    }
    fmt.Println(id)
}
```

## 部署模式选择

| 场景 | MachineID | Sequence | 说明 |
|---|---|---|---|
| 单机/固定节点 | 固定 `MachineID` | 本地序列 | 最简单、最低依赖 |
| Serverless/弹性实例 | `MachineIDProvider` | 本地序列 | 机器ID集中分配，性能优先 |
| 全分布式序列 | `MachineIDProvider` | `SequenceProvider` | 跨实例强一致序列 |

## 稳定性承诺（v1）

- `v1` 内不做破坏性 API 变更。
- 新能力优先通过新增可选字段/函数扩展。
- 废弃策略：`Deprecated` 标注后至少保留一个 minor 版本窗口。

## 测试与质量门禁

- 单元与集成：`go test ./...`
- Redis 强制模式：`SHORTID_REQUIRE_REDIS_TESTS=1 go test ./...`
- 竞态检测：`go test -race ./...`
- 基准回归门禁：`scripts/bench_guard.sh`

## 常见排查

- `invalid machine id`：`MachineID` 必须在 `0~63`。
- `invalid business type`：`BusinessType` 超出约束范围。
- Redis 测试被跳过：未开启 Redis 或未设置 `SHORTID_REQUIRE_REDIS_TESTS=1`。

更多见 [docs/API_CONTRACT.md](docs/API_CONTRACT.md) 与 [docs/RELEASE_POLICY.md](docs/RELEASE_POLICY.md)。
