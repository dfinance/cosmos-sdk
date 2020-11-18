package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"strconv"
)

type ParamHooks struct {
	k Keeper
}

var _ params.ParamsHooks = ParamHooks{}

func (k Keeper) ParamHooks() ParamHooks { return ParamHooks{k} }

// BeforeParamChanged Implements BeforeParamChanged
func (h ParamHooks) BeforeParamChanged(ctx sdk.Context, c params.ParamChange) (err error) {
	if c.Subspace != DefaultParamspace {
		return nil
	}

	switch c.Key {
	case string(types.KeyMaxValidators):
		err = beforeChangedMaxValidators(ctx, h.k, c)
	}

	return
}

// AfterParamChanged Implements AfterParamChanged
func (h ParamHooks) AfterParamChanged(ctx sdk.Context, c params.ParamChange) (err error) {
	if c.Subspace != DefaultParamspace {
		return nil
	}

	switch c.Key {
	case string(types.KeyMaxDelegationsRatio):
		err = afterChangedMaxDelegationsRatio(ctx, h.k)
	}

	return
}

// beforeChangedMaxValidators checks possibility for changing MaxValidators parameter
func beforeChangedMaxValidators(ctx sdk.Context, k Keeper, c params.ParamChange) error {
	vc := len(k.GetAllValidators(ctx))
	value, err := strconv.Atoi(c.Value)
	if err != nil {
		return errors.Wrap(err, "can not convert value to the int")
	}

	if vc < value {
		return errors.Wrap(types.ErrDeniedChangingParam, "can not reduce validator quantity less than current amount")
	}

	return nil
}

// afterChangedMaxDelegationsRatio reduces delegations to the validator if delegations is overflow.
func afterChangedMaxDelegationsRatio(ctx sdk.Context, k Keeper) error {
	for _, v := range k.GetAllValidators(ctx) {
		selfStake, totalStake := k.GetValidatorStakingState(ctx, v.OperatorAddress).GetSelfAndTotalStakes(v)
		if isOverflow, limit := k.HasValidatorDelegationsOverflow(ctx, selfStake, totalStake); isOverflow {
			delegatedAmount := totalStake.Sub(selfStake)
			needReduce := totalStake.Sub(limit)
			reducedAmount := needReduce.ToDec()

			delegators := k.GetValidatorDelegations(ctx, v.OperatorAddress)
			for i, d := range delegators {
				if d.DelegatorAddress.Equals(v.GetOperator()) {
					continue
				}

				reducedShares := d.BondingShares.Mul(needReduce.ToDec()).Quo(delegatedAmount.ToDec())

				if len(delegators) == i+1 {
					reducedShares = reducedAmount
				} else {
					reducedAmount = reducedAmount.Sub(reducedShares)
				}

				_, err := k.Undelegate(ctx, d.DelegatorAddress, v.OperatorAddress, types.BondingDelOpType, reducedShares, true)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
