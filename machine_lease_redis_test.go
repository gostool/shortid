package shortid

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestRedisMachineIDLeaseProvider_TokenCAS(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)
	p1, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer p1.Close()

	p2, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer p2.Close()

	ctx := context.Background()
	lease, err := p1.AcquireMachineIDLease(ctx, 2*time.Second)
	if err != nil {
		t.Fatalf("AcquireMachineIDLease() error = %v", err)
	}

	forged := *lease
	forged.Token = "forged"

	ok, err := p2.RenewMachineIDLease(ctx, &forged, 2*time.Second)
	if err != nil {
		t.Fatalf("RenewMachineIDLease(forged) error = %v", err)
	}
	if ok {
		t.Fatal("RenewMachineIDLease(forged) = true, want false")
	}

	if err := p2.ReleaseMachineIDLease(ctx, &forged); err != nil {
		t.Fatalf("ReleaseMachineIDLease(forged) error = %v", err)
	}
	ok, err = p1.RenewMachineIDLease(ctx, lease, 2*time.Second)
	if err != nil {
		t.Fatalf("RenewMachineIDLease(real) error = %v", err)
	}
	if !ok {
		t.Fatal("RenewMachineIDLease(real) = false, want true")
	}
}

func TestRedisMachineIDLeaseProvider_ConfigurableSlots(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)
	provider, err := NewRedisMachineIDLeaseProviderWithConfig(redisAddr, RedisMachineIDLeaseOptions{
		Slots:          8,
		CursorKey:      "shortid:test:lease:cursor",
		LeaseKeyPrefix: "shortid:test:lease:",
	})
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProviderWithConfig() error = %v", err)
	}
	defer provider.Close()

	ctx := context.Background()
	leases := make([]*MachineIDLease, 0, 8)
	for i := 0; i < 8; i++ {
		lease, err := provider.AcquireMachineIDLease(ctx, 10*time.Second)
		if err != nil {
			t.Fatalf("AcquireMachineIDLease() #%d error = %v", i, err)
		}
		if lease.MachineID >= 8 {
			t.Fatalf("MachineID = %d, want < 8", lease.MachineID)
		}
		leases = append(leases, lease)
	}

	if _, err := provider.AcquireMachineIDLease(ctx, 10*time.Second); err == nil {
		t.Fatal("AcquireMachineIDLease() expected unavailable error when slots exhausted")
	}

	for _, lease := range leases {
		if releaseErr := provider.ReleaseMachineIDLease(ctx, lease); releaseErr != nil {
			t.Fatalf("ReleaseMachineIDLease() error = %v", releaseErr)
		}
	}
}

func TestPerf_RedisLeaseMode_SingleInstance(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)
	provider, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer provider.Close()

	g, err := NewGenerator(Config{
		MachineIDLeaseProvider: provider,
		BusinessType:           BusinessOrder,
		MachineIDLeaseDuration: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const count = 50000
	duration := runLeasePerfLoad(t, []*Generator{g}, count)
	t.Logf("[single-instance] count=%d duration=%v avg=%v qps=%.0f",
		count, duration, duration/time.Duration(count), float64(count)/duration.Seconds())
}

func TestPerf_RedisLeaseMode_TwoInstances(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)
	p1, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer p1.Close()

	p2, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer p2.Close()

	g1, err := NewGenerator(Config{
		MachineIDLeaseProvider: p1,
		BusinessType:           BusinessOrder,
		MachineIDLeaseDuration: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewGenerator(g1) error = %v", err)
	}

	g2, err := NewGenerator(Config{
		MachineIDLeaseProvider: p2,
		BusinessType:           BusinessOrder,
		MachineIDLeaseDuration: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewGenerator(g2) error = %v", err)
	}

	const each = 30000
	duration := runLeasePerfLoad(t, []*Generator{g1, g2}, each)
	total := each * 2
	t.Logf("[two-instances] total=%d duration=%v avg=%v qps=%.0f",
		total, duration, duration/time.Duration(total), float64(total)/duration.Seconds())
}

func TestPerf_RedisLeaseMode_MultiInstances(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)

	tests := []struct {
		name      string
		instances int
		each      int
	}{
		{name: "16_instances", instances: 16, each: 5000},
		{name: "32_instances", instances: 32, each: 3000},
		{name: "64_instances", instances: 64, each: 1500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgID := fmt.Sprintf("%d-%d", tt.instances, time.Now().UnixNano())
			cursorKey := "shortid:test:lease:multi:cursor:" + cfgID
			prefix := "shortid:test:lease:multi:key:" + cfgID + ":"

			providers := make([]*RedisMachineIDLeaseProvider, 0, tt.instances)
			generators := make([]*Generator, 0, tt.instances)
			for i := 0; i < tt.instances; i++ {
				p, err := NewRedisMachineIDLeaseProviderWithConfig(redisAddr, RedisMachineIDLeaseOptions{
					Slots:          64,
					CursorKey:      cursorKey,
					LeaseKeyPrefix: prefix,
				})
				if err != nil {
					t.Fatalf("NewRedisMachineIDLeaseProviderWithConfig() #%d error = %v", i, err)
				}
				providers = append(providers, p)

				g, err := NewGenerator(Config{
					MachineIDLeaseProvider: p,
					BusinessType:           BusinessOrder,
					MachineIDLeaseDuration: time.Minute,
				})
				if err != nil {
					t.Fatalf("NewGenerator() #%d error = %v", i, err)
				}
				generators = append(generators, g)
			}
			defer func() {
				for _, p := range providers {
					_ = p.Close()
				}
			}()

			duration := runLeasePerfLoad(t, generators, tt.each)
			total := tt.instances * tt.each
			t.Logf("[multi-instances] instances=%d total=%d duration=%v avg=%v qps=%.0f",
				tt.instances, total, duration, duration/time.Duration(total), float64(total)/duration.Seconds())
		})
	}
}

func runLeasePerfLoad(t *testing.T, generators []*Generator, each int) time.Duration {
	t.Helper()

	start := time.Now()
	errCh := make(chan error, len(generators))
	seen := make(map[uint64]struct{}, each*len(generators))
	var mu sync.Mutex

	run := func(g *Generator) {
		ctx := context.Background()
		for i := 0; i < each; i++ {
			id, err := g.NextID(ctx)
			if err != nil {
				errCh <- err
				return
			}
			mu.Lock()
			if _, exists := seen[id]; exists {
				mu.Unlock()
				errCh <- fmt.Errorf("duplicate id detected: %d", id)
				return
			}
			seen[id] = struct{}{}
			mu.Unlock()
		}
		errCh <- nil
	}

	for _, g := range generators {
		go run(g)
	}
	for i := 0; i < len(generators); i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("runLeasePerfLoad() error: %v", err)
		}
	}
	return time.Since(start)
}
