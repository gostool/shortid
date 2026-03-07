package shortid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultLeaseSlots = 64
	leaseCursorKey    = "shortid:machine:lease:cursor"
	leaseKeyPrefix    = "shortid:machine:lease:"
)

// RedisMachineIDLeaseOptions Redis租约提供者配置。
type RedisMachineIDLeaseOptions struct {
	// Slots 可分配槽位数，默认64，范围[1,64]。
	Slots int

	// CursorKey 游标键名（可选，默认 shortid:machine:lease:cursor）。
	CursorKey string

	// LeaseKeyPrefix 租约键前缀（可选，默认 shortid:machine:lease:）。
	LeaseKeyPrefix string
}

var acquireMachineIDLeaseLua = redis.NewScript(`
local cursor_key = KEYS[1]
local lease_prefix = ARGV[1]
local token = ARGV[2]
local ttl_ms = tonumber(ARGV[3])
local slots = tonumber(ARGV[4])

local start = redis.call('INCR', cursor_key)
for i = 0, slots - 1 do
  local id = (start + i) % slots
  local key = lease_prefix .. id
  local ok = redis.call('SET', key, token, 'NX', 'PX', ttl_ms)
  if ok then
    return id
  end
end
return -1
`)

var renewMachineIDLeaseLua = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('PEXPIRE', KEYS[1], ARGV[2])
end
return 0
`)

var releaseMachineIDLeaseLua = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
  return redis.call('DEL', KEYS[1])
end
return 0
`)

// RedisMachineIDLeaseProvider 基于 Redis 的机器ID租约实现。
//
// 安全语义：
//   - 获取：单次Lua脚本在Redis侧完成“游标递增 + 64槽位抢占”。
//   - 续租/释放：必须token匹配（CAS），防止误续租/误释放。
type RedisMachineIDLeaseProvider struct {
	client         *redis.Client
	slots          int
	cursorKey      string
	leaseKeyPrefix string
}

// NewRedisMachineIDLeaseProvider 创建 Redis 租约提供者。
func NewRedisMachineIDLeaseProvider(addr string) (*RedisMachineIDLeaseProvider, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return newRedisMachineIDLeaseProviderWithClient(client, RedisMachineIDLeaseOptions{})
}

// NewRedisMachineIDLeaseProviderWithOptions 使用自定义选项创建实现。
func NewRedisMachineIDLeaseProviderWithOptions(opts *redis.Options) (*RedisMachineIDLeaseProvider, error) {
	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return newRedisMachineIDLeaseProviderWithClient(client, RedisMachineIDLeaseOptions{})
}

// NewRedisMachineIDLeaseProviderWithConfig 使用地址 + 租约配置创建实现。
func NewRedisMachineIDLeaseProviderWithConfig(addr string, cfg RedisMachineIDLeaseOptions) (*RedisMachineIDLeaseProvider, error) {
	client := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	return newRedisMachineIDLeaseProviderWithClient(client, cfg)
}

func newRedisMachineIDLeaseProviderWithClient(client *redis.Client, cfg RedisMachineIDLeaseOptions) (*RedisMachineIDLeaseProvider, error) {
	slots := cfg.Slots
	if slots == 0 {
		slots = defaultLeaseSlots
	}
	if slots < 1 || slots > defaultLeaseSlots {
		_ = client.Close()
		return nil, fmt.Errorf("%w: Slots must be in [1,%d]", ErrInvalidConfig, defaultLeaseSlots)
	}

	cursor := cfg.CursorKey
	if cursor == "" {
		cursor = leaseCursorKey
	}
	prefix := cfg.LeaseKeyPrefix
	if prefix == "" {
		prefix = leaseKeyPrefix
	}

	return &RedisMachineIDLeaseProvider{
		client:         client,
		slots:          slots,
		cursorKey:      cursor,
		leaseKeyPrefix: prefix,
	}, nil
}

func (r *RedisMachineIDLeaseProvider) AcquireMachineIDLease(ctx context.Context, ttl time.Duration) (*MachineIDLease, error) {
	if ttl <= 0 {
		return nil, ErrInvalidConfig
	}
	token, err := newLeaseToken()
	if err != nil {
		return nil, fmt.Errorf("generate lease token failed: %w", err)
	}
	id, err := acquireMachineIDLeaseLua.Run(
		ctx,
		r.client,
		[]string{r.cursorKey},
		r.leaseKeyPrefix,
		token,
		ttl.Milliseconds(),
		r.slots,
	).Int64()
	if err != nil {
		return nil, fmt.Errorf("acquire machine id lease failed: %w", err)
	}
	if id < 0 {
		return nil, ErrMachineIDLeaseUnavailable
	}
	return &MachineIDLease{
		MachineID: uint16(id),
		Token:     token,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

func (r *RedisMachineIDLeaseProvider) RenewMachineIDLease(ctx context.Context, lease *MachineIDLease, ttl time.Duration) (bool, error) {
	if lease == nil || ttl <= 0 {
		return false, ErrInvalidConfig
	}
	res, err := renewMachineIDLeaseLua.Run(
		ctx,
		r.client,
		[]string{r.leaseKey(lease.MachineID)},
		lease.Token,
		ttl.Milliseconds(),
	).Int64()
	if err != nil {
		return false, fmt.Errorf("renew machine id lease failed: %w", err)
	}
	if res == 1 {
		lease.ExpiresAt = time.Now().Add(ttl)
		return true, nil
	}
	return false, nil
}

func (r *RedisMachineIDLeaseProvider) ReleaseMachineIDLease(ctx context.Context, lease *MachineIDLease) error {
	if lease == nil {
		return nil
	}
	if _, err := releaseMachineIDLeaseLua.Run(
		ctx,
		r.client,
		[]string{r.leaseKey(lease.MachineID)},
		lease.Token,
	).Result(); err != nil {
		return fmt.Errorf("release machine id lease failed: %w", err)
	}
	return nil
}

func (r *RedisMachineIDLeaseProvider) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisMachineIDLeaseProvider) Close() error {
	return r.client.Close()
}

func (r *RedisMachineIDLeaseProvider) leaseKey(machineID uint16) string {
	return fmt.Sprintf("%s%d", r.leaseKeyPrefix, machineID)
}

func newLeaseToken() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
