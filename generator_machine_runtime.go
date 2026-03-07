package shortid

import (
	"context"
	"fmt"
	"time"
)

// ensureMachineIdentity 确保当前实例已经拥有可用的 machineID。
//
// 调用方需在持有 g.mu 的情况下调用，保证初始化与续租不会并发竞态。
func (g *Generator) ensureMachineIdentity(ctx context.Context, now time.Time) error {
	if g.useMachineLeaseProvider {
		return g.ensureMachineLease(ctx, now)
	}
	if g.useMachineProvider {
		return g.ensureMachineProvider(ctx)
	}
	return nil
}

func (g *Generator) ensureMachineLease(ctx context.Context, now time.Time) error {
	if !g.machineReady {
		lease, err := g.machineIDLeaseProvider.AcquireMachineIDLease(ctx, g.leaseDuration)
		if err != nil {
			return fmt.Errorf("failed to acquire machine id lease: %w", err)
		}
		if lease == nil {
			return ErrMachineIDLeaseUnavailable
		}
		g.machineID = lease.MachineID
		g.machineLease = lease
		g.machineReady = true
		g.leaseRenewAt = nextLeaseRenewTime(now, g.leaseDuration)
		return nil
	}

	if now.Before(g.leaseRenewAt) {
		return nil
	}

	ok, err := g.machineIDLeaseProvider.RenewMachineIDLease(ctx, g.machineLease, g.leaseDuration)
	if err != nil {
		return fmt.Errorf("failed to renew machine id lease: %w", err)
	}
	if !ok {
		return ErrMachineIDLeaseLost
	}
	if g.machineLease != nil {
		g.machineLease.ExpiresAt = now.Add(g.leaseDuration)
	}
	g.leaseRenewAt = nextLeaseRenewTime(now, g.leaseDuration)
	return nil
}

func (g *Generator) ensureMachineProvider(ctx context.Context) error {
	if g.machineReady {
		return nil
	}
	machineID, err := g.machineIDProvider.GetMachineID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get machine id: %w", err)
	}
	g.machineID = machineID
	g.machineReady = true
	_ = g.machineIDProvider.SetMachineIDExpiration(ctx, machineID, g.leaseDuration)
	return nil
}
