package types

import "github.com/cosmos/cosmos-sdk/x/staking"

// ABCIVote keeps extended consensus vote info used for rewards allocation.
// ABCIVote is used in the BeginBlocker.
type ABCIVote struct {
	Validator         staking.ValidatorI
	DistributionPower int64
	SignedLastBlock   bool
}

// ABCIVotes is a ABCIVote slice.
type ABCIVotes []ABCIVote

// TotalDistributionPower returns sum of all voters distribution powers.
func (vv ABCIVotes) TotalDistributionPower() int64 {
	total := int64(0)
	for _, v := range vv {
		total += v.DistributionPower
	}

	return total
}
