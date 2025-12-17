package shortid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
)

// HTTPServer HTTP服务器，提供ID生成API
type HTTPServer struct {
	generator *Generator
	server    *http.Server

	// 统计信息
	stats *ServerStats
}

// ServerStats 服务器统计信息
type ServerStats struct {
	mu sync.RWMutex

	// 请求统计
	TotalRequests   int64 // 总请求数
	SuccessRequests int64 // 成功请求数
	FailedRequests  int64 // 失败请求数

	// 响应时间统计
	TotalResponseTime int64 // 总响应时间（纳秒）
	MinResponseTime   int64 // 最小响应时间（纳秒）
	MaxResponseTime   int64 // 最大响应时间（纳秒）

	// 时间窗口统计（用于计算QPS）
	windowStartTime time.Time
	windowRequests  int64

	// 服务器信息
	startTime time.Time
}

// HealthResponse 健康检查响应
type HealthResponse struct {
	Status string `json:"status"`

	// 请求统计
	TotalRequests   int64   `json:"total_requests"`
	SuccessRequests int64   `json:"success_requests"`
	FailedRequests  int64   `json:"failed_requests"`
	SuccessRate     float64 `json:"success_rate"`

	// 性能统计
	QPS             float64 `json:"qps"`               // 每秒请求数（最近1分钟）
	AvgResponseTime float64 `json:"avg_response_time"` // 平均响应时间（毫秒）
	MinResponseTime float64 `json:"min_response_time"` // 最小响应时间（毫秒）
	MaxResponseTime float64 `json:"max_response_time"` // 最大响应时间（毫秒）

	// 服务器信息
	Uptime        string `json:"uptime"`          // 运行时间
	BusinessType  string `json:"business_type"`   // 业务类型
	MachineID     uint16 `json:"machine_id"`      // 当前机器ID
	MachineIDMode string `json:"machine_id_mode"` // 机器ID模式：fixed/redis
	SequenceMode  string `json:"sequence_mode"`   // 序列号模式：local/redis

	// 健康检查
	RedisHealth string `json:"redis_health,omitempty"` // Redis健康状态
}

// IDResponse ID生成响应
type IDResponse struct {
	ID    uint64 `json:"id"`
	Error string `json:"error,omitempty"`
}

// NewHTTPServer 创建HTTP服务器
//
// 参数：
//   - addr: 服务器地址，例如 ":8080"
//   - redisAddr: Redis地址，例如 "localhost:6379"
//   - businessType: 业务类型
//
// 返回：
//   - *HTTPServer: HTTP服务器实例
//   - error: 如果创建失败，返回错误
func NewHTTPServer(addr, redisAddr string, businessType BusinessType) (*HTTPServer, error) {
	// 创建Redis机器ID提供者
	machineProvider, err := createRedisMachineIDProviderForHTTP(redisAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create machine provider: %w", err)
	}

	// // 创建Redis序列号提供者
	// sequenceProvider, err := createRedisSequenceProviderForHTTP(redisAddr)
	// if err != nil {
	// 	machineProvider.Close()
	// 	return nil, fmt.Errorf("failed to create sequence provider: %w", err)
	// }

	// 创建ID生成器
	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		// SequenceProvider:  sequenceProvider,
		BusinessType: businessType,
	})
	if err != nil {
		machineProvider.Close()
		// sequenceProvider.Close()
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	// 服务器启动时立即获取机器ID（Serverless模式）
	if generator.useMachineProvider {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		machineID, err := machineProvider.GetMachineID(ctx)
		cancel()
		if err != nil {
			machineProvider.Close()
			return nil, fmt.Errorf("failed to get machine id on startup: %w", err)
		}
		generator.machineID = machineID
		// 设置过期时间（20分钟）
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		_ = machineProvider.SetMachineIDExpiration(ctx2, machineID, 20*time.Minute)
		cancel2()
	}

	s := &HTTPServer{
		generator: generator,
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		stats: &ServerStats{
			startTime:       time.Now(),
			windowStartTime: time.Now(),
			MinResponseTime: int64(^uint64(0) >> 1), // 初始化为最大值
			MaxResponseTime: 0,                      // 初始化为0
		},
	}

	// 设置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/nextid", s.handleNextID)
	mux.HandleFunc("/health", s.handleHealth)
	s.server.Handler = mux

	return s, nil
}

