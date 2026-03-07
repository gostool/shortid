package shortid

import (
	"context"
	"testing"
	"time"
)

type generatorAPISnapshot interface {
	Generate() (string, error)
	GenerateWithContext(context.Context) (string, error)
	NextID(context.Context) (uint64, error)
}

type machineIDProviderSnapshot interface {
	GetMachineID(context.Context) (uint16, error)
	SetMachineIDExpiration(context.Context, uint16, time.Duration) error
	HealthCheck(context.Context) error
	Close() error
}

type sequenceProviderSnapshot interface {
	GetSequence(context.Context, string) (uint16, error)
	SetSequenceExpiration(context.Context, string, time.Duration) error
	HealthCheck(context.Context) error
	Close() error
}

// TestPublicAPISignatureSnapshot ensures core exported signatures remain stable.
func TestPublicAPISignatureSnapshot(t *testing.T) {
	// Core constructor/validator snapshot.
	var _ func(Config) error = ValidateConfig
	var _ func(Config) (*Generator, error) = NewGenerator

	// Generator method snapshot.
	var _ generatorAPISnapshot = (*Generator)(nil)

	// Provider interface contract snapshot.
	var _ machineIDProviderSnapshot = (MachineIDProvider)(nil)
	var _ sequenceProviderSnapshot = (SequenceProvider)(nil)

	// Timestamp/base helpers snapshot.
	var _ func(uint64) string = EncodeBase62
	var _ func(string) (int64, error) = DecodeBase62
	var _ func(int64) string = ToTimestampShort
	var _ func(string) (int64, error) = FromTimestampShort
	var _ func(int64) string = ToTimestampDynamic
	var _ func(string, int64) (int64, error) = FromTimestampDynamic
	var _ func(int64) string = ToTimestampCompact
	var _ func(string) (int64, error) = FromTimestampCompact
}
