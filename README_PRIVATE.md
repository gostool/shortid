# 私有包使用说明

## 📦 包信息

- **模块路径**: `github.com/gostool/shortid`
- **仓库类型**: 私有 GitHub 仓库
- **访问方式**: 需要配置 GOPRIVATE 和 Git 认证

## 🚀 快速开始

### 1. 配置环境（首次使用）

```bash
# 设置 GOPRIVATE
export GOPRIVATE=github.com/gostool

# 配置 Git SSH（推荐）
git config --global url."git@github.com:".insteadOf "https://github.com/"

# 永久设置（添加到 ~/.zshrc）
echo 'export GOPRIVATE=github.com/gostool' >> ~/.zshrc
source ~/.zshrc
```

### 2. 获取最新代码

```bash
# 获取最新版本
go get -u github.com/gostool/shortid@latest

# 或指定版本
go get github.com/gostool/shortid@v1.0.0
```

### 3. 在项目中使用

```go
package main

import (
    "github.com/gostool/shortid"
)

func main() {
    generator, _ := shortid.NewGenerator(shortid.Config{
        MachineID:    1,
        BusinessType: shortid.BusinessOrder,
    })
    id, _ := generator.Generate()
    println(id)
}
```

## 📚 详细文档

- [私有包配置指南](docs/PRIVATE_SETUP.md) - 完整的配置说明
- [文件结构](docs/FILE_STRUCTURE.md) - 项目文件结构
- [性能测试报告](docs/PERFORMANCE_TEST.md) - 性能基准测试结果

## ⚠️ 注意事项

1. **访问权限**: 确保你的 GitHub 账号有仓库访问权限
2. **SSH 密钥**: 推荐使用 SSH 方式访问，避免每次输入密码
3. **CI/CD**: 在 CI/CD 中需要配置相应的认证方式

## 🔗 相关链接

- [Go Modules 私有仓库配置](https://go.dev/ref/mod#private-modules)
- [GitHub SSH 密钥配置](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)

