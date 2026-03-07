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
	p, err := NewRedisMachineIDLeaseProviderWithConfig(redisAddr, RedisMachineIDLeaseOptions{
		Slots:          8,
		CursorKey:      "shortid:test:lease:cursor",
		LeaseKeyPrefix: "shortid:test:lease:",
	})
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProviderWithConfig() error = %v", err)
	}
	defer p.Close()

	lease, err := p.AcquireMachineIDLease(context.Background(), time.Second)
	if err != nil {
		t.Fatalf("AcquireMachineIDLease() error = %v", err)
	}
	if lease.MachineID >= 8 {
		t.Fatalf("MachineID = %d, want < 8", lease.MachineID)
	}
}

func TestPerf_RedisLeaseMode_SingleInstance(t *testing.T) {
	redisAddr := requireRedisForSDKTest(t)
	leaseProvider, err := NewRedisMachineIDLeaseProvider(redisAddr)
	if err != nil {
		t.Fatalf("NewRedisMachineIDLeaseProvider() error = %v", err)
	}
	defer leaseProvider.Close()

	g, err := NewGenerator(Config{
		MachineIDLeaseProvider: leaseProvider,
		BusinessType:           BusinessOrder,
		MachineIDLeaseDuration: time.Minute,
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const count = 50000
	idMap := make(map[uint64]struct{}, count)
	ctx := context.Background()
	start := time.Now()
	for i := 0; i < count; i++ {
		id, genErr := g.NextID(ctx)
		if genErr != nil {
			t.Fatalf("NextID() error = %v", genErr)
		}
		if _, ok := idMap[id]; ok {
			t.Fatalf("duplicate id at %d: %d", i, id)
		}
		idMap[id] = struct{}{}
	}
	duration := time.Since(start)
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
	var mu sync.Mutex
	idMap := make(map[uint64]struct{}, each*2)
	errCh := make(chan error, 2)
	start := time.Now()

	run := func(g *Generator) {
		ctx := context.Background()
		for i := 0; i < each; i++ {
			id, genErr := g.NextID(ctx)
			if genErr != nil {
				errCh <- genErr
				return
			}
			mu.Lock()
			if _, ok := idMap[id]; ok {
				mu.Unlock()
				errCh <- fmt.Errorf("duplicate id detected: %d", id)
				return
			}
			idMap[id] = struct{}{}
			mu.Unlock()
		}
		errCh <- nil
	}

	go run(g1)
	go run(g2)
	for i := 0; i < 2; i++ {
		if runErr := <-errCh; runErr != nil {
			t.Fatalf("concurrent run error: %v", runErr)
		}
	}
	duration := time.Since(start)
	total := each * 2
	t.Logf("[two-instances] total=%d duration=%v avg=%v qps=%.0f",
		total, duration, duration/time.Duration(total), float64(total)/duration.Seconds())
}
