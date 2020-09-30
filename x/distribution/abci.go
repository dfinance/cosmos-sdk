package distribution

import (
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint"
)

// BeginBlocker sets the proposer for determining distribution during endBlock and distributes rewards for the previous block.
// Validator power is converted to distribution power which includes lockedRewards.
// Moving from stakingPower to distributionPower is used to rebalance distribution proportions.
func BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock, k keeper.Keeper, mk mint.Keeper) {
	consVotes := req.LastCommitInfo.GetVotes()

	// process the rewards unlock queue here as distributionPower might change
	k.ProcessAllMatureRewardsUnlockQueueItems(ctx)

	// get LPRatioCoef
	lpRatio := k.GetValidatorLPDistrRatio(ctx)

	// determine the total distribution power signing the block, override voter's power with distribution power
	var previousTotalPower, previousProposerPower int64
	abciVotes := make(ABCIVotes, 0, len(consVotes))
	for _, consVote := range consVotes {
		validator := k.ValidatorByConsAddr(ctx, consVote.Validator.Address)
		distrPower := k.GetDistributionPower(ctx, validator.GetOperator(), consVote.Validator.Power, validator.LPPower(), lpRatio)

		previousTotalPower += distrPower
		if consVote.SignedLastBlock {
			previousProposerPower += distrPower
		}

		abciVotes = append(abciVotes, ABCIVote{
			Validator:         validator,
			DistributionPower: distrPower,
			SignedLastBlock:   consVote.SignedLastBlock,
		})
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		// calculate dynamic FoundationPool tax based on previous mint results
		minter := mk.GetMinter(ctx)
		dynamicFoundationPoolTax := minter.FoundationInflation.Quo(minter.Inflation.Add(minter.FoundationInflation))

		previousProposer := k.GetPreviousProposerConsAddr(ctx)

		k.AllocateTokens(ctx, previousProposerPower, previousTotalPower, previousProposer, abciVotes, dynamicFoundationPoolTax)
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
