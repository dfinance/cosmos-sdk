package mint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

// NewHandler returns a handler for "mint" type messages.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgSetFoundationAllocationRatio:
			return handleMsgSetFoundationAllocationRatio(ctx, k, msg)
		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized mint message type: %T", msg)
		}
	}
}

// Handle MsgSetFoundationAllocationRatio.
func handleMsgSetFoundationAllocationRatio(ctx sdk.Context, k keeper.Keeper, msg types.MsgSetFoundationAllocationRatio) (*sdk.Result, error) {
	hasPermission := false
	for _, nominee := range k.GetNominees(ctx) {
		if nominee.Equals(msg.FromAddress) {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "operation is allowed just for foundation nominee")
	}

	abpy, err := k.GetAvgBlocksPerYear(ctx)
	if err != nil {
		return nil, err
	}

	chainAge := float64(ctx.BlockHeight()) / float64(abpy)

	if chainAge > ChangeFoundationAllocationRatioTTL {
		return nil, sdkerrors.Wrapf(ErrExceededTimeLimit, "is not allowed to change after %d year", ChangeFoundationAllocationRatioTTL)
	}

	params := k.GetParams(ctx)
	params.FoundationAllocationRatio = msg.Ratio
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
		),
	)

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}
