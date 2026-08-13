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

// mockBankKeeper is a minimal in-memory implementation of types.BankKeeper.
type mockBankKeeper struct {
	supply sdk.Coin
}

func (m *mockBankKeeper) SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins {
	return nil
}

func (m *mockBankKeeper) MintCoins(_ context.Context, _ string, amt sdk.Coins) error {
	if len(amt) == 0 {
		return nil
	}
	m.supply = m.supply.Add(amt[0])
	return nil
}

func (m *mockBankKeeper) GetSupply(_ context.Context, denom string) sdk.Coin {
	if m.supply.Denom == denom {
		return m.supply
	}
	return sdk.NewCoin(denom, math.NewInt(0))
}

func (m *mockBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

// mintFixture builds a keeper wired to a mock bank keeper.
func mintFixture(t *testing.T, signal types.SignalKeeper) (*keeper.Keeper, *mockBankKeeper, sdk.Context) {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)
	bank := &mockBankKeeper{supply: sdk.NewCoin(types.PraxDenom, math.NewInt(0))}

	k := keeper.NewKeeper(storeService, encCfg.Codec, addressCodec, authority, bank, signal)
	return &k, bank, ctx
}

func TestNoSignalMeansNoMint(t *testing.T) {
	k, bank, ctx := mintFixture(t, nil)

	// height 1 (not a sol boundary): no mint
	if err := k.MintBlock(sdk.WrapSDKContext(ctx.WithBlockHeight(1))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bank.supply.Amount.Int64() != 0 {
		t.Fatalf("expected no mint at height 1, got %d", bank.supply.Amount.Int64())
	}

	// Even at a sol boundary, no verified Signal means no issuance.
	if err := k.MintBlock(sdk.WrapSDKContext(ctx.WithBlockHeight(types.BlocksPerSol))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bank.supply.Amount.Int64() != 0 {
		t.Fatalf("expected no mint without signal, got %d", bank.supply.Amount.Int64())
	}
}

func TestMintHalving(t *testing.T) {
	addr := sdk.AccAddress([]byte("address_00000000001")).String()
	signal := &mockSignalKeeper{weights: map[string]uint64{addr: 1}}
	k, bank, ctx := mintFixture(t, signal)

	// after one Martian year (HalvingBlocks, a sol boundary): mint is halved
	height := types.HalvingBlocks // = 670 * BlocksPerSol
	if err := k.MintBlock(sdk.WrapSDKContext(ctx.WithBlockHeight(height))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	half := types.InitialMintPerSol / 2
	if bank.supply.Amount.Int64() != half {
		t.Fatalf("expected mint %d after one halving, got %d", half, bank.supply.Amount.Int64())
	}
}
