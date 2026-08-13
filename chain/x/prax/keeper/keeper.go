package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mars/x/prax/types"
)

type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec
	// Address capable of executing a MsgUpdateParams message.
	// Typically, this should be the x/gov module account.
	authority []byte

	Schema collections.Schema
	Params collections.Item[types.Params]

	bankKeeper   types.BankKeeper
	signalKeeper types.SignalKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	bankKeeper types.BankKeeper,
	signalKeeper types.SignalKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,

		bankKeeper:   bankKeeper,
		signalKeeper: signalKeeper,
		Params:       collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// MintBlock mints the block reward, applying halving and the total supply cap,
// and distributes it by Signal weight. Called from the EndBlocker once per sol.
func (k Keeper) MintBlock(ctx context.Context) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	height := sdkCtx.BlockHeight()

	// mint once per sol (civilization cycle), not per block
	if height%types.BlocksPerSol != 0 {
		return nil
	}

	// amount = InitialMintPerSol / 2^(height / HalvingBlocks)
	amount := types.InitialMintPerSol
	for h := height / types.HalvingBlocks; h > 0 && amount > 0; h-- {
		amount /= 2
	}
	if amount <= 0 {
		return nil
	}

	// respect the total supply cap
	total := k.bankKeeper.GetSupply(sdkCtx, types.PraxDenom).Amount.Int64()
	if total >= types.TotalSupplyCap {
		return nil
	}
	if total+amount > types.TotalSupplyCap {
		amount = types.TotalSupplyCap - total
	}

	// collect decayed signal weights
	type weightedAddr struct {
		addr   string
		weight uint64
	}
	var weights []weightedAddr
	totalWeight := math.NewInt(0)
	if k.signalKeeper != nil {
		if err := k.signalKeeper.IterateSignalWeights(ctx, func(addr string, w uint64) error {
			weights = append(weights, weightedAddr{addr, w})
			totalWeight = totalWeight.Add(math.NewIntFromUint64(w))
			return nil
		}); err != nil {
			return err
		}
	}

	// No verified contribution means no issuance. Unissued scheduled rewards do
	// not roll over; the total supply cap is a maximum, not a promise to mint it.
	if totalWeight.IsZero() {
		return nil
	}

	// distribute by signal weight proportion
	var minted int64
	var remainderRecipient string
	var largestWeight uint64
	for _, w := range weights {
		if w.weight > largestWeight || (w.weight == largestWeight && (remainderRecipient == "" || w.addr < remainderRecipient)) {
			largestWeight = w.weight
			remainderRecipient = w.addr
		}
		share := math.NewInt(amount).
			Mul(math.NewIntFromUint64(w.weight)).
			Quo(totalWeight).
			Int64()
		if share <= 0 {
			continue
		}
		coins := sdk.NewCoins(sdk.NewCoin(types.PraxDenom, math.NewInt(share)))
		if err := k.bankKeeper.MintCoins(sdkCtx, types.ModuleName, coins); err != nil {
			return err
		}
		addr, err := sdk.AccAddressFromBech32(w.addr)
		if err != nil {
			return err
		}
		if err := k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, addr, coins); err != nil {
			return err
		}
		minted += share
	}

	// Send integer-division dust to the largest contributor (ties resolve by
	// address) so all issued Prax remains attributable to verified Signal.
	if remainder := amount - minted; remainder > 0 {
		coins := sdk.NewCoins(sdk.NewCoin(types.PraxDenom, math.NewInt(remainder)))
		if err := k.bankKeeper.MintCoins(sdkCtx, types.ModuleName, coins); err != nil {
			return err
		}
		addr, err := sdk.AccAddressFromBech32(remainderRecipient)
		if err != nil {
			return err
		}
		return k.bankKeeper.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, addr, coins)
	}
	return nil
}
