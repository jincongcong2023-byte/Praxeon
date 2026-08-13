package keeper_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"mars/x/identity/keeper"
	module "mars/x/identity/module"
	"mars/x/identity/types"
)

type fixture struct {
	ctx          context.Context
	keeper       keeper.Keeper
	addressCodec address.Codec
}

func identityAddr(n int) string {
	return sdk.AccAddress([]byte(fmt.Sprintf("address_%012d", n))).String()
}

func TestRegisterAgentLimit(t *testing.T) {
	f := initFixture(t)
	creator := sdk.AccAddress([]byte("creator_0000000000001")).String()
	if _, err := f.keeper.IssueCard(f.ctx, creator); err != nil {
		t.Fatalf("issue card: %v", err)
	}

	for i := 0; i < types.MaxAgentsPerCard; i++ {
		if err := f.keeper.RegisterAgent(f.ctx, creator, identityAddr(100+i)); err != nil {
			t.Fatalf("register agent %d: %v", i, err)
		}
	}
	if err := f.keeper.RegisterAgent(f.ctx, creator, identityAddr(999)); !errors.Is(err, types.ErrAgentLimitReached) {
		t.Fatalf("expected ErrAgentLimitReached, got %v", err)
	}
}

func TestAgentCannotBeRebound(t *testing.T) {
	f := initFixture(t)
	creatorA := sdk.AccAddress([]byte("creator_0000000000001")).String()
	creatorB := sdk.AccAddress([]byte("creator_0000000000002")).String()
	if _, err := f.keeper.IssueCard(f.ctx, creatorA); err != nil {
		t.Fatalf("issue card A: %v", err)
	}
	if _, err := f.keeper.IssueCard(f.ctx, creatorB); err != nil {
		t.Fatalf("issue card B: %v", err)
	}
	sharedAgent := identityAddr(100)
	if err := f.keeper.RegisterAgent(f.ctx, creatorA, sharedAgent); err != nil {
		t.Fatalf("register agent: %v", err)
	}
	if err := f.keeper.RegisterAgent(f.ctx, creatorB, sharedAgent); !errors.Is(err, types.ErrAgentAlreadyBound) {
		t.Fatalf("expected ErrAgentAlreadyBound, got %v", err)
	}
}

func TestRegisterAgentRejectsInvalidAddress(t *testing.T) {
	f := initFixture(t)
	creator := identityAddr(1)
	if _, err := f.keeper.IssueCard(f.ctx, creator); err != nil {
		t.Fatalf("issue card: %v", err)
	}
	if err := f.keeper.RegisterAgent(f.ctx, creator, "not-an-address"); !errors.Is(err, types.ErrInvalidAddress) {
		t.Fatalf("expected ErrInvalidAddress, got %v", err)
	}
}

func initFixture(t *testing.T) *fixture {
	t.Helper()

	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModule{})
	addressCodec := addresscodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	storeService := runtime.NewKVStoreService(storeKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("transient_test")).Ctx

	authority := authtypes.NewModuleAddress(types.GovModuleName)

	k := keeper.NewKeeper(
		storeService,
		encCfg.Codec,
		addressCodec,
		authority,
		nil,
	)

	// Initialize params
	if err := k.Params.Set(ctx, types.DefaultParams()); err != nil {
		t.Fatalf("failed to set params: %v", err)
	}

	return &fixture{
		ctx:          ctx,
		keeper:       k,
		addressCodec: addressCodec,
	}
}
