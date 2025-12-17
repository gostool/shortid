# 版本测试指南

## 📋 说明

在项目目录内无法使用 `go get` 获取自己的版本，这是 Go 的正常行为。

## ✅ 验证 Tag 是否正确发布

### 方法1：检查远程 Tag

```bash
# 查看远程 tag
git ls-remote --tags origin | grep v0.4.0

# 应该显示类似：
# abc123...refs/tags/v0.4.0
```

### 方法2：查看本地 Tag

```bash
# 查看 tag 信息
git show v0.4.0

# 查看 tag 列表
git tag -l
```

### 方法3：在临时项目中测试

```bash
# 创建临时测试目录
cd /tmp
mkdir test_shortid_v0.4.0
cd test_shortid_v0.4.0

# 初始化 Go 模块
go mod init test_shortid

# 设置 GOPRIVATE（如果使用私有仓库）
export GOPRIVATE=github.com/gostool

# 获取指定版本
go get github.com/gostool/shortid@v0.4.0

# 查看 go.mod，应该包含：
# require github.com/gostool/shortid v0.4.0

# 清理
cd /tmp
rm -rf test_shortid_v0.4.0
```

## 🚀 在其他项目中使用

### 获取特定版本

```bash
# 在项目目录中
go get github.com/gostool/shortid@v0.4.0

# 或更新到最新版本
go get -u github.com/gostool/shortid@latest
```

### 在 go.mod 中指定版本

```go
require (
    github.com/gostool/shortid v0.4.0
)
```

然后运行：
```bash
go mod tidy
```

## ⚠️ 常见问题

### Q: 为什么在项目目录内运行 `go get` 会报错？

**A**: 这是 Go 的正常行为。Go 不允许在模块内部获取自己的特定版本，因为当前工作目录就是该模块本身。

### Q: 如何验证 tag 是否正确？

**A**: 使用上面的方法3，在临时项目中测试获取版本。

### Q: 如何更新到最新版本？

**A**: 在其他项目中使用 `go get -u github.com/gostool/shortid@latest`。