// Start 启动HTTP服务器
func (s *HTTPServer) Start() error {
	return s.server.ListenAndServe()
}

// StartTLS 启动HTTPS服务器
func (s *HTTPServer) StartTLS(certFile, keyFile string) error {
	return s.server.ListenAndServeTLS(certFile, keyFile)
}

// Shutdown 优雅关闭服务器
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// handleNextID 处理NextID请求
func (s *HTTPServer) handleNextID(w http.ResponseWriter, r *http.Request) {
	// 只支持GET和POST方法
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 记录请求开始时间
	startTime := time.Now()

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 生成ID
	ctx := r.Context()
	id, err := s.generator.NextID(ctx)

	// 计算响应时间
	responseTime := time.Since(startTime)

	// 更新统计信息
	s.updateStats(err == nil, responseTime)

	if err != nil {
		response := IDResponse{
			Error: err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 返回成功响应
	response := IDResponse{
		ID: id,
	}
	json.NewEncoder(w).Encode(response)
}

// updateStats 更新统计信息
func (s *HTTPServer) updateStats(success bool, responseTime time.Duration) {
	atomic.AddInt64(&s.stats.TotalRequests, 1)

	if success {
		atomic.AddInt64(&s.stats.SuccessRequests, 1)
	} else {
		atomic.AddInt64(&s.stats.FailedRequests, 1)
	}

	// 更新响应时间统计
	responseTimeNs := int64(responseTime)
	atomic.AddInt64(&s.stats.TotalResponseTime, responseTimeNs)

	// 更新最小响应时间
	for {
		oldMin := atomic.LoadInt64(&s.stats.MinResponseTime)
		if responseTimeNs >= oldMin {
			break
		}
		if atomic.CompareAndSwapInt64(&s.stats.MinResponseTime, oldMin, responseTimeNs) {
			break
		}
	}

	// 更新最大响应时间
	for {
		oldMax := atomic.LoadInt64(&s.stats.MaxResponseTime)
		if responseTimeNs <= oldMax {
			break
		}
		if atomic.CompareAndSwapInt64(&s.stats.MaxResponseTime, oldMax, responseTimeNs) {
			break
		}
	}

	// 更新时间窗口统计（用于计算QPS）
	s.stats.mu.Lock()
	now := time.Now()
	if now.Sub(s.stats.windowStartTime) >= time.Minute {
		// 重置时间窗口
		s.stats.windowStartTime = now
		s.stats.windowRequests = 0
	}
	s.stats.windowRequests++
	s.stats.mu.Unlock()
}

// handleHealth 健康检查端点
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 获取统计信息
	stats := s.getStats()

	// 检查Redis健康状态
	redisHealth := "unknown"
	if s.generator.machineIDProvider != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := s.generator.machineIDProvider.HealthCheck(ctx); err == nil {
			redisHealth = "ok"
		} else {
			redisHealth = fmt.Sprintf("error: %v", err)
		}
	}

	// 确定机器ID模式
	machineIDMode := "fixed"
	if s.generator.useMachineProvider {
		machineIDMode = "redis"
	}

	// 确定序列号模式
	sequenceMode := "local"
	if s.generator.useSequenceProvider {
		sequenceMode = "redis"
	}

	// 计算QPS（最近1分钟）
	s.stats.mu.RLock()
	windowDuration := time.Since(s.stats.windowStartTime)
	windowRequests := s.stats.windowRequests
	s.stats.mu.RUnlock()

	var qps float64
	if windowDuration > 0 {
		qps = float64(windowRequests) / windowDuration.Seconds()
	}

	// 计算成功率
	var successRate float64
	if stats.TotalRequests > 0 {
		successRate = float64(stats.SuccessRequests) / float64(stats.TotalRequests) * 100
	}

	// 计算平均响应时间
	var avgResponseTime float64
	if stats.SuccessRequests > 0 {
		avgResponseTime = float64(stats.TotalResponseTime) / float64(stats.SuccessRequests) / 1e6 // 转换为毫秒
	}

	// 计算最小和最大响应时间（如果有请求）
	var minResponseTime, maxResponseTime float64
	if stats.SuccessRequests > 0 {
		// 检查MinResponseTime是否被初始化（如果还是初始值，说明没有成功请求）
		if stats.MinResponseTime < int64(^uint64(0)>>1) {
			minResponseTime = float64(stats.MinResponseTime) / 1e6 // 转换为毫秒
		}
		maxResponseTime = float64(stats.MaxResponseTime) / 1e6 // 转换为毫秒
	}

	// 构建响应
	response := HealthResponse{
		Status:          "ok",
		TotalRequests:   stats.TotalRequests,
		SuccessRequests: stats.SuccessRequests,
		FailedRequests:  stats.FailedRequests,
		SuccessRate:     successRate,
		QPS:             qps,
		AvgResponseTime: avgResponseTime,
		MinResponseTime: minResponseTime,
		MaxResponseTime: maxResponseTime,
		Uptime:          time.Since(s.stats.startTime).String(),
		BusinessType:    BusinessType(s.generator.businessType).String(),
		MachineID:       s.generator.machineID,
		MachineIDMode:   machineIDMode,
		SequenceMode:    sequenceMode,
		RedisHealth:     redisHealth,
	}

	json.NewEncoder(w).Encode(response)
}

