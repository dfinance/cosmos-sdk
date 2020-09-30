package staking

import (
	"time"

	tmstrings "github.com/tendermint/tendermint/libs/strings"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case types.MsgCreateValidator:
			return handleMsgCreateValidator(ctx, msg, k)

		case types.MsgEditValidator:
			return handleMsgEditValidator(ctx, msg, k)

		case types.MsgDelegate:
			return handleMsgDelegate(ctx, msg, k)

		case types.MsgBeginRedelegate:
			return handleMsgBeginRedelegate(ctx, msg, k)

		case types.MsgUndelegate:
			return handleMsgUndelegate(ctx, msg, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized %s message type: %T", ModuleName, msg)
		}
	}
}

// These functions assume everything has been authenticated,
// now we just perform action and save

func handleMsgCreateValidator(ctx sdk.Context, msg types.MsgCreateValidator, k keeper.Keeper) (*sdk.Result, error) {
	// check if staking ops are denied
	if k.IsAccountBanned(ctx, msg.DelegatorAddress) {
		return nil, ErrDeniedStakingOps
	}

	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetValidator(ctx, msg.ValidatorAddress); found {
		return nil, ErrValidatorOwnerExists
	}

	if _, found := k.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(msg.PubKey)); found {
		return nil, ErrValidatorPubKeyExists
	}

	if msg.Value.Denom != k.BondDenom(ctx) {
		return nil, ErrBadDenom
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	if minValue := k.MinSelfDelegationLvl(ctx); msg.MinSelfDelegation.LT(minValue) {
		return nil, sdkerrors.Wrapf(ErrInvalidMinSelfDelegation, "should be GTE to %s (%s)", minValue.String(), msg.MinSelfDelegation.String())
	}

	if ctx.ConsensusParams() != nil {
		tmPubKey := tmtypes.TM2PB.PubKey(msg.PubKey)
		if !tmstrings.StringInSlice(tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes) {
			return nil, sdkerrors.Wrapf(
				ErrValidatorPubKeyTypeNotSupported,
				"got: %s, valid: %s", tmPubKey.Type, ctx.ConsensusParams().Validator.PubKeyTypes,
			)
		}
	}

	validator := NewValidator(msg.ValidatorAddress, msg.PubKey, msg.Description)
	commission := NewCommissionWithTime(
		msg.Commission.Rate, msg.Commission.MaxRate,
		msg.Commission.MaxChangeRate, ctx.BlockHeader().Time,
	)
	validator, err := validator.SetInitialCommission(commission)
	if err != nil {
		return nil, err
	}

	validator.MinSelfDelegation = msg.MinSelfDelegation

	k.SetValidator(ctx, validator)
	k.SetValidatorByConsAddr(ctx, validator)
	k.SetNewValidatorByPowerIndex(ctx, validator)

	// call the after-creation hook
	k.AfterValidatorCreated(ctx, validator.OperatorAddress)

	// move coins from the msg.Address account to a (self-delegation) delegator account
	// the validator account and global shares are updated within here
	// NOTE source will always be from a wallet which are unbonded
	_, err = k.Delegate(ctx, msg.DelegatorAddress, types.BondingDelOpType, msg.Value.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCreateValidator,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyCoin, msg.Value.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgEditValidator(ctx sdk.Context, msg types.MsgEditValidator, k keeper.Keeper) (*sdk.Result, error) {
	// check if staking ops are denied
	if k.IsAccountBanned(ctx, sdk.AccAddress(msg.ValidatorAddress)) {
		return nil, ErrDeniedStakingOps
	}

	// validator must already be registered
	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return nil, ErrNoValidatorFound
	}

	// replace all editable fields (clients should autofill existing values)
	description, err := validator.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	validator.Description = description

	if msg.CommissionRate != nil {
		commission, err := k.UpdateValidatorCommission(ctx, validator, *msg.CommissionRate)
		if err != nil {
			return nil, err
		}

		// call the before-modification hook since we're about to update the commission
		k.BeforeValidatorModified(ctx, msg.ValidatorAddress)

		validator.Commission = commission
	}

	if msg.MinSelfDelegation != nil {
		if !msg.MinSelfDelegation.GT(validator.MinSelfDelegation) {
			return nil, ErrMinSelfDelegationDecreased
		}
		if msg.MinSelfDelegation.GT(validator.GetBondedTokens()) {
			return nil, ErrSelfDelegationBelowMinimum
		}

		validator.MinSelfDelegation = *msg.MinSelfDelegation
	}

	k.SetValidator(ctx, validator)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeEditValidator,
			sdk.NewAttribute(types.AttributeKeyCommissionRate, validator.Commission.String()),
			sdk.NewAttribute(types.AttributeKeyMinSelfDelegation, validator.MinSelfDelegation.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.ValidatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgDelegate(ctx sdk.Context, msg types.MsgDelegate, k keeper.Keeper) (*sdk.Result, error) {
	// check if staking ops are denied
	if k.IsAccountBanned(ctx, msg.DelegatorAddress) {
		return nil, ErrDeniedStakingOps
	}

	validator, found := k.GetValidator(ctx, msg.ValidatorAddress)
	if !found {
		return nil, ErrNoValidatorFound
	}

	var delOpType types.DelegationOpType
	switch msg.Amount.Denom {
	case k.BondDenom(ctx):
		delOpType = types.BondingDelOpType
	case k.LPDenom(ctx):
		delOpType = types.LiquidityDelOpType
	default:
		return nil, ErrBadDenom
	}

	// NOTE: source funds are always unbonded
	_, err := k.Delegate(ctx, msg.DelegatorAddress, delOpType, msg.Amount.Amount, sdk.Unbonded, validator, true)
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDelegate,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyCoin, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Events: ctx.EventManager().Events()}, nil
}

func handleMsgUndelegate(ctx sdk.Context, msg types.MsgUndelegate, k keeper.Keeper) (*sdk.Result, error) {
	// check if staking ops are denied
	if k.IsAccountBanned(ctx, msg.DelegatorAddress) {
		return nil, ErrDeniedStakingOps
	}

	var delOpType types.DelegationOpType
	switch msg.Amount.Denom {
	case k.BondDenom(ctx):
		delOpType = types.BondingDelOpType
	case k.LPDenom(ctx):
		delOpType = types.LiquidityDelOpType
	default:
		return nil, ErrBadDenom
	}

	shares, err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddress, msg.ValidatorAddress, delOpType, msg.Amount.Amount)
	if err != nil {
		return nil, err
	}

	completionTime, err := k.Undelegate(ctx, msg.DelegatorAddress, msg.ValidatorAddress, delOpType, shares, false)
	if err != nil {
		return nil, err
	}

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(completionTime)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeUnbond,
			sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyCoin, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().Events()}, nil
}

