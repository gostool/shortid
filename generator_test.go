package shortid

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type stubMachineIDProvider struct {
	id       uint16
	getErr   error
	getCalls int32
	expCalls int32
}

func (p *stubMachineIDProvider) GetMachineID(ctx context.Context) (uint16, error) {
	_ = ctx
	atomic.AddInt32(&p.getCalls, 1)
	if p.getErr != nil {
		return 0, p.getErr
	}
	return p.id, nil
}

func (p *stubMachineIDProvider) SetMachineIDExpiration(ctx context.Context, machineID uint16, expiration time.Duration) error {
	_ = ctx
	_ = machineID
	_ = expiration
	atomic.AddInt32(&p.expCalls, 1)
	return nil
}

func (p *stubMachineIDProvider) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *stubMachineIDProvider) Close() error {
	return nil
}

type stubMachineIDLeaseProvider struct {
	mu            sync.Mutex
	lease         *MachineIDLease
	acquireErr    error
	renewErr      error
	renewOK       bool
	acquireCalled int32
	renewCalled   int32
}

func (p *stubMachineIDLeaseProvider) AcquireMachineIDLease(ctx context.Context, ttl time.Duration) (*MachineIDLease, error) {
	_ = ctx
	atomic.AddInt32(&p.acquireCalled, 1)
	if p.acquireErr != nil {
		return nil, p.acquireErr
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.lease == nil {
		p.lease = &MachineIDLease{MachineID: 7, Token: "lease-token", ExpiresAt: time.Now().Add(ttl)}
	}
	return p.lease, nil
}

func (p *stubMachineIDLeaseProvider) RenewMachineIDLease(ctx context.Context, lease *MachineIDLease, ttl time.Duration) (bool, error) {
	_ = ctx
	_ = lease
	_ = ttl
	atomic.AddInt32(&p.renewCalled, 1)
	if p.renewErr != nil {
		return false, p.renewErr
	}
	return p.renewOK, nil
}

func (p *stubMachineIDLeaseProvider) ReleaseMachineIDLease(ctx context.Context, lease *MachineIDLease) error {
	_ = ctx
	_ = lease
	return nil
}

func (p *stubMachineIDLeaseProvider) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *stubMachineIDLeaseProvider) Close() error {
	return nil
}

type failSequenceProvider struct{}

func (p *failSequenceProvider) GetSequence(ctx context.Context, key string) (uint16, error) {
	_ = ctx
	_ = key
	return 0, errors.New("sequence provider unavailable")
}

func (p *failSequenceProvider) SetSequenceExpiration(ctx context.Context, key string, expiration time.Duration) error {
	_ = ctx
	_ = key
	_ = expiration
	return nil
}

func (p *failSequenceProvider) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

func (p *failSequenceProvider) Close() error {
	return nil
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
		err  error
	}{
		{
			name: "invalid business type",
			cfg: Config{
				MachineID:    1,
				BusinessType: BusinessType(BusinessReservedZ + 1),
			},
			err: ErrInvalidBusinessType,
		},
		{
			name: "invalid machine id",
			cfg: Config{
				MachineID:    100,
				BusinessType: BusinessOrder,
			},
			err: ErrInvalidMachineID,
		},
		{
			name: "conflict machine and provider",
			cfg: Config{
				MachineID:         1,
				MachineIDProvider: &stubMachineIDProvider{id: 1},
				BusinessType:      BusinessOrder,
			},
			err: ErrInvalidConfig,
		},
		{
			name: "conflict provider and lease provider",
			cfg: Config{
				MachineIDProvider:      &stubMachineIDProvider{id: 1},
				MachineIDLeaseProvider: &stubMachineIDLeaseProvider{renewOK: true},
				BusinessType:           BusinessOrder,
			},
			err: ErrInvalidConfig,
		},
		{
			name: "negative lease duration",
			cfg: Config{
				MachineIDLeaseProvider: &stubMachineIDLeaseProvider{renewOK: true},
				BusinessType:           BusinessOrder,
				MachineIDLeaseDuration: -time.Second,
			},
			err: ErrInvalidConfig,
		},
		{
			name: "valid fixed machine config",
			cfg: Config{
				MachineID:    1,
				BusinessType: BusinessOrder,
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if !errors.Is(err, tt.err) {
				t.Fatalf("ValidateConfig() error = %v, want %v", err, tt.err)
			}
		})
	}
}

