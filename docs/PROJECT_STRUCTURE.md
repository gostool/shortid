# 项目结构说明

## 目录结构

```
shortid/
├── go.mod                 # Go 模块定义文件
├── README.md              # 项目说明文档
├── .gitignore            # Git 忽略文件配置
│
├── codec.go              # Base62/Base91 编码核心实现
├── timestamp.go          # 时间戳编码实现
├── id.go                 # ID 生成器实现
├── sdk.go                # 便捷 API 接口
│
├── codec_test.go         # 编码功能测试（如果存在）
├── timestamp_test.go     # 时间戳功能测试
├── id_test.go            # ID 生成器测试
├── conv_test.go          # 转换功能测试
├── conv_sdk_test.go      # SDK 功能测试
├── base62_table_test.go  # Base62 表测试
│
└── examples/             # 示例程序目录
    ├── base62_table/            # Base62 进制对应关系表示例
    │   └── base62_table.go
    ├── conv_sdk/                # SDK 完整使用示例
    │   └── conv_sdk_example.go
    ├── convert/                  # Base62 编码示例
    │   └── convert_example.go
    ├── timestamp_min/            # 时间戳最小长度分析示例
    │   └── timestamp_min_length.go
    └── timestamp_summary/        # 时间戳编码方案总结示例
        └── timestamp_summary.go
```

## 文件说明

### 核心文件

- **codec.go**: Base62/Base91 编码解码核心实现
- **timestamp.go**: 时间戳压缩编码实现（短编码、动态编码、日期编码等）
- **id.go**: ID 生成器实现，支持业务类型、日期、机器ID等配置
- **sdk.go**: 提供便捷的快速 API，如 `QuickID()`, `QuickTimestamp()` 等

### 测试文件

所有 `*_test.go` 文件位于根目录，遵循 Go 标准测试规范。

### 示例程序

`examples/` 目录包含可运行的示例程序（`package main`），展示库的各种用法。

## 使用方式

### 作为库使用

```go
import "shortid"

// 使用便捷 API
id := shortid.QuickID(12345)
```

### 运行示例

```bash
cd examples
go run convert_example.go
```

### 运行测试

```bash
go test -v ./...
```

