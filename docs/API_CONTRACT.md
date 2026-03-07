# API Contract (v1)

本文档定义 `shortid` 对外稳定契约。

## 1. 核心公开接口

- `type Config`
- `func ValidateConfig(Config) error`
- `func NewGenerator(Config) (*Generator, error)`
- `(*Generator).Generate() (string, error)`
- `(*Generator).GenerateWithContext(context.Context) (string, error)`
- `(*Generator).NextID(context.Context) (uint64, error)`
- `type MachineIDProvider interface`
- `type SequenceProvider interface`
- 公开编码与时间戳函数（`base.go`/`timestamp.go`）

## 2. 行为契约

- `Generator` 并发安全；同一实例内生成 ID 应保持唯一。
- `ValidateConfig` 仅做参数与组合合法性校验，不做依赖连通性探测。
- `MachineID` 与 `MachineIDProvider` 互斥。
- `MachineID` 取值范围 `0~63`，`Sequence` 取值范围 `0~127`。
- provider 由调用方通过 `context` 控制超时与取消；实现方不应无限重试。

## 3. 错误语义

- 参数错误：`ErrInvalidMachineID`、`ErrInvalidBusinessType`、`ErrInvalidConfig`。
- 资源上限：`ErrOverTimeLimit`。
- 依赖错误：provider 实现返回的错误会向上透传并包装上下文。

## 4. 废弃策略

- 使用 Go 标准 `Deprecated:` 注释。
- 废弃后至少保留一个 minor 版本。
- 文档必须同时给出替代路径。
