# 版本说明

## 当前版本

**v1.0.4** - 最新稳定版

## 版本历史

### v1.0.4 (最新)

#### 新增功能
- ✨ 新增 `GenerateID` 方法，同时返回10进制ID和Base62编码字符串
  - 方法签名: `GenerateID(ctx context.Context) (uint64, string, error)`
  - 返回原始数字ID（uint64）和对应的Base62编码字符串
  - 支持Serverless模式和分布式序列号

#### 改进
- 优化代码结构，提供更灵活的ID获取方式

---

### v1.0.3

#### 功能
- 稳定的分布式唯一ID生成器实现
- 支持多种部署模式（单机、Serverless、分布式）
- 完善的错误处理和日志记录

---

### v1.0.0 (2024-12-17)

#### 主要功能
- 🚀 完整的分布式唯一ID生成器实现
  - 基于Sonyflake算法
  - 支持时间戳压缩和Base62编码
  - 生成8-12字符的短ID

- 🌐 HTTP服务器和API
  - RESTful API接口
  - 健康检查端点
  - 性能统计和监控

- 🔧 多种部署模式
  - 单机模式（固定机器ID）
  - Serverless模式（动态机器ID）
  - 分布式序列号模式

- 📚 完善的文档和测试
  - 完整的API文档
  - 性能测试报告
  - 使用示例和最佳实践

#### API方法
- `Generate()` - 生成短ID（字符串）
- `GenerateWithContext(ctx)` - 支持上下文的ID生成
- `NextID(ctx)` - 生成原始数字ID（uint64）

---

### v0.4.0 (2024-12-17)

#### 新增功能
- ✨ 添加HTTP服务器实现
  - 提供RESTful API接口
  - 支持批量ID生成
  - 健康检查端点

- ✨ 添加健康检查端点
  - `/health` - 健康状态检查
  - `/stats` - 性能统计信息

#### 修复
- 🔧 修复文档文件名问题
- 🔧 优化错误处理机制

---

## 版本号规则

本项目遵循 [语义化版本](https://semver.org/lang/zh-CN/) 规范：

- **主版本号（MAJOR）**: 不兼容的API修改
- **次版本号（MINOR）**: 向下兼容的功能性新增
- **修订号（PATCH）**: 向下兼容的问题修正

## 查看版本

### 在代码中获取版本

```go
import "github.com/gostool/shortid"

fmt.Println(shortid.Version) // 输出: 1.0.4
```

### 查看Git标签

```bash
# 查看所有版本标签
git tag -l

# 查看特定版本信息
git show v1.0.4
```

### 在其他项目中使用

```bash
# 获取特定版本
go get github.com/gostool/shortid@v1.0.4

# 获取最新版本
go get -u github.com/gostool/shortid@latest
```

## 升级指南

### 从 v1.0.3 升级到 v1.0.4

无需修改现有代码，新版本完全向后兼容。

如果需要使用新功能 `GenerateID`：

```go
// 旧方式（仍然支持）
id, err := generator.Generate()
if err != nil {
    // 处理错误
}

// 新方式：同时获取数字ID和Base62编码
numID, b62Str, err := generator.GenerateID(context.Background())
if err != nil {
    // 处理错误
}
// numID: 10进制ID (uint64)
// b62Str: Base62编码字符串
```

### 从 v0.4.0 升级到 v1.0.0

主要变更：
- API接口保持兼容
- 新增 `NextID` 方法用于获取原始数字ID
- 改进的错误处理机制

## 版本计划

### 未来版本规划

- [ ] v1.1.0 - 计划添加更多业务类型支持
- [ ] v1.2.0 - 计划优化性能，提升QPS
- [ ] v2.0.0 - 计划重构API，提供更灵活的配置选项

## 相关文档

- [版本测试指南](VERSION_TEST.md) - 如何验证版本是否正确发布
- [文件结构](FILE_STRUCTURE.md) - 项目文件结构说明
- [性能测试报告](PERFORMANCE_TEST.md) - 详细的性能测试结果

---

**最后更新**: 2024-12-17