// getStats 获取统计信息快照
func (s *HTTPServer) getStats() ServerStats {
	return ServerStats{
		TotalRequests:     atomic.LoadInt64(&s.stats.TotalRequests),
		SuccessRequests:   atomic.LoadInt64(&s.stats.SuccessRequests),
		FailedRequests:    atomic.LoadInt64(&s.stats.FailedRequests),
		TotalResponseTime: atomic.LoadInt64(&s.stats.TotalResponseTime),
		MinResponseTime:   atomic.LoadInt64(&s.stats.MinResponseTime),
		MaxResponseTime:   atomic.LoadInt64(&s.stats.MaxResponseTime),
	}
}

// createRedisMachineIDProviderForHTTP 创建Redis机器ID提供者（HTTP服务用）
func createRedisMachineIDProviderForHTTP(addr string) (MachineIDProvider, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &redisMachineIDProviderImpl{
		client: client,
	}, nil
}

// createRedisSequenceProviderForHTTP 创建Redis序列号提供者（HTTP服务用）
func createRedisSequenceProviderForHTTP(addr string) (SequenceProvider, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &redisSequenceProviderImpl{
		client: client,
	}, nil
}

// redisMachineIDProviderImpl 实现MachineIDProvider接口
type redisMachineIDProviderImpl struct {
	client *redis.Client
}

func (r *redisMachineIDProviderImpl) GetMachineID(ctx context.Context) (uint16, error) {
	result, err := r.client.Incr(ctx, "shortid:machine:id").Result()
	if err != nil {
		return 0, err
	}
	return uint16(result % 64), nil
}

func (r *redisMachineIDProviderImpl) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	return r.client.Expire(ctx, "shortid:machine:id", expiration).Err()
}

func (r *redisMachineIDProviderImpl) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisMachineIDProviderImpl) Close() error {
	return r.client.Close()
}

// redisSequenceProviderImpl 实现SequenceProvider接口
type redisSequenceProviderImpl struct {
	client *redis.Client
}

func (r *redisSequenceProviderImpl) GetSequence(ctx context.Context, key string) (uint16, error) {
	redisKey := "shortid:sequence:" + key
	result, err := r.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return 0, err
	}
	return uint16(result % 128), nil
}

func (r *redisSequenceProviderImpl) SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error {
	redisKey := "shortid:sequence:" + key
	return r.client.Expire(ctx, redisKey, expiration).Err()
}

func (r *redisSequenceProviderImpl) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisSequenceProviderImpl) Close() error {
	return r.client.Close()
}
