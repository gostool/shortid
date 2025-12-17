# Go分布式唯一ID生成方案调研

# 分布式雪花id

* https://github.com/bwmarrin/snowflake
* https://github.com/sony/sonyflake



# 部署要求
* ecs
* k8s
* serverless



# 这个库的基础上来改造， + redis 实现 部署到serverless


## 1.分布式状态存储协议定义

```
package shortid

import (
	"context"
	"time"
)

// DistributedStateProvider 分布式状态提供者协议
// 定义了ID生成器所需的分布式状态操作能力
type DistributedStateProvider interface {
	// IncrementCounter 原子递增计数器，返回新值
	IncrementCounter(ctx context.Context, key string) (int64, error)
	
	// IncrementCounterBy 原子增加计数器指定值，返回新值
	IncrementCounterBy(ctx context.Context, key string, value int64) (int64, error)

    // IncrementCounterBy 原子增加计数器随机值，返回新值
	IncrementCounterRand(ctx context.Context, key string, min int64, max int64) (int64, error)
	
	// SetExpiration 设置键的过期时间
	SetExpiration(ctx context.Context, key string, expiration time.Duration) error
	
	// HealthCheck 健康检查，用于验证连接是否可用
	HealthCheck(ctx context.Context) error
	
	// Close 关闭连接，释放资源
	Close() error
}
```

# 基于 Sonyflake 的 Serverless ID 生成器


# 思考如何引入base62进制对唯一ID进行编码变短