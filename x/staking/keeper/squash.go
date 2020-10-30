package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	// Operations order:
	//   1: paramsOps
	//   2: addValidatorOps
	//   3: main squash ops (including jailWhitelist handling)
	SquashOptions struct {
		// Slice of valAddrs not to be jailed (no jailing if empty)
		jailWhitelist []sdk.ValAddress
		// Params modification operation
		paramsOps paramsOperation
		// New validator operations
		addValidatorOps []addValidatorOperation
	}

	paramsOperation struct {
		// Modify bonding denom (empty - no modification)
		BondingDenom string
	}

	addValidatorOperation struct {
		// Operator account address
		OperatorAddress sdk.AccAddress
		// Moniker ID
		Moniker string
		// Validator public key
		PubKey crypto.PubKey
		// Self-delegation amount
		SelfDelegationAmount sdk.Int
	}
)

func (opts *SquashOptions) SetParamsOp(bondingDenomRaw string) error {
	op := paramsOperation{}
	if bondingDenomRaw != "" {
		if err := sdk.ValidateDenom(bondingDenomRaw); err != nil {
			return fmt.Errorf("bondingDenom (%s): invalid: %w", bondingDenomRaw, err)
		}
		op.BondingDenom = bondingDenomRaw
	}
	opts.paramsOps = op

	return nil
}

func (opts *SquashOptions) SetAddValidatorOp(operatorAddrRaw, moniker, pubKeyRaw, selfDelegationAmountRaw string) error {
	op := addValidatorOperation{}

	operatorAddr, err := sdk.AccAddressFromBech32(operatorAddrRaw)
	if err != nil {
		return fmt.Errorf("operatorAddr (%s): invalid AccAddress: %w", operatorAddrRaw, err)
	}
	op.OperatorAddress = operatorAddr

	if moniker == "" {
		return fmt.Errorf("moniker: empty")
	}
	op.Moniker = moniker

	pubKey, err := sdk.GetPubKeyFromBech32(sdk.Bech32PubKeyTypeConsPub, pubKeyRaw)
	if err != nil {
		return fmt.Errorf("pubKey (%s): invalid ConsPubKey: %w", pubKeyRaw, err)
	}
	op.PubKey = pubKey

	selfDelegationAmount, ok := sdk.NewIntFromString(selfDelegationAmountRaw)
	if !ok {
		return fmt.Errorf("selfDelegationAmount (%s): invalid sdk.Int", selfDelegationAmountRaw)
	}
	op.SelfDelegationAmount = selfDelegationAmount

	opts.addValidatorOps = append(opts.addValidatorOps, op)

	return nil
}

func (opts *SquashOptions) SetJailWhitelistSquashOption(jailWhitelist []string) error {
	opts.jailWhitelist = make([]sdk.ValAddress, 0, len(jailWhitelist))
	for i, addrRaw := range jailWhitelist {
		valAddr, err := sdk.ValAddressFromBech32(addrRaw)
		if err != nil {
			return fmt.Errorf("JailWhitelist[%d] (%s): invalid ValAddress: %w", i, addrRaw, err)
		}
		opts.jailWhitelist = append(opts.jailWhitelist, valAddr)
	}

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		jailWhitelist:   nil,
		paramsOps:       paramsOperation{},
		addValidatorOps: nil,
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// paramsOps
	{
		if opts.paramsOps.BondingDenom != "" {
			params := k.GetParams(ctx)
			params.BondDenom = opts.paramsOps.BondingDenom
			k.SetParams(ctx, params)
		}
	}

	// addValidatorOps
	for i, valOp := range opts.addValidatorOps {
		description := types.NewDescription(valOp.Moniker, "", "", "", "")
		validator := types.NewValidator(sdk.ValAddress(valOp.OperatorAddress), valOp.PubKey, description)
		commission := types.NewCommissionWithTime(
			sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(5, 1),
			sdk.NewDecWithPrec(1, 2), ctx.BlockHeader().Time,
		)
		validator, err := validator.SetInitialCommission(commission)
		if err != nil {
			return fmt.Errorf("addValidatorOps[%d]: SetInitialCommission: %w", i, err)
		}
		validator.MinSelfDelegation = valOp.SelfDelegationAmount

		k.SetValidator(ctx, validator)
		k.SetValidatorByConsAddr(ctx, validator)
		k.SetNewValidatorByPowerIndex(ctx, validator)
		k.AfterValidatorCreated(ctx, validator.OperatorAddress)

		_, err = k.Delegate(
			ctx, valOp.OperatorAddress,
			types.BondingDelOpType, valOp.SelfDelegationAmount,
			sdk.Unbonded, validator, true,
		)
		if err != nil {
			return fmt.Errorf("addValidatorOps[%d]: Delegate: %w", i, err)
		}
	}

	// main squash operations
	{
		// check if whitelist should be applied
		applyWhiteList := false
		if len(opts.jailWhitelist) > 0 {
			applyWhiteList = true
		}
		whiteListMap := make(map[string]bool)
		for _, addr := range opts.jailWhitelist {
			whiteListMap[addr.String()] = true
		}

		// reset banned accounts height
		k.IterateBannedAccounts(ctx, func(accAddr sdk.AccAddress, _ int64) (stop bool) {
			k.BanAccount(ctx, accAddr, 0)
			return false
		})

		// reset redelegations  creation height
		k.IterateRedelegations(ctx, func(_ int64, red types.Redelegation) (stop bool) {
			for i := range red.Entries {
				red.Entries[i].CreationHeight = 0
			}
			k.SetRedelegation(ctx, red)
			return false
		})

		// reset unbonding delegations creation height
		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) (stop bool) {
			for i := range ubd.Entries {
				ubd.Entries[i].CreationHeight = 0
			}
			k.SetUnbondingDelegation(ctx, ubd)
			return false
		})

		// reset validators bond height and scheduled unbond height
		// jail if jailing is enabled and validator is not in the whitelist
		for _, val := range k.GetAllValidators(ctx) {
			val.UnbondingHeight = 0
			val.ScheduledUnbondHeight = 0
			k.SetValidator(ctx, val)

			if applyWhiteList && !whiteListMap[val.OperatorAddress.String()] {
				if !val.Jailed {
					k.jailValidator(ctx, val)
				}
			}
		}
	}

	_ = k.ApplyAndReturnValidatorSetUpdates(ctx)

	return nil
}
