package shortid

import (
	"context"
	"testing"
)

func TestEndpoint_NextID(t *testing.T) {
	g, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("new generator failed: %v", err)
	}

	e := NewEndpoint(g)
	id, err := e.NextID(context.Background())
	if err != nil {
		t.Fatalf("next id failed: %v", err)
	}
	if id == 0 {
		t.Fatalf("id should not be zero")
	}
}

func TestEndpoint_Health(t *testing.T) {
	g, err := NewGenerator(Config{
		MachineID:    1,
		BusinessType: BusinessOrder,
	})
	if err != nil {
		t.Fatalf("new generator failed: %v", err)
	}

	e := NewEndpoint(g)
	if err := e.Health(context.Background()); err != nil {
		t.Fatalf("health should be ok: %v", err)
	}
}

func TestEndpoint_Nil(t *testing.T) {
	var e *Endpoint
	if _, err := e.NextID(context.Background()); err == nil {
		t.Fatalf("nil endpoint should fail")
	}
	if err := e.Health(context.Background()); err == nil {
		t.Fatalf("nil endpoint should fail")
	}
}
