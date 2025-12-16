#!/bin/bash
# 运行测试 - 使用本地 Go 安装，而不是自动下载 toolchain

# 确保使用本地 Go 安装，并取消设置可能被错误设置的 GOROOT
export GOTOOLCHAIN=local
unset GOROOT

# 运行测试
go test "$@"

