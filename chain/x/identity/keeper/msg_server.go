package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"mars/x/identity/types"
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

// IssueCard issues a new identity card for the message creator.
func (k msgServer) IssueCard(ctx context.Context, msg *types.MsgIssueCard) (*types.MsgIssueCardResponse, error) {
	num, err := k.Keeper.IssueCard(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	sdkCtx.EventManager().EmitEvent(
		sdk.NewEvent("identity_card_issued",
			sdk.NewAttribute("creator", msg.Creator),
			sdk.NewAttribute("card_number", num),
		),
	)

	return &types.MsgIssueCardResponse{CardNumber: num}, nil
}

// RegisterAgent registers an agent address under the caller's card.
func (k msgServer) RegisterAgent(ctx context.Context, msg *types.MsgRegisterAgent) (*types.MsgRegisterAgentResponse, error) {
	if err := k.Keeper.RegisterAgent(ctx, msg.Creator, msg.Agent); err != nil {
		return nil, err
	}
	return &types.MsgRegisterAgentResponse{}, nil
}
