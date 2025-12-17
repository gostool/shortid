# 私有包配置指南

## 📋 概述

`shortid` 包托管在私有 GitHub 仓库 `github.com/gostool/shortid`。本文档说明如何配置 Go 环境以访问私有仓库。

## 🔧 配置步骤

### 1. 配置 GOPRIVATE 环境变量

设置 Go 私有仓库，避免 Go 尝试从公共代理获取：

```bash
# 设置 GitHub 私有仓库
export GOPRIVATE=github.com/gostool

# 或者设置为整个 gostool 组织下的所有仓库
export GOPRIVATE=github.com/gostool/*

# 永久设置（添加到 ~/.zshrc 或 ~/.bashrc）
echo 'export GOPRIVATE=github.com/gostool' >> ~/.zshrc
source ~/.zshrc
```

**验证配置**：
```bash
go env GOPRIVATE
# 应该显示: github.com/gostool
```

### 2. 配置 Git 认证

私有仓库需要认证，推荐使用 SSH 方式：

#### 方式1：使用 SSH（推荐）

```bash
# 配置 Git 使用 SSH 访问 GitHub
git config --global url."git@github.com:".insteadOf "https://github.com/"

# 验证 SSH 密钥是否已添加到 GitHub
ssh -T git@github.com
# 应该显示: Hi username! You've successfully authenticated...
```

如果没有 SSH 密钥，需要生成并添加到 GitHub：
```bash
# 生成 SSH 密钥（如果还没有）
ssh-keygen -t ed25519 -C "your_email@example.com"

# 复制公钥
cat ~/.ssh/id_ed25519.pub

# 将公钥添加到 GitHub: Settings -> SSH and GPG keys -> New SSH key
```

#### 方式2：使用 HTTPS + Personal Access Token

```bash
# 配置 Git 凭证存储
git config --global credential.helper store

# 使用 Personal Access Token 作为密码
# 创建 Token: GitHub -> Settings -> Developer settings -> Personal access tokens -> Tokens (classic)
# 权限需要: repo (Full control of private repositories)
```

### 3. 验证配置

```bash
# 1. 检查模块路径
cat go.mod | grep "^module"
# 应该显示: module github.com/gostool/shortid

# 2. 检查 GOPRIVATE 设置
go env GOPRIVATE
# 应该显示: github.com/gostool

# 3. 清理模块缓存（如果需要）
go clean -modcache

# 4. 尝试获取最新代码
go get -u github.com/gostool/shortid@latest

# 5. 或者直接下载依赖
go mod download

# 6. 运行测试验证
go test ./...
```

### 4. 在其他项目中使用

在其他项目中使用私有包时，需要：

1. **配置 GOPRIVATE**（同上）
2. **配置 Git 认证**（同上）
3. **在 go.mod 中引用**：
   ```go
   require github.com/gostool/shortid v0.0.0-20231217123456-abcdef123456
   ```
4. **运行 go mod tidy**：
   ```bash
   go mod tidy
   ```

## 🔍 常见问题

### Q1: `go get` 报错 "repository not found" 或 "authentication required"

**解决方案**：
1. 确认 GOPRIVATE 已正确设置：`go env GOPRIVATE`
2. 确认 Git 认证配置正确：`ssh -T git@github.com`
3. 确认有仓库访问权限

### Q2: `go get` 报错 "module github.com/gostool/shortid: Get ... 404 Not Found"

**解决方案**：
1. 确认仓库是私有的（不是公开的）
2. 确认 GOPRIVATE 包含 `github.com/gostool`
3. 确认 Git 认证配置正确

### Q3: 如何在 CI/CD 中使用私有包？

**GitHub Actions**：
```yaml
- name: Set up Go
  uses: actions/setup-go@v4
  with:
    go-version: '1.22'

- name: Configure Git for private repos
  run: |
    git config --global url."git@github.com:".insteadOf "https://github.com/"
    mkdir -p ~/.ssh
    echo "${{ secrets.SSH_PRIVATE_KEY }}" > ~/.ssh/id_rsa
    chmod 600 ~/.ssh/id_rsa
    ssh-keyscan github.com >> ~/.ssh/known_hosts

- name: Set GOPRIVATE
  run: echo "GOPRIVATE=github.com/gostool" >> $GITHUB_ENV

- name: Get dependencies
  run: go mod download
```

**GitLab CI**：
```yaml
variables:
  GOPRIVATE: "github.com/gostool"

before_script:
  - git config --global url."git@github.com:".insteadOf "https://github.com/"
  - mkdir -p ~/.ssh
  - echo "$SSH_PRIVATE_KEY" > ~/.ssh/id_rsa
  - chmod 600 ~/.ssh/id_rsa
  - ssh-keyscan github.com >> ~/.ssh/known_hosts
```

## 🔒 安全建议

1. **使用 SSH 密钥认证**：避免在代码中硬编码凭证
2. **使用 Personal Access Token**：如果必须使用 HTTPS，使用 Token 而不是密码
3. **限制 Token 权限**：只授予必要的权限（repo）
4. **定期轮换 Token**：定期更新 Personal Access Token
5. **使用 CI/CD 密钥管理**：在 CI/CD 中使用密钥管理服务（如 GitHub Secrets）

## 📚 参考文档

- [Go Modules 私有仓库配置](https://go.dev/ref/mod#private-modules)
- [GOPRIVATE 环境变量](https://pkg.go.dev/cmd/go#hdr-Environment_variables)
- [GitHub SSH 密钥配置](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
- [GitHub Personal Access Token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)

## ✅ 快速检查清单

- [ ] 设置 `GOPRIVATE=github.com/gostool`
- [ ] 配置 Git SSH 认证或 HTTPS Token
- [ ] 验证 SSH 连接：`ssh -T git@github.com`
- [ ] 测试获取包：`go get -u github.com/gostool/shortid@latest`
- [ ] 验证模块下载：`go mod download`
- [ ] 运行测试：`go test ./...`
