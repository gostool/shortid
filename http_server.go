package shortid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// HTTPServer HTTP服务器，提供ID生成API
type HTTPServer struct {
	generator *Generator
	server    *http.Server
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

	// 创建Redis序列号提供者
	sequenceProvider, err := createRedisSequenceProviderForHTTP(redisAddr)
	if err != nil {
		machineProvider.Close()
		return nil, fmt.Errorf("failed to create sequence provider: %w", err)
	}

	// 创建ID生成器
	generator, err := NewGenerator(Config{
		MachineIDProvider: machineProvider,
		SequenceProvider:  sequenceProvider,
		BusinessType:      businessType,
	})
	if err != nil {
		machineProvider.Close()
		sequenceProvider.Close()
		return nil, fmt.Errorf("failed to create generator: %w", err)
	}

	s := &HTTPServer{
		generator: generator,
		server: &http.Server{
			Addr:         addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
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

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 生成ID
	ctx := r.Context()
	id, err := s.generator.NextID(ctx)
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

// handleHealth 健康检查端点
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
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

