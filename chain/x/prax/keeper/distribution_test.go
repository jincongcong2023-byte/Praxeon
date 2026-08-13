package keeper_test

import (
	"context"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mars/x/prax/keeper"
	module "mars/x/prax/module"
	"mars/x/prax/types"
)

// mockSignalKeeper implements types.SignalKeeper with a fixed weight map.
type mockSignalKeeper struct {
	weights map[string]uint64
}

func (m *mockSignalKeeper) IterateSignalWeights(_ context.Context, fn func(addr string, weight uint64) error) error {
	for addr, w := range m.weights {
		if err := fn(addr, w); err != nil {
			return err
		}
	}
	return nil
}

// distributedBankKeeper records amounts sent to accounts.
type distributedBankKeeper struct {
	supply sdk.Coin
	sent   map[string]int64
}

func (m *distributedBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins {
	return nil
}

func (m *distributedBankKeeper) MintCoins(_ context.Context, _ string, amt sdk.Coins) error {
	if len(amt) == 0 {
		return nil
	}
	m.supply = m.supply.Add(amt[0])
	return nil
}

func (m *distributedBankKeeper) GetSupply(_ context.Context, denom string) sdk.Coin {
	if m.supply.Denom == denom {
		return m.supply
	}
	return sdk.NewCoin(denom, math.NewInt(0))
}

func (m *distributedBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, addr sdk.AccAddress, amt sdk.Coins) error {
	m.sent[addr.String()] += amt.AmountOf(types.PraxDenom).Int64()
	return nil
}

func TestMintDistribution(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	// two contributors: a1 weight 10, a2 weight 30 (total 40)
	a1 := sdk.AccAddress([]byte("address_00000000001")).String()
	a2 := sdk.AccAddress([]byte("address_00000000002")).String()
	sig := &mockSignalKeeper{weights: map[string]uint64{a1: 10, a2: 30}}

	bank := &distributedBankKeeper{
		supply: sdk.NewCoin(types.PraxDenom, math.NewInt(0)),
		sent:   map[string]int64{},
	}

	k := keeper.NewKeeper(storeService, encCfg.Codec, addressCodec, authtypes.NewModuleAddress(types.GovModuleName), bank, sig)

	// mint at the first sol boundary
	if err := k.MintBlock(sdk.WrapSDKContext(ctx.WithBlockHeight(types.BlocksPerSol))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// a1 should get 75000 * 10/40 = 18750
	// a2 should get 75000 * 30/40 = 56250
	want1 := types.InitialMintPerSol * 10 / 40
	want2 := types.InitialMintPerSol * 30 / 40
	if got := bank.sent[a1]; got != want1 {
		t.Fatalf("a1: expected %d, got %d", want1, got)
	}
	if got := bank.sent[a2]; got != want2 {
		t.Fatalf("a2: expected %d, got %d", want2, got)
	}
}

func TestMintRemainderGoesToContributor(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	a1 := sdk.AccAddress([]byte("address_00000000001")).String()
	a2 := sdk.AccAddress([]byte("address_00000000002")).String()
	sig := &mockSignalKeeper{weights: map[string]uint64{a1: 1, a2: 6}}
	bank := &distributedBankKeeper{
		supply: sdk.NewCoin(types.PraxDenom, math.NewInt(0)),
		sent:   map[string]int64{},
	}
	k := keeper.NewKeeper(storeService, encCfg.Codec, addressCodec, authtypes.NewModuleAddress(types.GovModuleName), bank, sig)

	if err := k.MintBlock(sdk.WrapSDKContext(ctx.WithBlockHeight(types.BlocksPerSol))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := bank.sent[a1] + bank.sent[a2]; got != types.InitialMintPerSol {
		t.Fatalf("expected all %d issued Prax to reach contributors, got %d", types.InitialMintPerSol, got)
	}
	wantLargest := types.InitialMintPerSol - types.InitialMintPerSol/7
	if got := bank.sent[a2]; got != wantLargest {
		t.Fatalf("largest contributor: expected %d including remainder, got %d", wantLargest, got)
	}
}
