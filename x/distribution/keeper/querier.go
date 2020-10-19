package keeper

import (
	"encoding/json"
	"strings"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, k)

		case types.QueryValidatorOutstandingRewards:
			return queryValidatorOutstandingRewards(ctx, path[1:], req, k)

		case types.QueryValidatorCommission:
			return queryValidatorCommission(ctx, path[1:], req, k)

		case types.QueryValidatorSlashes:
			return queryValidatorSlashes(ctx, path[1:], req, k)

		case types.QueryDelegationRewards:
			return queryDelegationRewards(ctx, path[1:], req, k)

		case types.QueryDelegatorTotalRewards:
			return queryDelegatorTotalRewards(ctx, path[1:], req, k)

		case types.QueryDelegatorValidators:
			return queryDelegatorValidators(ctx, path[1:], req, k)

		case types.QueryWithdrawAddr:
			return queryDelegatorWithdrawAddress(ctx, path[1:], req, k)

		case types.QueryPool:
			return queryPool(ctx, path[1:], req, k)

		case types.QueryLockedRewardsState:
			return queryLockedRewardsState(ctx, path[1:], req, k)

		case types.QueryLockedRatio:
			return queryLockedRatio(ctx, path[1:], req, k)

		case types.QueryValidatorExtended:
			return queryValidatorExtended(ctx, path[1:], req, k)

		case types.QueryValidatorsExtended:
			return queryValidatorsExtended(ctx, path[1:], req, k)

		default:
			return nil, sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unknown query path: %s", path[0])
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(k.cdc, params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return res, nil
}

