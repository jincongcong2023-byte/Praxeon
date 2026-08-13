package keeper

import (
	"context"

	"mars/x/mars/types"

	errorsmod "cosmossdk.io/errors"
)

func (k msgServer) IssueCard(ctx context.Context, msg *types.MsgIssueCard) (*types.MsgIssueCardResponse, error) {
	if _, err := k.addressCodec.StringToBytes(msg.Creator); err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// TODO: Handle the message

	return &types.MsgIssueCardResponse{}, nil
}
