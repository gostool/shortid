package shortid

import (
	"context"
	"net"
	"os"
	"testing"
	"time"
)

const (
	redisAddrEnv         = "REDIS_ADDR"
	requireRedisTestsEnv = "SHORTID_REQUIRE_REDIS_TESTS"
)

func requireRedisAddrForMachineIDTest(t *testing.T) string {
	t.Helper()

	addr := os.Getenv(redisAddrEnv)
	if addr == "" {
		addr = "localhost:6379"
	}
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		_ = conn.Close()
		return addr
	}
	if os.Getenv(requireRedisTestsEnv) == "1" {
		t.Fatalf("Redis is required but unavailable at %s (env %s=1)", addr, requireRedisTestsEnv)
	}
	t.Skipf("Redis not available at %s, skipping redis provider test", addr)
	return ""
}

func TestRedisMachineIDProvider_GetMachineID(t *testing.T) {
	addr := requireRedisAddrForMachineIDTest(t)

	provider, err := NewRedisMachineIDProvider(addr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDProvider() error = %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	id, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}
	if id > 63 {
		t.Fatalf("GetMachineID() = %d, want 0-63", id)
	}
}

func TestRedisMachineIDProvider_SetMachineIDExpiration(t *testing.T) {
	addr := requireRedisAddrForMachineIDTest(t)

	provider, err := NewRedisMachineIDProvider(addr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDProvider() error = %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	if _, err = provider.GetMachineID(ctx); err != nil {
		t.Fatalf("GetMachineID() error = %v", err)
	}
	if err = provider.SetMachineIDExpiration(ctx, 1, time.Second); err != nil {
		t.Fatalf("SetMachineIDExpiration() error = %v", err)
	}

	ttl, err := provider.client.TTL(ctx, redisMachineIDCounterKey).Result()
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("TTL() = %v, want > 0", ttl)
	}
}
