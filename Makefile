.PHONY: test test-verbose test-race test-redis test-ci bench-guard clean

# 确保使用本地 Go 安装，而不是自动下载 toolchain
export GOTOOLCHAIN=local
export GOROOT :=

test:
	@echo "运行测试..."
	@unset GOROOT; go test ./...

test-verbose:
	@echo "运行详细测试..."
	@unset GOROOT; go test -v

test-redis:
	@echo "运行Redis必测模式（要求Redis可用）..."
	@unset GOROOT; SHORTID_REQUIRE_REDIS_TESTS=1 go test ./...

test-race:
	@echo "运行竞态测试..."
	@unset GOROOT; SHORTID_REQUIRE_REDIS_TESTS=1 go test -race ./...

bench-guard:
	@echo "运行性能退化门禁..."
	@unset GOROOT; ./scripts/bench_guard.sh

test-ci: test-redis test-race bench-guard

clean:
	@echo "清理测试缓存..."
	@go clean -testcache
