package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mars/x/signal/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Signal sends a Use or Like signal to a target's contribution.
func (k msgServer) Signal(ctx context.Context, msg *types.MsgSignal) (*types.MsgSignalResponse, error) {
	// Use Signal is protocol-generated from a verified adoption receipt. The
	// public message endpoint only accepts Like; otherwise anyone could mint
	// reputation by claiming an unverifiable use event.
	if msg.SignalType == types.SignalTypeUse {
		return nil, types.ErrUseRequiresReceipt
	}
	if err := k.Keeper.Signal(ctx, msg.Sender, msg.Target, msg.SignalType); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent("signal_sent",
			sdk.NewAttribute("sender", msg.Sender),
			sdk.NewAttribute("target", msg.Target),
			sdk.NewAttribute("signal_type", string(rune('0'+msg.SignalType))),
		),
	)

	return &types.MsgSignalResponse{}, nil
}
