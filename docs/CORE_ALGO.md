# 核心算法分析

本文档详细分析 `shortid` 库的核心算法实现、时间空间复杂度以及算法特点。

## 📋 目录

1. [Base62 编码算法](#1-base62-编码算法)
2. [时间戳压缩算法](#2-时间戳压缩算法)
3. [ID 生成算法](#3-id-生成算法)
4. [日期编码算法](#4-日期编码算法)
5. [紧凑编码算法](#5-紧凑编码算法)
6. [算法性能总结](#6-算法性能总结)

---

## 1. Base62 编码算法

### 1.1 算法原理

Base62 编码将十进制数字转换为 62 进制字符串，使用字符集：
```
0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
```

### 1.2 编码实现

```go
func ToBase64(in int64) string {
    base := int64(62)
    var result []byte
    
    for in > 0 {
        remainder := in % base
        result = append([]byte{base62Chars[remainder]}, result...)
        in = in / base
    }
    
    return string(result)
}
```

### 1.3 解码实现

```go
func FromBase64(s string) (int64, error) {
    base := int64(62)
    result := int64(0)
    
    for _, char := range s {
        idx := charToIndex[char]
        result = result*base + idx
    }
    
    return result, nil
}
```

### 1.4 复杂度分析

| 操作 | 时间复杂度 | 空间复杂度 | 说明 |
|------|-----------|-----------|------|
| 编码 | O(log₆₂ n) | O(log₆₂ n) | n 为输入数字，结果长度与数字大小对数相关 |
| 解码 | O(k) | O(1) | k 为字符串长度 |

**特点**：
- ✅ **高效**：单次编码操作 < 150ns
- ✅ **压缩率高**：相比十进制字符串节省 30-50% 空间
- ✅ **URL 安全**：不使用特殊字符，可直接用于 URL
- ✅ **支持负数**：通过前缀 `-` 表示负数

### 1.5 压缩率分析

| 数字范围 | 十进制长度 | Base62长度 | 压缩率 |
|---------|-----------|-----------|--------|
| 0-61 | 1-2 | 1 | 50% |
| 62-3843 | 2-4 | 2 | 50% |
| 3844-238327 | 4-6 | 3 | 50% |
| 1000000 | 7 | 4 | 43% |
| 1000000000 | 10 | 6 | 40% |

---

## 2. 时间戳压缩算法

### 2.1 短编码算法（Short Encoding）

#### 算法原理

将时间戳分解为**天数**和**秒数**两部分，分别进行 Base62 编码：

```
时间戳 = baseline + days × 86400 + seconds
编码结果 = Base62(days) + "." + Base62(seconds)
```

**基准时间**：2020-01-01 00:00:00 UTC (1577836800)

#### 实现

```go
func ToTimestampShort(ts int64) string {
    baseline := 1577836800  // 2020-01-01
    diff := ts - baseline
    days := diff / 86400
    seconds := diff % 86400
    
    return ToBase64(days) + "." + ToBase64(seconds)
}
```

#### 复杂度

- **时间复杂度**：O(1) - 固定时间操作
- **空间复杂度**：O(1) - 结果长度固定（约 6-8 字符）

#### 压缩效果

| 时间范围 | 原始长度 | 压缩后长度 | 压缩率 |
|---------|---------|-----------|--------|
| 2020-01-01 | 10 | 3 ("1.0") | 70% |
| 2024-01-01 | 10 | 6 ("nz.0") | 40% |
| 2025-06-15 | 10 | 7 ("w7.cuc") | 30% |

### 2.2 动态编码算法（Dynamic Encoding）

#### 算法原理

相对当前时间编码，使用时间单位（分钟/小时/天）表示：

```
diff = target_time - now
if diff < 60: 返回 "m" + Base62(分钟数)
if diff < 3600: 返回 "h" + Base62(小时数)
if diff < 86400: 返回 "d" + Base62(天数)
```

#### 特点

- ✅ **极短**：1-3 个字符
- ✅ **相对时间**：适合短期场景（缓存过期、会话等）
- ⚠️ **需要基准时间**：解码时需要知道编码时的当前时间

### 2.3 紧凑编码算法（Compact Encoding）

#### 算法原理

将时间戳分解为**年份偏移**、**年内天数**、**当天秒数**三部分：

```
格式：YY + DD + SSS (7字符)
- YY: 年份偏移（2000年为基准，2字符，Base62）
- DD: 年内天数（1-366，2字符，Base62）
- SSS: 当天秒数（0-86399，3字符，Base62）
```

#### 实现

```go
func encodeCompact(ts int64) string {
    t := time.Unix(ts, 0).UTC()
    yearOffset := t.Year() - 2000
    dayOfYear := getDayOfYear(t)
    timeCode := t.Hour()*3600 + t.Minute()*60 + t.Second()
    
    yearStr := encodeWithFixedWidth(yearOffset, 62)   // 2字符
    dayStr := encodeWithFixedWidth(dayOfYear, 366)     // 2字符
    timeStr := encodeWithVariableWidth(timeCode, 86400) // 3字符
    
    return yearStr + dayStr + timeStr  // 总长度7字符
}
```

#### 复杂度

- **时间复杂度**：O(1)
- **空间复杂度**：O(1) - 固定 7 字符
- **精度**：秒级精度，支持 2000-2100 年

---

## 3. ID 生成算法

### 3.1 算法原理

ID 生成器采用**分段组合**策略，将 ID 分解为多个部分：

```
ID = [Prefix] + BusinessType + TimePart + [MachineID] + [Sequence]
```

**组成部分**：
- **Prefix**（可选）：自定义前缀
- **BusinessType**（1字符）：业务类型标识
- **TimePart**：时间戳编码（日期或完整时间戳）
- **MachineID**（可选）：机器标识，用于分布式环境
- **Sequence**（可选）：序列号，同一秒内的递增序号

### 3.2 实现流程

```go
func GenerateWithTimestamp(ts int64) string {
    // 1. 获取当前时间
    now := time.Unix(ts, 0).UTC()
    
    // 2. 处理序列号（同一秒内递增）
    if now.Unix() == lastTime {
        sequence++
    } else {
        lastTime = now.Unix()
        sequence = 0
    }
    
    // 3. 构建ID各部分
    id := prefix + businessType
    
    if enableDate {
        id += EncodeDate(now)  // 日期编码（1-3字符）
    } else {
        id += ToBase64(now.Unix())  // 完整时间戳（6字符）
    }
    
    if machineID > 0 {
        id += ToBase64(machineID)
    }
    
    if sequence > 0 {
        id += ToBase64(sequence)
    }
    
    return id
}
```

### 3.3 序列号管理

**算法**：使用互斥锁保护，同一秒内序列号递增

```go
g.mu.Lock()
defer g.mu.Unlock()

if now.Unix() == g.lastTime {
    g.sequence++  // 同一秒内递增
} else {
    g.lastTime = now.Unix()
    g.sequence = 0  // 新秒重置
}
```

**特点**：
- ✅ **线程安全**：使用 `sync.RWMutex` 保护
- ✅ **唯一性保证**：时间戳 + 机器ID + 序列号
- ⚠️ **锁竞争**：高并发下可能成为瓶颈

**优化建议**：
- 使用多个生成器实例（不同 MachineID）
- 批量生成减少锁竞争
- 考虑无锁设计（atomic 操作）

### 3.4 复杂度分析

| 操作 | 时间复杂度 | 空间复杂度 | 说明 |
|------|-----------|-----------|------|
| 生成ID | O(1) | O(k) | k 为 ID 长度（通常 4-10 字符） |
| 解析ID | O(k) | O(1) | k 为 ID 长度 |

**性能**：
- 单次生成：~287.9 ns/op
- 吞吐量：~3,473,000 ops/sec（单线程）

---

## 4. 日期编码算法

### 4.1 算法原理

相对基准日期（2024-12-31）编码天数差：

```
days = (target_date - baseline_date) / 86400
if days == 0: 返回 "0"
if 0 < days < 62: 返回 Base62字符（1字符）
if days >= 62: 返回 "+" + Base62(days)（2-3字符）
if days < 0: 返回 "-" + Base62(-days)（2-3字符）
```

### 4.2 实现

```go
func EncodeDate(year, month, day int) string {
    baseline := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
    target := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
    days := int(target.Sub(baseline).Hours() / 24)
    
    if days == 0 {
        return "0"
    }
    
    if days > 0 {
        if days < 62 {
            return string(base62Chars[days])  // 1字符
        }
        return "+" + ToBase64(days)  // 2-3字符
    } else {
        return "-" + ToBase64(-days)  // 2-3字符
    }
}
```

### 4.3 编码示例

| 日期 | 天数差 | 编码结果 | 长度 |
|------|--------|---------|------|
| 2024-12-31 | 0 | "0" | 1 |
| 2025-01-01 | 1 | "1" | 1 |
| 2025-02-01 | 32 | "W" | 1 |
| 2025-06-15 | 166 | "+2G" | 3 |
| 2024-12-30 | -1 | "-1" | 2 |

### 4.4 复杂度

- **时间复杂度**：O(1)
- **空间复杂度**：O(1) - 结果长度 1-3 字符

**优势**：
- ✅ **极短**：1-3 字符表示日期
- ✅ **可排序**：编码结果按日期顺序排列
- ✅ **易解析**：前缀 `+/-` 表示方向

---

## 5. 紧凑编码算法

### 5.1 固定宽度编码

#### 算法原理

将数字编码为固定宽度的 Base62 字符串（通常 2 字符）：

```go
func encodeWithFixedWidth(num, max int64) string {
    base := 62
    width := 2
    
    // 从低位到高位编码
    for i := 0; i < width; i++ {
        remainder := num % base
        result = append([]byte{base62Chars[remainder]}, result...)
        num = num / base
    }
    
    // 左侧补零
    for len(result) < width {
        result = append([]byte{'0'}, result...)
    }
    
    return string(result)
}
```

#### 解码

```go
func decodeWithFixedWidth(s string, max int64) (int64, error) {
    base := 62
    result := charToIndex[s[0]]*base + charToIndex[s[1]]
    
    if result >= max {
        return 0, fmt.Errorf("value out of range")
    }
    
    return result, nil
}
```

#### 应用场景

- **年份偏移**：2 字符表示 0-100（2000-2100年）
- **年内天数**：2 字符表示 1-366
- **时间码**：2 字符表示 0-1439（分钟）或 0-86399（秒）

### 5.2 可变宽度编码

#### 算法原理

根据数字大小自动选择 1-3 字符编码：

```go
func encodeWithVariableWidth(num, max int64) string {
    base := 62
    
    // 计算所需宽度
    var width int
    if num < base {          // 0-61
        width = 1
    } else if num < base*base {  // 62-3843
        width = 2
    } else {                 // 3844-238327
        width = 3
    }
    
    // 编码
    for i := 0; i < width; i++ {
        remainder := num % base
        result = append([]byte{base62Chars[remainder]}, result...)
        num = num / base
    }
    
    return string(result)
}
```

#### 特点

- ✅ **自适应**：根据数值大小选择最优宽度
- ✅ **节省空间**：小数字使用更少字符
- ⚠️ **解码复杂**：需要知道原始宽度或通过上下文推断

---

## 6. 算法性能总结

### 6.1 时间复杂度对比

| 算法 | 编码 | 解码 | 说明 |
|------|------|------|------|
| Base62 | O(log₆₂ n) | O(k) | k 为字符串长度 |
| 短时间戳 | O(1) | O(1) | 固定操作 |
| 动态时间 | O(1) | O(1) | 固定操作 |
| 紧凑编码 | O(1) | O(1) | 固定操作 |
| 日期编码 | O(1) | O(1) | 固定操作 |
| ID生成 | O(1) | O(k) | k 为 ID 长度 |

### 6.2 空间复杂度对比

| 算法 | 编码结果长度 | 压缩率 |
|------|------------|--------|
| Base62 | log₆₂ n | 30-50% |
| 短时间戳 | 6-8 字符 | 30-40% |
| 动态时间 | 1-3 字符 | 70-90% |
| 紧凑编码 | 7 字符（固定） | 30% |
| 日期编码 | 1-3 字符 | 70-90% |
| ID生成 | 4-10 字符 | 40-60% |

### 6.3 实际性能数据

基于 Apple M1 Pro 的基准测试：

```
BenchmarkPerformance/base64_encode-8          9361320    121.3 ns/op
BenchmarkPerformance/timestamp_encode-8       8678692    140.4 ns/op
BenchmarkPerformance/id_generate-8            4306465    287.9 ns/op

BenchmarkTimestampEncoders/dynamic-8         13585750     88.51 ns/op  // 最快
BenchmarkTimestampEncoders/short-8           8578692    140.6 ns/op
BenchmarkTimestampEncoders/base62-8           7167398    162.1 ns/op
```

### 6.4 算法特点总结

#### 优势

1. **高性能**
   - 单次操作 < 150ns
   - 无复杂计算，主要是字符串操作
   - 无外部依赖，纯 Go 实现

2. **高压缩率**
   - Base62 相比十进制节省 30-50%
   - 时间戳压缩节省 30-70%
   - 日期编码仅需 1-3 字符

3. **URL 安全**
   - 所有字符都是 URL 安全字符
   - 无需额外编码即可用于 URL

4. **灵活配置**
   - 支持多种编码模式
   - 可自定义基准时间、机器ID等
   - 支持业务类型标识

#### 限制

1. **ID 生成器锁竞争**
   - 高并发下可能成为瓶颈
   - 建议使用多实例或批量生成

2. **动态编码需要基准时间**
   - 解码时需要知道编码时的当前时间
   - 不适合长期存储

3. **紧凑编码年份限制**
   - 仅支持 2000-2100 年（100年范围）
   - 超出范围回退到 Base62 编码

### 6.5 优化建议

1. **减少锁竞争**
   - 使用多个生成器实例（不同 MachineID）
   - 批量生成减少锁获取次数
   - 考虑无锁设计（atomic 操作）

2. **缓存优化**
   - 缓存日期编码结果（同一日期多次编码）
   - 复用生成器实例

3. **批量操作**
   - 批量生成 ID 时复用生成器
   - 减少函数调用开销

---

## 📚 参考文献

- [Base62 编码原理](https://en.wikipedia.org/wiki/Base62)
- [时间戳压缩算法](https://en.wikipedia.org/wiki/Timestamp)
- [分布式 ID 生成策略](https://en.wikipedia.org/wiki/Snowflake_ID)

---

**最后更新**：2024-12-16

