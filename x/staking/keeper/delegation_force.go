package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

type (
	FRTDRedEvent struct {
		SrcVal sdk.ValAddress
		DstVal sdk.ValAddress
		Coins  sdk.Coins
	}

	FRTDUdEvent struct {
		Val   sdk.ValAddress
		Coins sdk.Coins
	}
)

// ForceRemoveTypedDelegations stops all active undelegations and redelegations for the specified delegator and delegation type.
// All the necessary events are emitted.
func (k Keeper) ForceRemoveTypedDelegations(ctx sdk.Context, delAddr sdk.AccAddress, delOpType types.DelegationOpType) error {
	// get delOpType denom
	var udSenderName, denom string
	switch delOpType {
	case types.BondingDelOpType:
		udSenderName, denom = types.NotBondedPoolName, k.BondDenom(ctx)
	case types.LiquidityDelOpType:
		udSenderName, denom = types.LiquidityPoolName, k.LPDenom(ctx)
	default:
		return fmt.Errorf("unknown delOpType: %w", delOpType.Validate())
	}

	// force stop all active redelegations
	redCompleteEvents := make([]*FRTDRedEvent, 0)
	err := k.handleTypedActiveRedelegations(ctx, delAddr, delOpType, func(srcValAddr, dstValAddr sdk.ValAddress, redEntry types.RedelegationEntry) (remove bool, retErr error) {
		// update an existing RED event / update the existing one
		redCoin := sdk.NewCoin(denom, redEntry.InitialBalance)
		var existingEvent *FRTDRedEvent
		for _, event := range redCompleteEvents {
			if event.SrcVal.Equals(srcValAddr) && event.DstVal.Equals(dstValAddr) {
				existingEvent = event
				break
			}
		}
		if existingEvent != nil {
			existingEvent.Coins = existingEvent.Coins.Add(redCoin)
		} else {
			redCompleteEvents = append(redCompleteEvents, &FRTDRedEvent{SrcVal: srcValAddr, DstVal: dstValAddr, Coins: sdk.NewCoins(redCoin)})
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("handling delAddr %s active redelegations of type %q: %w", delAddr, delOpType, err)
	}

	// undelegate all delegations
	for _, del := range k.GetAllDelegatorDelegations(ctx, delAddr) {
		var err error
		if delOpType == types.BondingDelOpType && del.BondingShares.IsPositive() {
			_, err = k.Undelegate(ctx, delAddr, del.ValidatorAddress, delOpType, del.BondingShares, true)
		}
		if delOpType == types.LiquidityDelOpType && del.LPShares.IsPositive() {
			_, err = k.Undelegate(ctx, delAddr, del.ValidatorAddress, delOpType, del.LPShares, true)
		}
		if err != nil {
			return fmt.Errorf("unbonding delAddr %s delegation of type %q for valAddr %s: %w", delAddr, delOpType, del.ValidatorAddress, err)
		}
	}

	// force stop all active undelegations
	udCompleteEvents := make([]*FRTDUdEvent, 0)
	err = k.handleTypedActiveUndelegations(ctx, delAddr, delOpType, func(valAddr sdk.ValAddress, udEntry types.UnbondingDelegationEntry) (remove bool, retErr error) {
		// update an existing UD event / update the existing one
		udCoin := sdk.NewCoin(denom, udEntry.Balance)
		var existingEvent *FRTDUdEvent
		for _, event := range udCompleteEvents {
			if event.Val.Equals(valAddr) {
				existingEvent = event
				break
			}
		}
		if existingEvent != nil {
			existingEvent.Coins = existingEvent.Coins.Add(udCoin)
		} else {
			udCompleteEvents = append(udCompleteEvents, &FRTDUdEvent{Val: valAddr, Coins: sdk.NewCoins(udCoin)})
		}

		// transfer UD coins to the specific delegator
		if udCoin.Amount.IsPositive() {
			err := k.supplyKeeper.UndelegateCoinsFromModuleToAccount(ctx, udSenderName, delAddr, sdk.NewCoins(udCoin))
			if err != nil {
				return false, fmt.Errorf("coins transfer from %s: %w", delAddr, err)
			}
		}

		return true, nil
	})

	// emit events
	for _, event := range redCompleteEvents {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteRedelegation,
				sdk.NewAttribute(sdk.AttributeKeyAmount, event.Coins.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
				sdk.NewAttribute(types.AttributeKeySrcValidator, event.SrcVal.String()),
				sdk.NewAttribute(types.AttributeKeyDstValidator, event.DstVal.String()),
			),
		)
	}
	for _, event := range udCompleteEvents {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeCompleteUnbonding,
				sdk.NewAttribute(sdk.AttributeKeyAmount, event.Coins.String()),
				sdk.NewAttribute(types.AttributeKeyDelegator, delAddr.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, event.Val.String()),
			),
		)
	}

	return nil
}

