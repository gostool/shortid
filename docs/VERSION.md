# 版本说明

## 当前版本

**v1.0.6** - 最新稳定版

## 版本历史
### v1.0.6 (最新)

#### 修复和改进
- 🔧 **类型统一**: 将所有 Base62 相关函数的 `int64` 类型统一改为 `uint64`
  - `EncodeBase62Int(num uint64)` - 参数类型从 `int64` 改为 `uint64`
  - `DecodeBase62(s string) (uint64, error)` - 返回值类型从 `int64` 改为 `uint64`
  - `GenerateIDBatch` 和 `GenerateIDBatchWithContext` 返回类型从 `map[int64]string` 改为 `map[uint64]string`
- 🔧 **修复类型转换**: 修复了 `timestamp.go` 中所有相关的类型转换问题
  - 修复了 `EncodeBase62Int` 调用时的类型转换（处理负数情况）
  - 修复了 `DecodeBase62` 调用时的类型转换（添加溢出检查）
- 📝 **文档更新**: 更新了所有相关文档中的函数签名说明

#### 向后兼容性
- ⚠️ **破坏性变更**: 此版本包含类型变更，可能影响使用 `int64` 类型的代码
- 建议：将代码中的 `int64` 类型改为 `uint64` 以匹配新的 API

---

### v1.0.5

#### 新增功能
- ✨ 新增批量ID生成方法，支持一次性生成多个ID，提升批量场景下的性能
  - `GenerateBatch(count int) ([]string, error)` - 批量生成ID（固定机器ID模式）
    - 返回ID字符串数组（短ID或数字ID字符串，取决于配置）
    - 适用于固定机器ID的传统部署场景
  - `GenerateBatchWithContext(ctx context.Context, count int) ([]string, error)` - 批量生成ID（支持Serverless模式）
    - 支持Serverless模式和分布式序列号
    - 返回ID字符串数组（短ID或数字ID字符串，取决于配置）
  - `GenerateIDBatch(count int) (map[uint64]string, error)` - 批量生成ID并返回map（固定机器ID模式）
    - 返回map结构，key为数字ID（uint64），value为Base62编码字符串
    - 便于通过数字ID快速查找对应的Base62编码
  - `GenerateIDBatchWithContext(ctx context.Context, count int) (map[uint64]string, error)` - 批量生成ID并返回map（支持Serverless模式）
    - 支持Serverless模式和分布式序列号
    - 返回map结构，key为数字ID（uint64），value为Base62编码字符串
  - `NextIDBatch(ctx context.Context, count int) ([]uint64, error)` - 批量生成原始数字ID（uint64）
    - 返回64位数字ID数组，不进行Base62编码
    - 适用于只需要数字ID的场景
  - 支持最大批量数量限制（MaxBatchCount = 10000）
  - 自动预分配切片和map容量，优化内存使用和性能

#### 改进
- 优化批量生成性能，减少重复的机器ID获取操作
- 改进错误处理，提供更详细的错误信息（包含失败索引位置）
- 统一批量生成方法的错误处理机制

---

### v1.0.4 

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

fmt.Println(shortid.Version) // 输出: 1.0.5
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

### 从 v1.0.4 升级到 v1.0.5

无需修改现有代码，新版本完全向后兼容。

如果需要使用新功能批量生成ID：

```go
// 批量生成短ID（固定机器ID模式）
ids, err := generator.GenerateBatch(100)
if err != nil {
    // 处理错误
}

// 批量生成短ID（Serverless模式）
ids, err := generator.GenerateBatchWithContext(context.Background(), 100)
if err != nil {
    // 处理错误
}

// 批量生成原始数字ID
numIDs, err := generator.NextIDBatch(context.Background(), 100)
if err != nil {
    // 处理错误
}

// 批量生成ID并返回map（同时包含数字ID和Base62编码）
result, err := generator.GenerateIDBatch(100)
if err != nil {
    // 处理错误
}
// result: map[uint64]string - key为数字ID，value为Base62编码字符串
for id, b62Str := range result {
    // 使用 id 和 b62Str
}

// Serverless模式
result, err := generator.GenerateIDBatchWithContext(context.Background(), 100)
if err != nil {
    // 处理错误
}
```

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

**最后更新**: 2025-12-18

