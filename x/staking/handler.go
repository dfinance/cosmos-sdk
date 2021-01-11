package staking

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	msgServer := keeper.NewMsgServerImpl(k)

	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case *types.MsgCreateValidator:
			res, err := msgServer.CreateValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgEditValidator:
			res, err := msgServer.EditValidator(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgDelegate:
			res, err := msgServer.Delegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgBeginRedelegate:
			res, err := msgServer.BeginRedelegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		case *types.MsgUndelegate:
			res, err := msgServer.Undelegate(sdk.WrapSDKContext(ctx), msg)
			return sdk.WrapServiceResult(ctx, res, err)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", types.ModuleName, msg)
		}
	}
}

// These functions assume everything has been authenticated,
// now we just perform action and save

//func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.Keeper) (*sdk.Result, error) {
//	// check to see if the pubkey or sender has been registered before
//	if _, found := k.GetValidator(ctx, msg.ValidatorAddress); found {
//		return nil, ErrValidatorOwnerExists
//	}
//
//	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(msg.PubKey)); found {
//		return nil, ErrValidatorPubKeyExists
//	}
//
//	if msg.Value.Denom != k.BondDenom(ctx) {
//		return nil, ErrBadDenom
//	}
//
//	if _, err := msg.Description.EnsureLength(); err != nil {
//		return nil, err
//	}
//
//	if minValue := k.MinSelfDelegationLvl(ctx); msg.MinSelfDelegation.LT(minValue) {
//		return nil, sdkerrors.Wrapf(ErrInvalidMinSelfDelegation, "should be GTE to %s (%s)", minValue.String(), msg.MinSelfDelegation.String())
//	}
//
//	if ctx.ConsensusParams() != nil {
//		tmPubKey := tmtypes.TM2PB.PubKey(msg.PubKey)
//		if !tmstrings.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
//			return nil, sdkerrors.Wrapf(
//				ErrValidatorPubKeyTypeNotSupported,
//				"got: %s, valid: %s", tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes,
//			)
//		}
//	}
//}
