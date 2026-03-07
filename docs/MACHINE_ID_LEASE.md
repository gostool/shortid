# 机器ID租约模式设计说明

## 1. 目标

机器ID分配改为“租约模型”，保证在分布式环境中：

- 同一时刻一个机器ID只能被一个实例持有；
- 续租和释放必须校验持有者身份（token CAS）；
- 租约失效后，实例必须停止继续使用该机器ID。

这与分布式锁的核心安全属性一致：**互斥 + 所有权校验 + 自动过期**。

## 2. 核心接口

`MachineIDLeaseProvider`：

- `AcquireMachineIDLease(ctx, ttl)`
- `RenewMachineIDLease(ctx, lease, ttl) (ok bool, err error)`
- `ReleaseMachineIDLease(ctx, lease)`
- `HealthCheck(ctx)`
- `Close()`

租约对象 `MachineIDLease`：

- `MachineID`：租约内分配的机器ID
- `Token`：租约令牌（持有者身份）
- `ExpiresAt`：本地视角到期时间

## 3. 安全语义（必须遵守）

适配方（Redis / etcd / 第三方）必须保证：

1. **互斥获取**
- 获取租约必须是原子互斥（如 Redis `SET NX PX`、etcd lease + txn）。

2. **续租必须 CAS**
- 只有 `key` 当前值等于 `Token` 才能续租成功；
- 若 token 不匹配或租约已过期，返回 `ok=false`。

3. **释放必须 CAS**
- 只有 `key` 当前值等于 `Token` 才能释放；
- 禁止无 token 直接删除（避免误删他人租约）。

4. **失租即失效**
- `RenewMachineIDLease` 返回 `ok=false` 时，调用方必须停止使用该机器ID。

## 4. Generator 行为

`Generator` 在租约模式下：

- 首次生成ID前先申请租约；
- 在租约周期中点自动续租；
- 续租失败（`ok=false`）返回 `ErrMachineIDLeaseLost`，拒绝继续发号。

默认租约时长：`20m`（可通过 `Config.MachineIDLeaseDuration` 配置）。

默认槽位数：`64`（与当前机器位宽一致），支持在 Redis 适配层按需下调。

## 5. Redis 参考实现

本仓库 Redis 适配实现要点：

- 获取：`SET leaseKey token NX PX ttl`
- 续租：Lua CAS（`GET == token` 后 `PEXPIRE`）
- 释放：Lua CAS（`GET == token` 后 `DEL`）

这保证了与分布式锁同级的所有权安全。

### 槽位配置

可通过 `RedisMachineIDLeaseOptions` 配置：

- `Slots`：可分配槽位数，范围 `[1,64]`，默认 `64`
- `CursorKey`：游标键
- `LeaseKeyPrefix`：租约键前缀

示例：

```go
provider, err := shortid.NewRedisMachineIDLeaseProviderWithConfig(
    "localhost:6379",
    shortid.RedisMachineIDLeaseOptions{
        Slots:          32,
        CursorKey:      "myapp:machine:lease:cursor",
        LeaseKeyPrefix: "myapp:machine:lease:",
    },
)
```

## 6. etcd / 第三方适配建议

建议等价实现：

- 使用“租约 + 事务条件更新”完成 CAS；
- Token 必须是不可预测随机值；
- 续租失败要明确返回 `ok=false`，不要静默重试掩盖失租。

## 7. 风险与边界

- 任意分布式租约都依赖时钟与网络，需设置合理 TTL 与超时；
- 业务侧应将 `ErrMachineIDLeaseLost` 视为高优先级故障并触发重建实例/告警；
- 租约模式解决“机器ID所有权”，不替代业务唯一约束与幂等策略。
