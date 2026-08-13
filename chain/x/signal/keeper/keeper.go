package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"mars/x/signal/types"
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

	SignalWeights collections.Map[string, uint64] // address -> decayed signal weight
	LastSignalSol collections.Map[string, int64]  // address -> last sol when weight was updated
	LikeCount     collections.Map[string, uint64] // address -> like count in current sol
	LikeSol       collections.Map[string, int64]  // address -> sol of the like count
	LikePairSol   collections.Map[string, int64]  // sender+target -> last sol liked

	bankKeeper     types.BankKeeper
	identityKeeper types.IdentityKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	bankKeeper types.BankKeeper,
	identityKeeper types.IdentityKeeper,
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

		bankKeeper:     bankKeeper,
		identityKeeper: identityKeeper,
		Params:         collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		SignalWeights:  collections.NewMap(sb, types.SignalWeightKey, "weights", collections.StringKey, collections.Uint64Value),
		LastSignalSol:  collections.NewMap(sb, types.LastSolKey, "last_sol", collections.StringKey, collections.Int64Value),
		LikeCount:      collections.NewMap(sb, types.LikeCountKey, "like_count", collections.StringKey, collections.Uint64Value),
		LikeSol:        collections.NewMap(sb, types.LikeSolKey, "like_sol", collections.StringKey, collections.Int64Value),
		LikePairSol:    collections.NewMap(sb, types.LikePairSolKey, "like_pair_sol", collections.StringKey, collections.Int64Value),
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

// Signal adds a Use or Like signal weight to the target address, applying
// decay (halving every Martian year), the per-sol Like quota, the sender
// weight floor (③), and same-owner anti-collusion (④).
func (k Keeper) Signal(ctx context.Context, sender, target string, signalType uint32) error {
	if _, err := k.addressCodec.StringToBytes(sender); err != nil {
		return types.ErrInvalidAddress
	}
	if _, err := k.addressCodec.StringToBytes(target); err != nil {
		return types.ErrInvalidAddress
	}

	var w uint64
	switch signalType {
	case types.SignalTypeUse:
		w = types.UseWeight
	case types.SignalTypeLike:
		w = types.LikeWeight
	default:
		return types.ErrInvalidSignalType
	}

	// ③ weight dynamic: a Like from a low-contribution sender is invalid
	if signalType == types.SignalTypeLike {
		senderWeight, _ := k.GetSignalWeight(ctx, sender)
		if senderWeight < types.LikeMinSenderWeight {
			return types.ErrSenderWeightTooLow
		}
	}

	// ④ anti-collusion: sender and target under the same card cannot signal
	if k.identityKeeper != nil {
		senderCard, err1 := k.identityKeeper.GetOwnerCard(ctx, sender)
		targetCard, err2 := k.identityKeeper.GetOwnerCard(ctx, target)
		if err1 == nil && err2 == nil && senderCard == targetCard {
			return types.ErrSelfSignal
		}
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentSol := sdkCtx.BlockHeight() / types.BlocksPerSol

	// Like quota belongs to the sender. A popular target must never consume
	// other people's endorsement capacity.
	if signalType == types.SignalTypeLike {
		if err := k.enforceLikePolicy(ctx, sender, target, currentSol); err != nil {
			return err
		}
	}

	// decay: weight halves every SignalHalvingSol
	cur, err := k.SignalWeights.Get(ctx, target)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	lastSol, err := k.LastSignalSol.Get(ctx, target)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if elapsed := currentSol - lastSol; elapsed > 0 {
		cur >>= uint64(elapsed / types.SignalHalvingSol)
	}

	if cur > ^uint64(0)-w {
		return types.ErrSignalOverflow
	}
	if err := k.SignalWeights.Set(ctx, target, cur+w); err != nil {
		return err
	}
	return k.LastSignalSol.Set(ctx, target, currentSol)
}

func (k Keeper) enforceLikePolicy(ctx context.Context, sender, target string, currentSol int64) error {
	pairKey := sender + "\x00" + target
	lastPairSol, err := k.LikePairSol.Get(ctx, pairKey)
	if err == nil && lastPairSol == currentSol {
		return types.ErrDuplicateLike
	}
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	likeSol, err := k.LikeSol.Get(ctx, sender)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	count, err := k.LikeCount.Get(ctx, sender)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	if likeSol != currentSol {
		count = 0
		if err := k.LikeSol.Set(ctx, sender, currentSol); err != nil {
			return err
		}
	}
	if count >= types.LikeQuotaPerSol {
		return types.ErrLikeQuotaExceeded
	}
	if err := k.LikeCount.Set(ctx, sender, count+1); err != nil {
		return err
	}
	return k.LikePairSol.Set(ctx, pairKey, currentSol)
}

// GetSignalWeight returns the current (decayed) signal weight for an address.
func (k Keeper) GetSignalWeight(ctx context.Context, address string) (uint64, error) {
	weight, err := k.SignalWeights.Get(ctx, address)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return 0, err
	}
	lastSol, err := k.LastSignalSol.Get(ctx, address)
	if err != nil && !errors.Is(err, collections.ErrNotFound) {
		return 0, err
	}
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentSol := sdkCtx.BlockHeight() / types.BlocksPerSol
	if elapsed := currentSol - lastSol; elapsed > 0 {
		weight >>= uint64(elapsed / types.SignalHalvingSol)
	}
	return weight, nil
}

// IterateSignalWeights walks all addresses with a positive (decayed) signal
// weight, calling fn(address, weight) for each. Used by x/prax to distribute
// the block reward by contribution.
func (k Keeper) IterateSignalWeights(ctx context.Context, fn func(addr string, weight uint64) error) error {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	currentSol := sdkCtx.BlockHeight() / types.BlocksPerSol

	return k.SignalWeights.Walk(ctx, nil, func(key string, value uint64) (bool, error) {
		lastSol, err := k.LastSignalSol.Get(ctx, key)
		if err != nil && !errors.Is(err, collections.ErrNotFound) {
			return true, err
		}
		if elapsed := currentSol - lastSol; elapsed > 0 {
			value >>= uint64(elapsed / types.SignalHalvingSol)
		}
		if value == 0 {
			return false, nil
		}
		if err := fn(key, value); err != nil {
			return true, err
		}
		return false, nil
	})
}
