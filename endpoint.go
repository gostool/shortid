package shortid

import (
	"context"
	"fmt"
)

// Endpoint defines transport-agnostic entry points for adapters (HTTP/gRPC/etc).
// It keeps framework-specific code out of the core SDK.
type Endpoint struct {
	generator *Generator
}

// NewEndpoint creates a reusable endpoint facade for adapters.
func NewEndpoint(generator *Generator) *Endpoint {
	return &Endpoint{generator: generator}
}

// NextID generates a raw uint64 ID.
func (e *Endpoint) NextID(ctx context.Context) (uint64, error) {
	if e == nil || e.generator == nil {
		return 0, fmt.Errorf("endpoint or generator is nil")
	}
	return e.generator.NextID(ctx)
}

// Health checks dependencies required by the generator.
func (e *Endpoint) Health(ctx context.Context) error {
	if e == nil || e.generator == nil {
		return fmt.Errorf("endpoint or generator is nil")
	}
	if e.generator.machineIDProvider != nil {
		if err := e.generator.machineIDProvider.HealthCheck(ctx); err != nil {
			return fmt.Errorf("machine id provider health check failed: %w", err)
		}
	}
	if e.generator.machineIDLeaseProvider != nil {
		if err := e.generator.machineIDLeaseProvider.HealthCheck(ctx); err != nil {
			return fmt.Errorf("machine id lease provider health check failed: %w", err)
		}
	}
	if e.generator.sequenceProvider != nil {
		if err := e.generator.sequenceProvider.HealthCheck(ctx); err != nil {
			return fmt.Errorf("sequence provider health check failed: %w", err)
		}
	}
	return nil
}
