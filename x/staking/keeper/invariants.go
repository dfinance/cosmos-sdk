package keeper

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RegisterInvariants registers all staking invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {

	ir.RegisterRoute(types.ModuleName, "module-accounts",
		ModuleAccountInvariants(k))
	ir.RegisterRoute(types.ModuleName, "nonnegative-power",
		NonNegativePowerInvariant(k))
	ir.RegisterRoute(types.ModuleName, "positive-delegation",
		PositiveDelegationInvariant(k))
	ir.RegisterRoute(types.ModuleName, "delegator-shares",
		DelegatorSharesInvariant(k))
}

// AllInvariants runs all invariants of the staking module.
func AllInvariants(k Keeper) sdk.Invariant {

	return func(ctx sdk.Context) (string, bool) {
		res, stop := ModuleAccountInvariants(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = NonNegativePowerInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		res, stop = PositiveDelegationInvariant(k)(ctx)
		if stop {
			return res, stop
		}

		return DelegatorSharesInvariant(k)(ctx)
	}
}

// ModuleAccountInvariants checks that the bonded and notBonded ModuleAccounts pools
// reflects the tokens actively bonded and not bonded
func ModuleAccountInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		bonded := sdk.ZeroInt()
		notBonded := sdk.ZeroInt()
		liquidity := sdk.ZeroInt()
		//
		bondedPool := k.GetBondedPool(ctx)
		notBondedPool := k.GetNotBondedPool(ctx)
		liquidityPool := k.GetLiquidityPool(ctx)
		//
		bondDenom := k.BondDenom(ctx)
		liquidityDenom := k.LPDenom(ctx)

		k.IterateValidators(ctx, func(_ int64, validator exported.ValidatorI) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetBondingTokens())
			case sdk.Unbonding, sdk.Unbonded:
				notBonded = notBonded.Add(validator.GetBondingTokens())
			default:
				panic("invalid validator status")
			}
			liquidity = liquidity.Add(validator.GetLPTokens())
			return false
		})

		k.IterateUnbondingDelegations(ctx, func(_ int64, ubd types.UnbondingDelegation) bool {
			for _, entry := range ubd.Entries {
				if entry.OpType.IsBonding() {
					notBonded = notBonded.Add(entry.Balance)
				}
				if entry.OpType.IsLiquidity() {
					liquidity = liquidity.Add(entry.Balance)
				}
			}
			return false
		})

		poolBonded := bondedPool.GetCoins().AmountOf(bondDenom)
		poolNotBonded := notBondedPool.GetCoins().AmountOf(bondDenom)
		poolLiquidity := liquidityPool.GetCoins().AmountOf(liquidityDenom)
		broken := !poolBonded.Equal(bonded) || !poolNotBonded.Equal(notBonded) || !poolLiquidity.Equal(liquidity)

		// Bonded tokens should equal sum of tokens with bonded validators
		// Not-bonded tokens should equal unbonding delegations	plus tokens on unbonded validators
		return sdk.FormatInvariant(types.ModuleName, "module account coins", fmt.Sprintf(
			"\tPool's bonded tokens: %v\n"+
				"\tsum of bonded tokens: %v\n"+
				"\tPool's not bonded tokens: %v\n"+
				"\tsum of not bonded tokens: %v\n"+
				"\tPool's liquidity tokens: %v\n"+
				"\tsum of liquidity tokens: %v\n"+
				"module accounts total (bonded + not bonded):\n"+
				"\tModule Accounts' tokens: %v\n"+
				"\tsum tokens:              %v\n",
			poolBonded, bonded,
			poolNotBonded, notBonded,
			poolLiquidity, liquidity,
			poolBonded.Add(poolNotBonded), bonded.Add(notBonded))), broken
	}
}

// NonNegativePowerInvariant checks that all stored validators have >= 0 power.
func NonNegativePowerInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var broken bool

		iterator := k.ValidatorsPowerStoreIterator(ctx)

		for ; iterator.Valid(); iterator.Next() {
			validator, found := k.GetValidator(ctx, iterator.Value())
			if !found {
				panic(fmt.Sprintf("validator record not found for address: %X\n", iterator.Value()))
			}

			powerKey := types.GetValidatorsByPowerIndexKey(validator)

			if !bytes.Equal(iterator.Key(), powerKey) {
				broken = true
				msg += fmt.Sprintf("power store invariance:\n\tvalidator.Power: %v"+
					"\n\tkey should be: %v\n\tkey in store: %v\n",
					validator.GetConsensusPower(), powerKey, iterator.Key())
			}

			if validator.Bonding.Tokens.IsNegative() {
				broken = true
				msg += fmt.Sprintf("\tnegative tokens for validator: %v\n", validator)
			}
		}
		iterator.Close()
		return sdk.FormatInvariant(types.ModuleName, "nonnegative power", fmt.Sprintf("found invalid validator powers\n%s", msg)), broken
	}
}

// PositiveDelegationInvariant checks that all stored delegations have > 0 shares.
func PositiveDelegationInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var count int

		delegations := k.GetAllDelegations(ctx)
		for _, delegation := range delegations {
			if delegation.BondingShares.IsNegative() {
				count++
				msg += fmt.Sprintf("\tdelegation with negative BondingShares: %+v\n", delegation)
			}
			if delegation.LPShares.IsNegative() {
				count++
				msg += fmt.Sprintf("\tdelegation with negative LPShares: %+v\n", delegation)
			}
			if delegation.BondingShares.IsZero() && delegation.LPShares.IsZero() {
				count++
				msg += fmt.Sprintf("\tdelegation with zero shares: %+v\n", delegation)
			}
		}
		broken := count != 0

		return sdk.FormatInvariant(types.ModuleName, "positive delegations", fmt.Sprintf(
			"%d invalid delegations found\n%s", count, msg)), broken
	}
}

// DelegatorSharesInvariant checks whether all the delegator shares which persist
// in the delegator object add up to the correct total delegator shares
// amount stored in each validator.
// Invariant checks ValidatorStakingState consistency.
func DelegatorSharesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var msg string
		var broken bool

		validators := k.GetAllValidators(ctx)
		for _, validator := range validators {
			valStakingState := k.GetValidatorStakingState(ctx, validator.OperatorAddress)

			valTotalDelBondingShares := validator.GetBondingDelegatorShares()
			valTotalDelLPShares := validator.GetLPDelegatorShares()

			totalDelBondingShares, totalDelLPShares := sdk.ZeroDec(), sdk.ZeroDec()
			delegations := k.GetValidatorDelegations(ctx, validator.GetOperator())
			for _, delegation := range delegations {
				totalDelBondingShares = totalDelBondingShares.Add(delegation.BondingShares)
				totalDelLPShares = totalDelLPShares.Add(delegation.LPShares)

				if err := valStakingState.InvariantCheck(validator, delegation); err != nil {
					broken = true
					msg += err.Error() + "\n"
				}
			}

			if !valTotalDelBondingShares.Equal(totalDelBondingShares) {
				broken = true
				msg += fmt.Sprintf("broken delegator BondingShares invariance:\n"+
					"\tvalidator.DelegatorShares: %v\n"+
					"\tsum of Delegator.Shares: %v\n", valTotalDelBondingShares, totalDelBondingShares)
			}
			if !valTotalDelLPShares.Equal(totalDelLPShares) {
				broken = true
				msg += fmt.Sprintf("broken delegator LPShares invariance:\n"+
					"\tvalidator.DelegatorShares: %v\n"+
					"\tsum of Delegator.Shares: %v\n", valTotalDelLPShares, totalDelLPShares)
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "delegator shares", msg), broken
	}
}
