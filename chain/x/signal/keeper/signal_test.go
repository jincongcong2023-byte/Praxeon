package keeper_test

import (
	"errors"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mars/x/signal/types"
)

func signalAddr(n int) string {
	return sdk.AccAddress([]byte(fmt.Sprintf("address_%012d", n))).String()
}

func TestLikeQuota(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	sender := signalAddr(1)

	// give the sender weight 10 (a Use signal from another address)
	// so its Likes pass defense ③ (sender weight floor)
	if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), signalAddr(2), sender, types.SignalTypeUse); err != nil {
		t.Fatalf("use: %v", err)
	}

	// 10 Likes should succeed (quota = 10 per sol)
	for i := 0; i < int(types.LikeQuotaPerSol); i++ {
		target := signalAddr(100 + i)
		if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, target, types.SignalTypeLike); err != nil {
			t.Fatalf("like %d: unexpected error: %v", i, err)
		}
	}

	// the 11th Like should fail with ErrLikeQuotaExceeded
	err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, signalAddr(999), types.SignalTypeLike)
	if !errors.Is(err, types.ErrLikeQuotaExceeded) {
		t.Fatalf("expected ErrLikeQuotaExceeded, got %v", err)
	}
}

func TestDuplicateLikeInSameSol(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	sender := signalAddr(1)
	target := signalAddr(2)

	if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), signalAddr(3), sender, types.SignalTypeUse); err != nil {
		t.Fatalf("seed sender: %v", err)
	}
	if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, target, types.SignalTypeLike); err != nil {
		t.Fatalf("first like: %v", err)
	}
	err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, target, types.SignalTypeLike)
	if !errors.Is(err, types.ErrDuplicateLike) {
		t.Fatalf("expected ErrDuplicateLike, got %v", err)
	}

	// A rejected duplicate must not consume the sender's daily quota.
	for i := 1; i < int(types.LikeQuotaPerSol); i++ {
		nextTarget := signalAddr(100 + i)
		if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, nextTarget, types.SignalTypeLike); err != nil {
			t.Fatalf("distinct like %d: %v", i, err)
		}
	}
	if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, signalAddr(999), types.SignalTypeLike); !errors.Is(err, types.ErrLikeQuotaExceeded) {
		t.Fatalf("expected quota after %d distinct likes, got %v", types.LikeQuotaPerSol, err)
	}
}

func TestPopularTargetDoesNotConsumeSenderQuota(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	target := signalAddr(1)

	for i := 0; i < int(types.LikeQuotaPerSol)+1; i++ {
		sender := signalAddr(100 + i)
		if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), signalAddr(2), sender, types.SignalTypeUse); err != nil {
			t.Fatalf("seed sender %d: %v", i, err)
		}
		if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), sender, target, types.SignalTypeLike); err != nil {
			t.Fatalf("like from sender %d: %v", i, err)
		}
	}
}

func TestLikeSenderWeightFloor(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)

	// a low-contribution sender (weight 0) cannot send a Like
	err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), signalAddr(1), signalAddr(2), types.SignalTypeLike)
	if !errors.Is(err, types.ErrSenderWeightTooLow) {
		t.Fatalf("expected ErrSenderWeightTooLow, got %v", err)
	}
}

func TestSignalDecay(t *testing.T) {
	f := initFixture(t)
	sdkCtx := sdk.UnwrapSDKContext(f.ctx)
	target := signalAddr(2)

	// Use signal (weight 10) at sol 0
	if err := f.keeper.Signal(sdk.WrapSDKContext(sdkCtx), signalAddr(1), target, types.SignalTypeUse); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	w, err := f.keeper.GetSignalWeight(sdk.WrapSDKContext(sdkCtx), target)
	if err != nil {
		t.Fatalf("get weight: %v", err)
	}
	if w != types.UseWeight {
		t.Fatalf("expected weight %d, got %d", types.UseWeight, w)
	}

	// advance one Martian year (670 sol): weight halves
	later := sdkCtx.WithBlockHeight(types.SignalHalvingSol * types.BlocksPerSol)
	w2, err := f.keeper.GetSignalWeight(sdk.WrapSDKContext(later), target)
	if err != nil {
		t.Fatalf("get weight: %v", err)
	}
	if w2 != types.UseWeight/2 {
		t.Fatalf("expected weight %d after one halving, got %d", types.UseWeight/2, w2)
	}
}
