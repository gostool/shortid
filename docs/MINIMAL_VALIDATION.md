# 最小可运行验证手册

## 1. 环境准备

- Go 1.22+
- Redis（本地默认地址 `localhost:6379`）

## 2. 快速验证（无Redis强依赖）

```bash
make test
```

说明：
- 会执行 `go test`。
- Redis相关用例在Redis不可用时会 `Skip`。

## 3. Redis必测模式（建议CI使用）

```bash
make test-redis
```

说明：
- 等价于 `SHORTID_REQUIRE_REDIS_TESTS=1 go test ./...`。
- Redis不可用会直接失败，确保分布式链路被真实验证。

## 4. 竞态与性能门禁（建议CI必开）

```bash
make test-race
make bench-guard
```

说明：
- `test-race` 会开启 `-race` 并强制 Redis 集成用例。
- `bench-guard` 会校验核心 benchmark 不超过预设回归阈值。

## 5. 指定Redis地址

```bash
REDIS_ADDR=127.0.0.1:6379 make test-redis
```

## 6. 运行HTTP示例

```bash
go run ./example_http
```

接口：
- `GET /nextid`
- `GET /health`
