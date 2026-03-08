package shortid

import (
	"context"
	"sync"
	"testing"
)

func TestMemoryMachineIDProvider_GetMachineID(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	first, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() first error = %v", err)
	}
	if first != 1 {
		t.Fatalf("GetMachineID() first = %d, want 1", first)
	}

	second, err := provider.GetMachineID(ctx)
	if err != nil {
		t.Fatalf("GetMachineID() second error = %v", err)
	}
	if second != 2 {
		t.Fatalf("GetMachineID() second = %d, want 2", second)
	}
}

func TestMemoryMachineIDProvider_RangeAndWrap(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	seen := make(map[uint16]struct{}, 64)
	for i := 0; i < 128; i++ {
		id, err := provider.GetMachineID(ctx)
		if err != nil {
			t.Fatalf("GetMachineID() error = %v", err)
		}
		if id > 63 {
			t.Fatalf("GetMachineID() = %d, want 0-63", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != 64 {
		t.Fatalf("covered slots = %d, want 64", len(seen))
	}
}

func TestMemoryMachineIDProvider_Concurrent(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	const workers = 256
	ids := make(chan uint16, workers)
	errCh := make(chan error, workers)
	var wg sync.WaitGroup

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			id, err := provider.GetMachineID(ctx)
			if err != nil {
				errCh <- err
				return
			}
			ids <- id
			errCh <- nil
		}()
	}
	wg.Wait()
	close(errCh)
	close(ids)

	for err := range errCh {
		if err != nil {
			t.Fatalf("GetMachineID() error = %v", err)
		}
	}

	count := 0
	for id := range ids {
		count++
		if id > 63 {
			t.Fatalf("GetMachineID() = %d, want 0-63", id)
		}
	}
	if count != workers {
		t.Fatalf("results = %d, want %d", count, workers)
	}
}

func TestMemoryMachineIDProvider_OtherMethods(t *testing.T) {
	provider := NewMemoryMachineIDProvider()
	ctx := context.Background()

	if err := provider.SetMachineIDExpiration(ctx, 1, 0); err != nil {
		t.Fatalf("SetMachineIDExpiration() error = %v", err)
	}
	if err := provider.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck() error = %v", err)
	}
	if err := provider.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
