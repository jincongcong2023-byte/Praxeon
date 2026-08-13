package keeper

import (
	"context"

	"mars/x/signal/types"
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

// SignalWeight returns the current (decayed) signal weight for an address.
func (q queryServer) SignalWeight(ctx context.Context, req *types.QuerySignalWeightRequest) (*types.QuerySignalWeightResponse, error) {
	w, err := q.k.GetSignalWeight(ctx, req.Address)
	if err != nil {
		return nil, err
	}
	return &types.QuerySignalWeightResponse{Weight: w}, nil
}