func handleMsgBeginRedelegate(ctx sdk.Context, msg types.MsgBeginRedelegate, k keeper.Keeper) (*sdk.Result, error) {
	// check if staking ops are denied
	if k.IsAccountBanned(ctx, msg.DelegatorAddress) {
		return nil, ErrDeniedStakingOps
	}

	var delOpType types.DelegationOpType
	switch msg.Amount.Denom {
	case k.BondDenom(ctx):
		delOpType = types.BondingDelOpType
	case k.LPDenom(ctx):
		delOpType = types.LiquidityDelOpType
	default:
		return nil, ErrBadDenom
	}

	shares, err := k.ValidateUnbondAmount(ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, delOpType, msg.Amount.Amount)
	if err != nil {
		return nil, err
	}

	completionTime, err := k.BeginRedelegation(ctx, msg.DelegatorAddress, msg.ValidatorSrcAddress, msg.ValidatorDstAddress, delOpType, shares)
	if err != nil {
		return nil, err
	}

	completionTimeBz := types.ModuleCdc.MustMarshalBinaryLengthPrefixed(completionTime)
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRedelegate,
			sdk.NewAttribute(types.AttributeKeySrcValidator, msg.ValidatorSrcAddress.String()),
			sdk.NewAttribute(types.AttributeKeyDstValidator, msg.ValidatorDstAddress.String()),
			sdk.NewAttribute(sdk.AttributeKeyCoin, msg.Amount.String()),
			sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress.String()),
		),
	})

	return &sdk.Result{Data: completionTimeBz, Events: ctx.EventManager().Events()}, nil
}
