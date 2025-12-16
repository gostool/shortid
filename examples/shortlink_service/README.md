# 短链接服务示例

这是一个完整的短链接服务实现示例，展示了如何使用 `shortid` 库构建高性能的短链接服务。

## 🎯 功能特性

- ✅ 短链接生成：将长URL转换为短码
- ✅ 短链接解析：根据短码获取原始URL
- ✅ 批量生成：支持批量创建短链接
- ✅ 高并发支持：多生成器实例减少锁竞争
- ✅ 性能测试：内置性能测试和可行性评估

## 🚀 快速开始

```bash
cd examples/shortlink_service
go run shortlink_service.go
```

## 📊 性能测试结果

基于测试环境（Apple M1 Pro）的性能数据：

### 测试配置
- **写入并发数**: 100 goroutines
- **读取并发数**: 1000 goroutines（读写比 1:10）
- **每个goroutine写入**: 10,000 次
- **每个goroutine读取**: 100,000 次

### 测试结果
```
写入QPS: 907,192
读取QPS: 14,421,418
总QPS: 15,328,610
```

### 每天1亿写入可行性评估

**需求**：
- 每天写入：100,000,000（1亿）
- 每天读取：1,000,000,000（10亿，读写比 1:10）
- 需要写入QPS：1,157
- 需要读取QPS：11,574
- 需要总QPS：12,731

**实际性能**：
- 实际写入QPS：907,192
- 实际读取QPS：14,421,418
- 实际总QPS：15,328,610

**性能余量**：
- ✅ 写入性能余量：**783.8倍**
- ✅ 读取性能余量：**1,246.0倍**
- ✅ 总体性能余量：**1,204.0倍**

**结论**：✅ **完全可行！性能完全满足需求。**

## 💡 核心实现

### 1. 多生成器实例设计

使用多个生成器实例，每个使用不同的 `MachineID`，减少锁竞争：

```go
generators := make([]*shortid.DefaultIDGenerator, 10)
for i := 0; i < 10; i++ {
    generators[i] = shortid.NewIDGenerator(shortid.IDGeneratorConfig{
        BusinessType: shortid.BusinessShare,
        EnableDate:   true,
        MachineID:    int64(i),  // 不同的机器ID
    })
}
```

### 2. 轮询使用生成器

```go
genIdx := atomic.AddInt64(&s.writeCount, 1) % int64(len(s.generators))
gen := s.generators[genIdx]
shortCode := gen.Generate()
```

### 3. 批量生成优化

```go
func (s *ShortLinkService) BatchCreateShortLinks(originalURLs []string) ([]string, error) {
    // 使用单个生成器批量生成，减少锁竞争
    gen := s.generators[0]
    // ...
}
```

## 🔧 使用示例

### 创建短链接

```go
service := NewShortLinkService()

originalURL := "https://www.example.com/very/long/url/path"
shortCode, err := service.CreateShortLink(originalURL)
// shortCode: "A+5E3d7"
// 短链接: https://short.ly/A+5E3d7
```

### 解析短链接

```go
originalURL, ok := service.GetOriginalURL("A+5E3d7")
if ok {
    fmt.Println("原始URL:", originalURL)
}
```

### 批量创建

```go
urls := []string{
    "https://example.com/page1",
    "https://example.com/page2",
    "https://example.com/page3",
}
codes, err := service.BatchCreateShortLinks(urls)
```

## 📈 性能优化建议

### 1. 生产环境优化

**数据库优化**：
```go
// 批量写入数据库
const batchSize = 1000
links := make([]ShortLink, 0, batchSize)
for _, url := range urls {
    code := service.CreateShortLink(url)
    links = append(links, ShortLink{Code: code, URL: url})
    if len(links) >= batchSize {
        db.BatchInsert(links)
        links = links[:0]
    }
}
```

**缓存优化**：
```go
// 使用 Redis 缓存热点数据
func (s *ShortLinkService) GetOriginalURL(shortCode string) (string, bool) {
    // 先查缓存
    if url, err := redis.Get("link:" + shortCode); err == nil {
        return url, true
    }
    
    // 再查数据库
    url, ok := s.db.Get(shortCode)
    if ok {
        redis.Set("link:"+shortCode, url, 7*24*time.Hour)
    }
    return url, ok
}
```

### 2. 分布式部署

```go
// 每个实例使用不同的 MachineID
config := shortid.IDGeneratorConfig{
    BusinessType: shortid.BusinessShare,
    MachineID:   getMachineIDFromEnv(),  // 从环境变量获取
    EnableDate:  true,
}
```

### 3. 监控指标

- 写入QPS
- 读取QPS
- 平均延迟（P50/P95/P99）
- 错误率
- 缓存命中率

## ⚠️ 注意事项

1. **存储实现**：本示例使用内存存储（`sync.Map`），生产环境应使用数据库
2. **唯一性保证**：使用 `MachineID` 区分不同实例，保证全局唯一
3. **过期清理**：实际应用中需要定期清理过期链接
4. **缓存策略**：热点数据应使用 Redis 缓存，提升读取性能

## 📚 相关文档

- [性能分析文档](../../PERFORMANCE_ANALYSIS.md)
- [主 README](../../README.md)

