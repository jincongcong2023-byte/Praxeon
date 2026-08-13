package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"mars/x/identity/types"
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

	Cards      collections.Map[string, string] // card number -> address
	CardByAddr collections.Map[string, string] // address -> card number
	AgentCard  collections.Map[string, string] // agent address -> owner card number

	bankKeeper types.BankKeeper
}

func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,

	bankKeeper types.BankKeeper,
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

		bankKeeper: bankKeeper,
		Params:     collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Cards:      collections.NewMap(sb, types.CardKey, "cards", collections.StringKey, collections.StringValue),
		CardByAddr: collections.NewMap(sb, types.CardByAddrKey, "card_by_addr", collections.StringKey, collections.StringValue),
		AgentCard:  collections.NewMap(sb, types.AgentCardKey, "agent_card", collections.StringKey, collections.StringValue),
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

// IssueCard issues a new identity card for the given address. One card per
// address. The card number is a random 16-digit Luhn number.
func (k Keeper) IssueCard(ctx context.Context, creator string) (string, error) {
	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return "", types.ErrInvalidAddress
	}
	if has, err := k.CardByAddr.Has(ctx, creator); err != nil {
		return "", err
	} else if has {
		return "", types.ErrCardAlreadyIssued
	}

	// try a few times to find an unused card number (collision is rare)
	for range 10 {
		num, err := types.GenerateCardNumber()
		if err != nil {
			return "", err
		}
		if has, err := k.Cards.Has(ctx, num); err != nil {
			return "", err
		} else if has {
			continue
		}
		if err := k.Cards.Set(ctx, num, creator); err != nil {
			return "", err
		}
		if err := k.CardByAddr.Set(ctx, creator, num); err != nil {
			return "", err
		}
		return num, nil
	}
	return "", fmt.Errorf("failed to generate a unique card number")
}

// RegisterAgent registers an agent address under the caller's card.
func (k Keeper) RegisterAgent(ctx context.Context, creator, agent string) error {
	if _, err := k.addressCodec.StringToBytes(creator); err != nil {
		return types.ErrInvalidAddress
	}
	if _, err := k.addressCodec.StringToBytes(agent); err != nil {
		return types.ErrInvalidAddress
	}
	card, err := k.CardByAddr.Get(ctx, creator)
	if err != nil {
		return types.ErrNoCard
	}

	if _, err := k.AgentCard.Get(ctx, agent); err == nil {
		return types.ErrAgentAlreadyBound
	} else if !errors.Is(err, collections.ErrNotFound) {
		return err
	}

	var count int
	if err := k.AgentCard.Walk(ctx, nil, func(_ string, ownerCard string) (bool, error) {
		if ownerCard == card {
			count++
		}
		return count >= types.MaxAgentsPerCard, nil
	}); err != nil {
		return err
	}
	if count >= types.MaxAgentsPerCard {
		return types.ErrAgentLimitReached
	}
	return k.AgentCard.Set(ctx, agent, card)
}

// GetOwnerCard returns the card number that owns the address, either as a
// card holder or as a registered agent. Returns ErrNoCard if unknown.
func (k Keeper) GetOwnerCard(ctx context.Context, addr string) (string, error) {
	if card, err := k.AgentCard.Get(ctx, addr); err == nil {
		return card, nil
	}
	card, err := k.CardByAddr.Get(ctx, addr)
	if err != nil {
		return "", types.ErrNoCard
	}
	return card, nil
}
