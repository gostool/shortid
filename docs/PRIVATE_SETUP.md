# 私有包配置指南

## 📋 概述

将 `shortid` 包配置为私有包，需要修改模块路径和配置 Go 的私有仓库设置。

## 🔧 配置步骤

### 1. 修改 go.mod 模块路径

将模块路径从公共路径改为私有路径：

```go
// 原路径
module github.com/gostool/shortid

// 改为私有路径（示例）
module git.company.com/shortid
// 或
module gitea.company.com/team/shortid
// 或
module gitlab.company.com/group/shortid
```

### 2. 配置 GOPRIVATE 环境变量

设置 Go 私有仓库，避免 Go 尝试从公共代理获取：

```bash
# 设置私有仓库域名
export GOPRIVATE=git.company.com,gitea.company.com,gitlab.company.com

# 或者设置为所有私有仓库
export GOPRIVATE=*.company.com

# 永久设置（添加到 ~/.bashrc 或 ~/.zshrc）
echo 'export GOPRIVATE=git.company.com,gitea.company.com,gitlab.company.com' >> ~/.zshrc
```

### 3. 配置 Git 认证

如果私有仓库需要认证，配置 Git：

```bash
# 方式1：使用 SSH
git config --global url."git@git.company.com:".insteadOf "https://git.company.com/"

# 方式2：使用 HTTPS + 凭证
git config --global credential.helper store
```

### 4. 更新所有导入路径

如果项目中有其他文件引用了旧的模块路径，需要更新：

```bash
# 查找所有引用
grep -r "github.com/gostool/shortid" .

# 批量替换（谨慎使用）
find . -type f -name "*.go" -exec sed -i '' 's|github.com/gostool/shortid|git.company.com/shortid|g' {} +
```

### 5. 更新文档

更新文档中的模块路径引用：
- `docs/FILE_STRUCTURE.md`
- `docs/step3:go唯一id.md`
- `README.md`（如果存在）

## 📝 常见私有仓库路径格式

### GitLab
```
module gitlab.company.com/group/shortid
```

### Gitea
```
module gitea.company.com/org/shortid
```

### GitHub Enterprise
```
module github.company.com/org/shortid
```

### 自建 Git 服务器
```
module git.company.com/shortid
```

## ✅ 验证配置

```bash
# 1. 检查模块路径
cat go.mod | grep "^module"

# 2. 检查 GOPRIVATE 设置
go env GOPRIVATE

# 3. 尝试下载依赖（应该从私有仓库获取）
go mod download

# 4. 运行测试
go test ./...
```

## 🔒 安全建议

1. **使用 SSH 密钥认证**：避免在代码中硬编码凭证
2. **配置 .gitignore**：确保不提交敏感信息
3. **使用 CI/CD 变量**：在 CI/CD 中配置私有仓库访问凭证
4. **限制访问权限**：在 Git 服务器上设置适当的访问控制

## 📚 参考文档

- [Go Modules 私有仓库配置](https://go.dev/ref/mod#private-modules)
- [GOPRIVATE 环境变量](https://pkg.go.dev/cmd/go#hdr-Environment_variables)

