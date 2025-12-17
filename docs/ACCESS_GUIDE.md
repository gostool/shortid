# 仓库访问指南

## 🔒 访问权限说明

`github.com/gostool/shortid` 是一个**私有 GitHub 仓库**，只有被授权的用户才能访问和下载。

## ✅ 其他人如何下载

### 前提条件

1. **GitHub 账号有仓库访问权限**
   - 仓库所有者需要在 GitHub 上添加协作者（Collaborator）
   - 或者用户需要是 `gostool` 组织的成员

### 步骤1: 检查访问权限

访问仓库 URL（需要登录 GitHub）：
```
https://github.com/gostool/shortid
```

如果能看到仓库内容，说明有访问权限。

### 步骤2: 配置 Go 环境

#### 2.1 设置 GOPRIVATE

```bash
# 设置 GOPRIVATE
export GOPRIVATE=github.com/gostool

# 永久设置（添加到 ~/.zshrc 或 ~/.bashrc）
echo 'export GOPRIVATE=github.com/gostool' >> ~/.zshrc
source ~/.zshrc
```

#### 2.2 配置 Git 认证

**方式1: 使用 SSH（推荐）**

```bash
# 配置 Git 使用 SSH
git config --global url."git@github.com:".insteadOf "https://github.com/"

# 验证 SSH 连接
ssh -T git@github.com
# 应该显示: Hi username! You've successfully authenticated...
```

如果没有 SSH 密钥：
```bash
# 生成 SSH 密钥
ssh-keygen -t ed25519 -C "your_email@example.com"

# 复制公钥
cat ~/.ssh/id_ed25519.pub

# 添加到 GitHub: Settings -> SSH and GPG keys -> New SSH key
```

**方式2: 使用 HTTPS + Personal Access Token**

```bash
# 配置 Git 凭证存储
git config --global credential.helper store

# 创建 Personal Access Token
# GitHub -> Settings -> Developer settings -> Personal access tokens -> Tokens (classic)
# 权限需要: repo (Full control of private repositories)
```

### 步骤3: 下载包

```bash
# 下载指定版本
go get github.com/gostool/shortid@v1.0.1

# 或下载最新版本
go get -u github.com/gostool/shortid@latest

# 在项目中使用
go mod tidy
```

## 🔍 验证配置

```bash
# 1. 检查 GOPRIVATE
go env GOPRIVATE
# 应该显示: github.com/gostool

# 2. 检查 Git 配置
git config --global --get url."git@github.com:".insteadOf
# 应该显示: https://github.com/

# 3. 测试 SSH 连接
ssh -T git@github.com
# 应该显示成功消息

# 4. 尝试下载
go get github.com/gostool/shortid@v1.0.1
```

## ❌ 常见错误

### 错误1: "repository not found" 或 "authentication required"

**原因**: 没有仓库访问权限或认证配置不正确

**解决方案**:
1. 确认 GitHub 账号有仓库访问权限
2. 确认 GOPRIVATE 已设置：`go env GOPRIVATE`
3. 确认 Git 认证配置正确：`ssh -T git@github.com`

### 错误2: "module github.com/gostool/shortid: Get ... 404 Not Found"

**原因**: Go 尝试从公共代理获取私有仓库

**解决方案**:
1. 确认 GOPRIVATE 包含 `github.com/gostool`
2. 清除模块缓存：`go clean -modcache`
3. 重新下载：`go get github.com/gostool/shortid@v1.0.1`

### 错误3: "Permission denied (publickey)"

**原因**: SSH 密钥未配置或未添加到 GitHub

**解决方案**:
1. 检查 SSH 密钥是否存在：`ls -la ~/.ssh/`
2. 生成 SSH 密钥：`ssh-keygen -t ed25519 -C "your_email@example.com"`
3. 添加公钥到 GitHub：`cat ~/.ssh/id_ed25519.pub`

## 👥 添加协作者

如果你是仓库所有者，需要添加协作者：

1. 访问仓库：`https://github.com/gostool/shortid`
2. 点击 **Settings** -> **Collaborators**
3. 点击 **Add people**
4. 输入协作者的 GitHub 用户名或邮箱
5. 选择权限级别（通常选择 **Write**）
6. 协作者会收到邀请邮件，接受后即可访问

## 🏢 组织成员访问

如果仓库属于 `gostool` 组织：

1. 组织管理员需要将用户添加到组织
2. 设置用户的仓库访问权限
3. 用户配置 GOPRIVATE 和 Git 认证后即可访问

## 📋 快速检查清单

- [ ] GitHub 账号有仓库访问权限
- [ ] 设置 `GOPRIVATE=github.com/gostool`
- [ ] 配置 Git SSH 或 HTTPS Token
- [ ] 验证 SSH 连接：`ssh -T git@github.com`
- [ ] 测试下载：`go get github.com/gostool/shortid@v1.0.1`

## 🔗 相关文档

- [私有包配置指南](PRIVATE_SETUP.md) - 详细的配置说明
- [版本测试指南](VERSION_TEST.md) - 版本验证方法

