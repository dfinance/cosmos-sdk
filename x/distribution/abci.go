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
	votes := req.LastCommitInfo.GetVotes()

	// determine the total distribution power signing the block
	// override voter's power with distribution power
	var previousTotalPower, previousProposerPower int64
	for i := 0; i < len(votes); i++ {
		voteInfo := votes[i]
		distrPower := k.GetDistributionPower(ctx, voteInfo.Validator.Address, voteInfo.Validator.Power)

		previousTotalPower += distrPower
		if voteInfo.SignedLastBlock {
			previousProposerPower += distrPower
		}

		voteInfo.Validator.Power = distrPower
		votes[i] = voteInfo
	}

	// TODO this is Tendermint-dependent
	// ref https://github.com/cosmos/cosmos-sdk/issues/3095
	if ctx.BlockHeight() > 1 {
		// calculate dynamic FoundationPool tax based on previous mint results
		minter := mk.GetMinter(ctx)
		dynamicFoundationPoolTax := minter.FoundationInflation.Quo(minter.Inflation.Add(minter.FoundationInflation))

		previousProposer := k.GetPreviousProposerConsAddr(ctx)

		k.AllocateTokens(ctx, previousProposerPower, previousTotalPower, previousProposer, votes, dynamicFoundationPoolTax)
	}

	// record the proposer for when we payout on the next block
	consAddr := sdk.ConsAddress(req.Header.ProposerAddress)
	k.SetPreviousProposerConsAddr(ctx, consAddr)
}
