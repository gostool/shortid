# Release Policy

## 版本规范

- 采用 SemVer：`MAJOR.MINOR.PATCH`。
- `v1` 内不做破坏性 API 变更。

## 变更分类

- `PATCH`：bugfix、文档修正、无行为变化的内部重构。
- `MINOR`：向后兼容的新能力、可选配置、新 helper。
- `MAJOR`：破坏性变更，仅在明确迁移文档后发布。

## 发布门禁

- `go test ./...`
- `SHORTID_REQUIRE_REDIS_TESTS=1 go test ./...`
- `go test -race ./...`
- `scripts/bench_guard.sh`

## 变更记录模板

每次发布在 `CHANGELOG.md` 至少包含：

- `Breaking`（若无写 `None`）
- `Deprecations`
- `Behavior Changes`
- `Bug Fixes`
- `Performance`