func TestConstructorHelpers(t *testing.T) {
	t.Run("new convenience constructor", func(t *testing.T) {
		g, err := New(1, BusinessOrder)
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		if g == nil {
			t.Fatal("New() returned nil")
		}
	})

	t.Run("must new panic on invalid config", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("MustNew() did not panic on invalid input")
			}
		}()
		_ = MustNew(100, BusinessOrder)
	})
}

func TestNewGenerator_Modes(t *testing.T) {
	t.Run("fixed machine id", func(t *testing.T) {
		g, err := NewGenerator(Config{MachineID: 3, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if !g.machineReady {
			t.Fatal("machineReady = false, want true")
		}
		if g.machineID != 3 {
			t.Fatalf("machineID = %d, want 3", g.machineID)
		}
		if g.useMachineProvider || g.useMachineLeaseProvider {
			t.Fatal("fixed mode should not enable providers")
		}
	})

	t.Run("legacy machine provider lazy init", func(t *testing.T) {
		provider := &stubMachineIDProvider{id: 5}
		g, err := NewGenerator(Config{MachineIDProvider: provider, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if g.machineReady {
			t.Fatal("machineReady = true, want false")
		}
		if !g.useMachineProvider {
			t.Fatal("useMachineProvider = false, want true")
		}
	})

	t.Run("lease provider lazy init", func(t *testing.T) {
		provider := &stubMachineIDLeaseProvider{renewOK: true}
		g, err := NewGenerator(Config{MachineIDLeaseProvider: provider, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if g.machineReady {
			t.Fatal("machineReady = true, want false")
		}
		if !g.useMachineLeaseProvider {
			t.Fatal("useMachineLeaseProvider = false, want true")
		}
	})
}

func TestGenerator_GenerateAndNextID(t *testing.T) {
	g, err := NewGenerator(Config{MachineID: 1, BusinessType: BusinessOrder})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	raw, err := g.NextID(context.Background())
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if raw == 0 {
		t.Fatal("NextID() returned 0")
	}

	shortID, err := g.Generate()
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if shortID == "" {
		t.Fatal("Generate() returned empty string")
	}

	rawStrGen, err := NewGenerator(Config{MachineID: 1, BusinessType: BusinessOrder, ReturnRawID: true})
	if err != nil {
		t.Fatalf("NewGenerator(raw) error = %v", err)
	}
	rawStr, err := rawStrGen.GenerateWithContext(context.Background())
	if err != nil {
		t.Fatalf("GenerateWithContext(raw) error = %v", err)
	}
	if rawStr == "" {
		t.Fatal("GenerateWithContext(raw) returned empty string")
	}
	if _, parseErr := fmt.Sscanf(rawStr, "%d", new(uint64)); parseErr != nil {
		t.Fatalf("raw output is not numeric: %q", rawStr)
	}
}

func TestGenerator_ConcurrentUniqueness(t *testing.T) {
	g, err := NewGenerator(Config{MachineID: 1, BusinessType: BusinessOrder})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	const workers = 200
	ids := make(chan uint64, workers)
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			id, err := g.NextID(context.Background())
			if err != nil {
				errCh <- err
				return
			}
			ids <- id
			errCh <- nil
		}()
	}

	uniq := make(map[uint64]struct{}, workers)
	for i := 0; i < workers; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
		id := <-ids
		uniq[id] = struct{}{}
	}
	if len(uniq) != workers {
		t.Fatalf("unique ids = %d, want %d", len(uniq), workers)
	}
}

func TestGenerator_MachineProviderLifecycle(t *testing.T) {
	t.Run("initialized once for zero machine id", func(t *testing.T) {
		provider := &stubMachineIDProvider{id: 0}
		g, err := NewGenerator(Config{MachineIDProvider: provider, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		for i := 0; i < 8; i++ {
			if _, err := g.NextID(context.Background()); err != nil {
				t.Fatalf("NextID() error = %v", err)
			}
		}

		if got := atomic.LoadInt32(&provider.getCalls); got != 1 {
			t.Fatalf("GetMachineID() calls = %d, want 1", got)
		}
		if got := atomic.LoadInt32(&provider.expCalls); got != 1 {
			t.Fatalf("SetMachineIDExpiration() calls = %d, want 1", got)
		}
	})

	t.Run("initialized once under concurrency", func(t *testing.T) {
		provider := &stubMachineIDProvider{id: 9}
		g, err := NewGenerator(Config{MachineIDProvider: provider, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		const workers = 64
		var wg sync.WaitGroup
		errCh := make(chan error, workers)
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				_, err := g.NextID(context.Background())
				errCh <- err
			}()
		}
		wg.Wait()
		close(errCh)

		for err := range errCh {
			if err != nil {
				t.Fatalf("NextID() error = %v", err)
			}
		}
		if got := atomic.LoadInt32(&provider.getCalls); got != 1 {
			t.Fatalf("GetMachineID() calls = %d, want 1", got)
		}
	})

	t.Run("provider failure bubbles up", func(t *testing.T) {
		provider := &stubMachineIDProvider{getErr: errors.New("machine provider unavailable")}
		g, err := NewGenerator(Config{MachineIDProvider: provider, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if _, err := g.NextID(context.Background()); err == nil {
			t.Fatal("NextID() error = nil, want non-nil")
		}
	})
}

func TestGenerator_MachineLeaseLifecycle(t *testing.T) {
	t.Run("acquire once", func(t *testing.T) {
		provider := &stubMachineIDLeaseProvider{renewOK: true}
		g, err := NewGenerator(Config{
			MachineIDLeaseProvider: provider,
			BusinessType:           BusinessOrder,
			MachineIDLeaseDuration: time.Minute,
		})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		for i := 0; i < 8; i++ {
			if _, err := g.NextID(context.Background()); err != nil {
				t.Fatalf("NextID() error = %v", err)
			}
		}
		if got := atomic.LoadInt32(&provider.acquireCalled); got != 1 {
			t.Fatalf("AcquireMachineIDLease() calls = %d, want 1", got)
		}
	})

	t.Run("renew when due", func(t *testing.T) {
		provider := &stubMachineIDLeaseProvider{renewOK: true}
		g, err := NewGenerator(Config{
			MachineIDLeaseProvider: provider,
			BusinessType:           BusinessOrder,
			MachineIDLeaseDuration: 6 * time.Millisecond,
		})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if _, err := g.NextID(context.Background()); err != nil {
			t.Fatalf("first NextID() error = %v", err)
		}
		time.Sleep(4 * time.Millisecond)
		if _, err := g.NextID(context.Background()); err != nil {
			t.Fatalf("second NextID() error = %v", err)
		}
		if got := atomic.LoadInt32(&provider.renewCalled); got < 1 {
			t.Fatalf("RenewMachineIDLease() calls = %d, want >= 1", got)
		}
	})

	t.Run("renew lost returns sentinel error", func(t *testing.T) {
		provider := &stubMachineIDLeaseProvider{renewOK: false}
		g, err := NewGenerator(Config{
			MachineIDLeaseProvider: provider,
			BusinessType:           BusinessOrder,
			MachineIDLeaseDuration: 6 * time.Millisecond,
		})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}
		if _, err := g.NextID(context.Background()); err != nil {
			t.Fatalf("first NextID() error = %v", err)
		}
		time.Sleep(4 * time.Millisecond)
		if _, err := g.NextID(context.Background()); !errors.Is(err, ErrMachineIDLeaseLost) {
			t.Fatalf("second NextID() error = %v, want ErrMachineIDLeaseLost", err)
		}
	})

	t.Run("acquire unavailable returns sentinel error", func(t *testing.T) {
		g, err := NewGenerator(Config{
			MachineIDLeaseProvider: machineIDLeaseProviderFunc(func(ctx context.Context, ttl time.Duration) (*MachineIDLease, error) {
				_ = ctx
				_ = ttl
				return nil, nil
			}),
			BusinessType:           BusinessOrder,
			MachineIDLeaseDuration: time.Second,
		})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		if _, err := g.NextID(context.Background()); !errors.Is(err, ErrMachineIDLeaseUnavailable) {
			t.Fatalf("NextID() error = %v, want ErrMachineIDLeaseUnavailable", err)
		}
	})
}

func TestGenerator_SequenceProviderFailure(t *testing.T) {
	g, err := NewGenerator(Config{
		MachineID:        1,
		BusinessType:     BusinessOrder,
		SequenceProvider: &failSequenceProvider{},
	})
	if err != nil {
		t.Fatalf("NewGenerator() error = %v", err)
	}

	// 进入同时间片分支，从而调用 SequenceProvider。
	g.startTime = time.Now().UnixMilli()/10 - 1
	g.elapsedTime = 1

	if _, err := g.NextID(context.Background()); err == nil {
		t.Fatal("NextID() error = nil, want non-nil")
	}
}

func TestGenerator_TimeBehavior(t *testing.T) {
	t.Run("sequence overflow advances elapsed time", func(t *testing.T) {
		g, err := NewGenerator(Config{MachineID: 1, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		nowUnit := time.Now().UnixMilli() / 10
		g.startTime = nowUnit
		g.elapsedTime = 0
		g.sequence = SnowflakeMaxSequence

		if _, err := g.NextID(context.Background()); err != nil {
			t.Fatalf("NextID() error = %v", err)
		}
		if g.elapsedTime < 1 {
			t.Fatalf("elapsedTime = %d, want >= 1", g.elapsedTime)
		}
	})

	t.Run("clock backward still monotonic per instance", func(t *testing.T) {
		g, err := NewGenerator(Config{MachineID: 1, BusinessType: BusinessOrder})
		if err != nil {
			t.Fatalf("NewGenerator() error = %v", err)
		}

		current := time.Now().UnixMilli()/10 - g.startTime
		g.elapsedTime = current + 10
		g.sequence = 0

		first, err := g.NextID(context.Background())
		if err != nil {
			t.Fatalf("first NextID() error = %v", err)
		}
		second, err := g.NextID(context.Background())
		if err != nil {
			t.Fatalf("second NextID() error = %v", err)
		}
		if second <= first {
			t.Fatalf("second id (%d) should be greater than first id (%d)", second, first)
		}
	})
}

type machineIDLeaseProviderFunc func(ctx context.Context, ttl time.Duration) (*MachineIDLease, error)

func (f machineIDLeaseProviderFunc) AcquireMachineIDLease(ctx context.Context, ttl time.Duration) (*MachineIDLease, error) {
	return f(ctx, ttl)
}

func (f machineIDLeaseProviderFunc) RenewMachineIDLease(ctx context.Context, lease *MachineIDLease, ttl time.Duration) (bool, error) {
	_ = ctx
	_ = lease
	_ = ttl
	return true, nil
}

func (f machineIDLeaseProviderFunc) ReleaseMachineIDLease(ctx context.Context, lease *MachineIDLease) error {
	_ = ctx
	_ = lease
	return nil
}

func (f machineIDLeaseProviderFunc) HealthCheck(ctx context.Context) error {
	_ = ctx
	return nil
}

func (f machineIDLeaseProviderFunc) Close() error {
	return nil
}
