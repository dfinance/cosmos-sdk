package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyPublicTreasuryCapacity = "PublicTreasuryPoolCapacity"
	keyBaseProposerReward     = "BaseProposerReward"
	keyBonusProposerReward    = "BonusProposerReward"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyPublicTreasuryCapacity,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenPublicTreasuryCapacity(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyBaseProposerReward,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBaseProposerReward(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyBonusProposerReward,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenBonusProposerReward(r))
			},
		),
	}
}
