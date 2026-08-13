package keeper

import (
	"context"

	"mars/x/identity/types"
)

var _ types.QueryServer = queryServer{}

// NewQueryServerImpl returns an implementation of the QueryServer interface
// for the provided Keeper.
func NewQueryServerImpl(k Keeper) types.QueryServer {
	return queryServer{k}
}

type queryServer struct {
	k Keeper
}

// CardByAddress returns the card number for a given address.
func (q queryServer) CardByAddress(ctx context.Context, req *types.QueryCardByAddressRequest) (*types.QueryCardByAddressResponse, error) {
	num, err := q.k.CardByAddr.Get(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	return &types.QueryCardByAddressResponse{CardNumber: num}, nil
}

// CardByNumber returns the address for a given card number.
func (q queryServer) CardByNumber(ctx context.Context, req *types.QueryCardByNumberRequest) (*types.QueryCardByNumberResponse, error) {
	addr, err := q.k.Cards.Get(ctx, req.CardNumber)
	if err != nil {
		return nil, err
	}
	return &types.QueryCardByNumberResponse{Address: addr}, nil
}
