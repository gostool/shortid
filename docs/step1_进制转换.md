# 任意进制转换

* django.core.singning.b62_encode
* https://github.com/pyhuo/bases

```python
import time

from django.core.signing import b62_encode, b62_decode

def main():
    s = b62_encode(int(time.time()))
    print(s)
    print(b62_decode(s))
    pass

if __name__ == '__main__':
    main()
```

# 实现base.go 

## 功能实现
* ✅ 支持 Base62 进制转换（数字+大小写字母，62个字符）
* ✅ 支持 Base36 进制转换（数字+小写字母，36个字符）
* ✅ 支持 Base58 进制转换（Bitcoin风格，去除易混淆字符，58个字符）
* ✅ 通用进制转换函数 `EncodeWithBase` / `DecodeWithBase`（支持任意字符集）

## 核心函数
* `EncodeBase62(num uint64) string` - Base62编码
* `DecodeBase62(s string) (uint64, error)` - Base62解码为uint64
* `DecodeBase62ToUint(s string) (uint64, error)` - Base62解码为uint64
* `EncodeWithBase(num uint64, charset string) string` - 通用编码函数
* `DecodeWithBase(s string, charset string) (uint64, error)` - 通用解码函数

## 工具函数
* `IsValidBase62(s string) bool` - 验证Base62字符串有效性
* `Base62Length(num uint64) int` - 计算数字编码为Base62后的长度

## 测试覆盖
* ✅ Base62编码/解码单元测试
* ✅ Base58编码测试
* ✅ 编码解码往返测试（确保可逆性）
* ✅ 验证函数和长度计算测试
* ✅ 性能基准测试

## 实现说明
* 使用标准库实现，无外部依赖
* 支持 uint64 和 int64 类型
* 完整的错误处理机制