// handleTypedActiveRedelegations iterates over redelegation queue (active redelegations), filters out entries for
// the specified delegator and delegation type and executes provided handler.
// Entry is removed if handlers returns a corresponding flag.
func (k Keeper) handleTypedActiveRedelegations(
	ctx sdk.Context, delAddr sdk.AccAddress, delOpType types.DelegationOpType,
	handler func(srcValAddr, dstValAddr sdk.ValAddress, redEntry types.RedelegationEntry) (remove bool, retErr error),
) error {

	store := ctx.KVStore(k.storeKey)

	// we fake the endTime for RedelegationQueue to get all the active redelegations
	iterator := k.RedelegationQueueIterator(ctx, ctx.BlockTime().Add(types.MaxUnbondingTime))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		tripletsRead := make([]types.DVVTriplet, 0)
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &tripletsRead)

		tripletsToSet := make([]types.DVVTriplet, 0, len(tripletsRead))
		for _, triplet := range tripletsRead {
			// filter out redelegations for the specific delegator
			if !triplet.DelegatorAddress.Equals(delAddr) {
				tripletsToSet = append(tripletsToSet, triplet)
				continue
			}

			// get redelegation entries
			red, found := k.GetRedelegation(ctx, delAddr, triplet.ValidatorSrcAddress, triplet.ValidatorDstAddress)
			if !found {
				return fmt.Errorf("get redelegation for an existing triplet of srcVal/dstVal (%s/%s)", triplet.ValidatorSrcAddress, triplet.ValidatorDstAddress)
			}

			// handle redelegation entries
			rdEntriesToUpdate := make([]types.RedelegationEntry, 0, len(red.Entries))
			for i, redEntry := range red.Entries {
				// filter out entries by delOpType
				if redEntry.OpType != delOpType {
					rdEntriesToUpdate = append(rdEntriesToUpdate, redEntry)
					continue
				}

				removeEntry, err := handler(red.ValidatorSrcAddress, red.ValidatorDstAddress, redEntry)
				if err != nil {
					return fmt.Errorf("handling redelegation entry [%d] for triplet of srcVal/dstVal (%s/%s): %w", i, red.ValidatorSrcAddress, red.ValidatorDstAddress, err)
				}
				if !removeEntry {
					rdEntriesToUpdate = append(rdEntriesToUpdate, redEntry)
				}
			}

			// remove / update redelegation
			if len(rdEntriesToUpdate) == 0 {
				k.RemoveRedelegation(ctx, red)
			} else {
				red.Entries = rdEntriesToUpdate
				k.SetRedelegation(ctx, red)
				tripletsToSet = append(tripletsToSet, triplet)
			}
		}

		// remove / update triplets
		if len(tripletsToSet) != len(tripletsRead) {
			if len(tripletsToSet) == 0 {
				store.Delete(iterator.Key())
			} else {
				store.Set(iterator.Key(), k.cdc.MustMarshalBinaryLengthPrefixed(tripletsToSet))
			}
		}
	}

	return nil
}

// handleTypedActiveUndelegations iterates over unbonding queue (active undelegations), filters out entries for
// the specified delegator and delegation type and executes provided handler.
// Entry is removed if handlers returns a corresponding flag.
func (k Keeper) handleTypedActiveUndelegations(
	ctx sdk.Context, delAddr sdk.AccAddress, delOpType types.DelegationOpType,
	handler func(valAddr sdk.ValAddress, udEntry types.UnbondingDelegationEntry) (remove bool, retErr error),
) error {

	store := ctx.KVStore(k.storeKey)

	// we fake the endTime for UBQueue to get all active undelegations
	iterator := k.UBDQueueIterator(ctx, ctx.BlockTime().Add(types.MaxUnbondingTime))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		pairsRead := make([]types.DVPair, 0)
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &pairsRead)

		pairsToSet := make([]types.DVPair, 0, len(pairsRead))
		for _, pair := range pairsRead {
			// filter out undelegations for the specific delegator
			if !pair.DelegatorAddress.Equals(delAddr) {
				pairsToSet = append(pairsToSet, pair)
				continue
			}

			// get undelegation entries
			ud, found := k.GetUnbondingDelegation(ctx, pair.DelegatorAddress, pair.ValidatorAddress)
			if !found {
				return fmt.Errorf("get undelegation for an existing pair of val (%s)", pair.ValidatorAddress)
			}

			// handle undelegation entries
			udEntriesToUpdate := make([]types.UnbondingDelegationEntry, 0, len(ud.Entries))
			for i, udEntry := range ud.Entries {
				// filter out entries by delOpType
				if udEntry.OpType != delOpType {
					udEntriesToUpdate = append(udEntriesToUpdate, udEntry)
					continue
				}

				removeEntry, err := handler(ud.ValidatorAddress, udEntry)
				if err != nil {
					return fmt.Errorf("handling undelegation entry [%d] for pair of val (%s): %w", i, ud.ValidatorAddress, err)
				}
				if !removeEntry {
					udEntriesToUpdate = append(udEntriesToUpdate, udEntry)
				}
			}

			// remove / update undelegation
			if len(udEntriesToUpdate) == 0 {
				k.RemoveUnbondingDelegation(ctx, ud)
			} else {
				ud.Entries = udEntriesToUpdate
				k.SetUnbondingDelegation(ctx, ud)
				pairsToSet = append(pairsToSet, pair)
			}
		}

		// remove / update triplets
		if len(pairsToSet) != len(pairsRead) {
			if len(pairsToSet) == 0 {
				store.Delete(iterator.Key())
			} else {
				store.Set(iterator.Key(), k.cdc.MustMarshalBinaryLengthPrefixed(pairsToSet))
			}
		}
	}

	return nil
}