func queryValidatorOutstandingRewards(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorOutstandingRewardsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	rewards := k.GetValidatorOutstandingRewards(ctx, params.ValidatorAddress)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, rewards)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorCommission(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorCommissionParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	commission := k.GetValidatorAccumulatedCommission(ctx, params.ValidatorAddress)
	if commission == nil {
		commission = sdk.DecCoins{}
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, commission)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorSlashes(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorSlashesParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	events := make([]types.ValidatorSlashEvent, 0)
	k.IterateValidatorSlashEventsBetween(ctx, params.ValidatorAddress, params.StartingHeight, params.EndingHeight,
		func(height uint64, event types.ValidatorSlashEvent) (stop bool) {
			events = append(events, event)
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, events)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegationRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegationRewardsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	val := k.stakingKeeper.Validator(ctx, params.ValidatorAddress)
	if val == nil {
		return nil, sdkerrors.Wrap(types.ErrNoValidatorExists, params.ValidatorAddress.String())
	}

	del := k.stakingKeeper.Delegation(ctx, params.DelegatorAddress, params.ValidatorAddress)
	if del == nil {
		return nil, types.ErrNoDelegationExists
	}

	// get main rewards and bank accumulated rewards
	endingPeriod := k.incrementValidatorPeriod(ctx, val)
	rewards := k.calculateDelegationTotalRewards(ctx, val, del, endingPeriod)
	if rewards == nil {
		rewards = sdk.DecCoins{}
	}
	total := k.addAccumulatedBankRewards(ctx, del.GetDelegatorAddr(), rewards)

	// build response
	resp := types.NewQueryDelegationRewardsResponse(types.NewDelegationDelegatorReward(params.ValidatorAddress, rewards), total)
	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorTotalRewards(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	total := sdk.DecCoins{}
	var delRewards []types.DelegationDelegatorReward

	// iterate over delegations and calculate total bonding and LP rewards
	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			valAddr := del.GetValidatorAddr()
			val := k.stakingKeeper.Validator(ctx, valAddr)
			endingPeriod := k.incrementValidatorPeriod(ctx, val)
			delReward := k.calculateDelegationTotalRewards(ctx, val, del, endingPeriod)

			delRewards = append(delRewards, types.NewDelegationDelegatorReward(valAddr, delReward))
			total = total.Add(delReward...)
			return false
		},
	)

	// include bank accumulated rewards
	total = k.addAccumulatedBankRewards(ctx, params.DelegatorAddress, total)

	// build response
	resp := types.NewQueryDelegatorTotalRewardsResponse(delRewards, total)
	bz, err := json.Marshal(resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorValidators(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()

	var validators []sdk.ValAddress

	k.stakingKeeper.IterateDelegations(
		ctx, params.DelegatorAddress,
		func(_ int64, del exported.DelegationI) (stop bool) {
			validators = append(validators, del.GetValidatorAddr())
			return false
		},
	)

	bz, err := codec.MarshalJSONIndent(k.cdc, validators)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryDelegatorWithdrawAddress(ctx sdk.Context, _ []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorWithdrawAddrParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// cache-wrap context as to not persist state changes during querying
	ctx, _ = ctx.CacheContext()
	withdrawAddr := k.GetDelegatorWithdrawAddr(ctx, params.DelegatorAddress)

	bz, err := codec.MarshalJSONIndent(k.cdc, withdrawAddr)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryPool(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	if len(path) == 0 {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "query path: empty")
	}

	var poolSupply sdk.DecCoins
	rewardPools := k.GetRewardPools(ctx)

	switch poolName := types.RewardPoolName(path[0]); poolName {
	case types.LiquidityProvidersPoolName:
		poolSupply = rewardPools.LiquidityProvidersPool
	case types.PublicTreasuryPoolName:
		poolSupply = rewardPools.PublicTreasuryPool
	case types.FoundationPoolName:
		poolSupply = rewardPools.FoundationPool
	case types.HARPName:
		poolSupply = rewardPools.HARP
	default:
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "unknown pool name: %s", poolName)
	}

	bz, err := k.cdc.MarshalJSON(poolSupply)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryLockedRewardsState(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryLockedRewardsStateParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	lockedState, found := k.GetValidatorLockedState(ctx, params.ValidatorAddress)
	if !found {
		return nil, types.ErrNoValidatorExists
	}

	resp := types.NewQueryLockedRewardsStateResponse(lockedState)
	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryLockedRatio(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	ratio := k.LockedRatio(ctx)

	return ratio.MarshalJSON()
}

func queryValidatorExtended(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// fetch staking data
	validator, found := k.stakingKeeper.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorExists
	}
	stakingState := k.stakingKeeper.GetValidatorStakingState(ctx, params.ValidatorAddr)
	maxDelegationsRatio := k.stakingKeeper.MaxDelegationsRatio(ctx)

	// fetch distribution data
	lpRatio := k.GetValidatorLPDistrRatio(ctx)
	lockedState, found := k.GetValidatorLockedState(ctx, params.ValidatorAddr)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrInternal, "locked state not found for validator: %s", params.ValidatorAddr)
	}
	distrPower, lpPower := k.GetDistributionPower(ctx, validator.GetOperator(), validator.GetConsensusPower(), validator.LPPower(), lpRatio)

	// build response
	resp, err := types.NewValidatorResp(validator, stakingState, lockedState, maxDelegationsRatio, distrPower, lpPower)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}

func queryValidatorsExtended(ctx sdk.Context, path []string, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorsParams
	err := k.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONUnmarshal, err.Error())
	}

	// fetch common staking data
	validators := k.stakingKeeper.GetAllValidators(ctx)
	maxDelegationsRatio := k.stakingKeeper.MaxDelegationsRatio(ctx)

	// fetch common distribution data
	lpRatio := k.GetValidatorLPDistrRatio(ctx)

	// filter and paginate validators
	filteredVals := make([]staking.Validator, 0, len(validators))
	for _, val := range validators {
		if strings.EqualFold(val.GetStatus().String(), params.Status) {
			filteredVals = append(filteredVals, val)
		}
	}
	start, end := client.Paginate(len(filteredVals), params.Page, params.Limit, 50)
	if start < 0 || end < 0 {
		filteredVals = []staking.Validator{}
	} else {
		filteredVals = filteredVals[start:end]
	}

	// build response fetching additional data
	resp := make([]types.ValidatorResp, 0, len(filteredVals))
	for _, val := range filteredVals {
		// fetch
		lockedState, found := k.GetValidatorLockedState(ctx, val.OperatorAddress)
		if !found {
			return nil, sdkerrors.Wrapf(types.ErrInternal, "locked state not found for validator: %s", val.OperatorAddress)
		}
		stakingState := k.stakingKeeper.GetValidatorStakingState(ctx, val.OperatorAddress)
		distrPower, lpPower := k.GetDistributionPower(ctx, val.GetOperator(), val.GetConsensusPower(), val.LPPower(), lpRatio)

		// build
		valExtended, err := types.NewValidatorResp(val, stakingState, lockedState, maxDelegationsRatio, distrPower, lpPower)
		if err != nil {
			return nil, err
		}

		resp = append(resp, valExtended)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, err.Error())
	}

	return bz, nil
}
