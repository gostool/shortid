# Upgrading Guide

## 从 pre-v1 到 v1

1. 启动阶段先调用 `ValidateConfig(config)`，提前暴露配置问题。
2. 若同时设置 `MachineID` 和 `MachineIDProvider`，改为二选一。
3. 如使用 `HTTPServer` 辅助 API，建议迁移到业务服务层自行封装 transport。

## 行为确认清单

- 核心 ID 生成路径是否仍走 `Generator`。
- Redis 相关 provider 是否可由 `context` 控制超时。
- CI 是否启用 Redis 强制模式与 race 检测。
