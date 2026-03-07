# GoFrame 集成说明（SDK 形态）

本仓库保持为纯 SDK，不强制依赖 GoFrame。  
如需在 GoFrame 项目中直接挂载路由，可使用可选适配层：`adapter/gfhttp`。

## 设计原则

- `shortid` 核心包保持框架无关，便于在任意 HTTP/gRPC 框架中复用。
- GoFrame 集成放在独立目录，并使用 build tag `goframe` 隔离可选依赖。

## 使用方式

1. 在你的业务项目中启用 `goframe` build tag（或复制适配层代码）。
2. 引入 `adapter/gfhttp`，将路由绑定到 `ghttp.Server`。

参考 API：

- `gfhttp.BindRoutes(server *ghttp.Server, generator *shortid.Generator)`
- `gfhttp.BindRoutesWithEndpoint(server *ghttp.Server, endpoint *shortid.Endpoint)`

也可直接使用传输无关端口（适用于 Gin/Fiber/gRPC/MQ 消费者等）：

- `endpoint := shortid.NewEndpoint(generator)`
- `id, err := endpoint.NextID(ctx)`
- `err := endpoint.Health(ctx)`

默认注册：

- `GET/POST /nextid`
- `GET /health`
