package simulation

// DONTCOVER

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

const (
	keyInflationMax            = "InflationMax"
	keyInflationMin            = "InflationMin"
	keyFeeBurningRatio         = "FeeBurningRatio"
	keyInfPwrBondedLockedRatio = "InfPwrBondedLockedRatio"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simulation.ParamChange {
	return []simulation.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, keyInflationMax,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMax(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyInflationMin,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInflationMin(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyFeeBurningRatio,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenFeeBurningRatio(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, keyInfPwrBondedLockedRatio,
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenInfPwrBondedLockedRatio(r))
			},
		),
	}
}
