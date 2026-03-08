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

func requireRedisAddrForSequenceTest(t *testing.T) string {
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

func TestRedisSequenceProvider_GetSequence(t *testing.T) {
	addr := requireRedisAddrForSequenceTest(t)

	provider, err := NewRedisSequenceProvider(addr)
	if err != nil {
		t.Fatalf("NewRedisSequenceProvider() error = %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	seq, err := provider.GetSequence(ctx, "unit-seq")
	if err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if seq > 127 {
		t.Fatalf("GetSequence() = %d, want 0-127", seq)
	}
}

func TestRedisSequenceProvider_SetSequenceExpiration(t *testing.T) {
	addr := requireRedisAddrForSequenceTest(t)

	provider, err := NewRedisSequenceProvider(addr)
	if err != nil {
		t.Fatalf("NewRedisSequenceProvider() error = %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	key := "unit-exp"
	if _, err = provider.GetSequence(ctx, key); err != nil {
		t.Fatalf("GetSequence() error = %v", err)
	}
	if err = provider.SetSequenceExpiration(ctx, key, time.Second); err != nil {
		t.Fatalf("SetSequenceExpiration() error = %v", err)
	}

	ttl, err := provider.client.TTL(ctx, "shortid:sequence:"+key).Result()
	if err != nil {
		t.Fatalf("TTL() error = %v", err)
	}
	if ttl <= 0 {
		t.Fatalf("TTL() = %v, want > 0", ttl)
	}
}
