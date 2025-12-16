.PHONY: test test-verbose clean

# 确保使用本地 Go 安装，而不是自动下载 toolchain
export GOTOOLCHAIN=local
export GOROOT :=

test:
	@echo "运行测试..."
	@unset GOROOT; go test

test-verbose:
	@echo "运行详细测试..."
	@unset GOROOT; go test -v

clean:
	@echo "清理测试缓存..."
	@go clean -testcache

