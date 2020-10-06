package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
)

// TODO: Break into several smaller functions for clarity

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (keeper Keeper) Tally(ctx sdk.Context, proposal types.Proposal) (passes bool, burnDeposits bool, tallyResults types.TallyResult) {
	results := make(map[types.VoteOption]sdk.Dec)
	results[types.OptionYes] = sdk.ZeroDec()
	results[types.OptionAbstain] = sdk.ZeroDec()
	results[types.OptionNo] = sdk.ZeroDec()
	results[types.OptionNoWithVeto] = sdk.ZeroDec()

	// get LP vs bonding vote ratio
	voteLPRatio := keeper.sk.LPDistrRatio(ctx)

	totalVotingPower := sdk.ZeroDec()
	totalBondedLPTokens := sdk.ZeroInt()
	currValidators := make(map[string]types.ValidatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator exported.ValidatorI) (stop bool) {
		currValidators[validator.GetOperator().String()] = types.NewValidatorGovInfo(
			validator.GetOperator(),
			validator.GetBondedTokens(), validator.GetLPTokens(),
			validator.GetBondingDelegatorShares(), sdk.ZeroDec(),
			validator.GetLPDelegatorShares(), sdk.ZeroDec(),
			types.OptionEmpty,
		)
		totalBondedLPTokens = totalBondedLPTokens.Add(validator.GetLPTokens())

		return false
	})

	keeper.IterateVotes(ctx, proposal.ProposalID, func(vote types.Vote) bool {
		// if validator, just record it in the map
		valAddrStr := sdk.ValAddress(vote.Voter).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Option
			currValidators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		keeper.sk.IterateDelegations(ctx, vote.Voter, func(index int64, delegation exported.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr().String()

			if val, ok := currValidators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will deduct voter's voting power from validator.

				// add deduction shares
				val.DelegatorBondingDeductions = val.DelegatorBondingDeductions.Add(delegation.GetBondingShares())
				val.DelegatorLPDeductions = val.DelegatorLPDeductions.Add(delegation.GetLPShares())
				currValidators[valAddrStr] = val

				// get delegator voting power
				votingPower := getVoterPower(
					delegation.GetBondingShares(), delegation.GetLPShares(),
					val.DelegatorBondingShares, val.DelegatorLPShares,
					val.BondedTokens, val.LPTokens,
					voteLPRatio,
				)

				results[vote.Option] = results[vote.Option].Add(votingPower)
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})

		keeper.deleteVote(ctx, vote.ProposalID, vote.Voter)
		return false
	})

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if val.Vote == types.OptionEmpty {
			continue
		}

		// get shares leftovers (after delegator votes)
		bondingSharesAfterDeductions := val.DelegatorBondingShares.Sub(val.DelegatorBondingDeductions)
		lpSharesAfterDeductions := val.DelegatorLPShares.Sub(val.DelegatorLPDeductions)

		// get validator voting power
		votingPower := getVoterPower(
			bondingSharesAfterDeductions, lpSharesAfterDeductions,
			val.DelegatorBondingShares, val.DelegatorLPShares,
			val.BondedTokens, val.LPTokens,
			voteLPRatio,
		)

		results[val.Vote] = results[val.Vote].Add(votingPower)
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	// calculate total voting power
	maxVotingPower := keeper.sk.TotalBondedTokens(ctx).ToDec()
	maxVotingPower = maxVotingPower.Add(totalBondedLPTokens.ToDec().Mul(voteLPRatio))

	tallyParams := keeper.GetTallyParams(ctx)
	tallyResults = types.NewTallyResultFromMap(results, maxVotingPower.TruncateInt())

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	if keeper.sk.TotalBondedTokens(ctx).IsZero() {
		return false, false, tallyResults
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(maxVotingPower)
	if percentVoting.LT(tallyParams.Quorum) {
		return false, true, tallyResults
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[types.OptionAbstain]).Equal(sdk.ZeroDec()) {
		return false, false, tallyResults
	}

	// If more than 1/3 of voters veto, proposal fails
	if results[types.OptionNoWithVeto].Quo(totalVotingPower).GT(tallyParams.Veto) {
		return false, true, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[types.OptionYes].Quo(totalVotingPower.Sub(results[types.OptionAbstain])).GT(tallyParams.Threshold) {
		return true, false, tallyResults
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults
}

// getVoterPower returns delegator / validator total voting power based on bonding/lp shares and tokens.
func getVoterPower(
	voterBondingShares, voterLPShares sdk.Dec,
	totalBondingShares, totalLPShares sdk.Dec,
	totalBondingTokens, totalLPTokens sdk.Int,
	lpRatio sdk.Dec,
) sdk.Dec {

	bondingVotingPower, lpVotingPower := sdk.ZeroDec(), sdk.ZeroDec()

	// estimate bonding voting power
	if voterBondingShares.IsPositive() {
		bondingShareRatio := voterBondingShares.Quo(totalBondingShares)
		bondingVotingPower = bondingShareRatio.MulInt(totalBondingTokens)
	}

	// estimate bonding voting power
	if voterLPShares.IsPositive() {
		lpShareRatio := voterLPShares.Quo(totalLPShares)
		lpVotingPower = lpShareRatio.MulInt(totalLPTokens)
	}

	// get total voting power (bondingPower + lpPower * lpRatio)
	votingPower := bondingVotingPower.Add(lpVotingPower.Mul(lpRatio))

	return votingPower
}
