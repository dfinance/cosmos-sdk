package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Minter represents the minting state.
type Minter struct {
	// Current annual inflation rate
	Inflation sdk.Dec `json:"inflation" yaml:"inflation"`
	// Current annual foundation inflation rate
	FoundationInflation sdk.Dec `json:"foundation_inflation" yaml:"foundation_inflation"`
	// Current annual expected provisions
	Provisions sdk.Dec `json:"provisions" yaml:"provisions"`
	// Current annual expected FoundationPool provisions
	FoundationProvisions sdk.Dec `json:"foundation_provisions" yaml:"foundation_provisions"`
	// Current annual number of blocks estimation
	BlocksPerYear uint64
}

// NewMinter returns a new Minter object with the given inflation and annual
// provisions values.
func NewMinter(inflation, foundationInflation, provisions, foundationProvisions sdk.Dec, blocksPerYear uint64) Minter {
	return Minter{
		Inflation:            inflation,
		FoundationInflation:  foundationInflation,
		Provisions:           provisions,
		FoundationProvisions: foundationProvisions,
		BlocksPerYear:        blocksPerYear,
	}
}

// InitialMinter returns an initial Minter object with a given inflation value.
func InitialMinter(inflation sdk.Dec) Minter {
	return NewMinter(
		inflation,
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		sdk.ZeroDec(),
		0,
	)
}

// DefaultInitialMinter returns a default initial Minter object for a new chain.
func DefaultInitialMinter() Minter {
	return InitialMinter(
		sdk.NewDecWithPrec(1776, 4), // 17.76%
	)
}

// validate minter
func ValidateMinter(minter Minter) error {
	if minter.Inflation.IsNegative() {
		return fmt.Errorf("mint parameter Inflation should be positive, is %s", minter.Inflation)
	}
	if minter.FoundationInflation.IsNegative() {
		return fmt.Errorf("mint parameter FoundationInflation should be positive, is %s", minter.FoundationInflation)
	}

	return nil
}

// NextMinMaxInflation returns next Min and Max inflation level (annual adjustment).
func (m Minter) NextMinMaxInflation(params Params) (nextMin, nextMax sdk.Dec) {
	// MinInflation = MinInflation / 2
	nextMin = params.InflationMin.Quo(sdk.NewDecWithPrec(2, 0))

	// MaxInflation = MaxInflation / 2 + (MaxInflation - ActualInflation)
	nextMax = params.InflationMax.Quo(sdk.NewDecWithPrec(2, 0)).Add(params.InflationMax.Sub(m.Inflation))

	// sanity check (TODO: removed later)
	if params.InflationMin.GT(params.InflationMax) {
		panic(fmt.Errorf("nextMinInlfation > nextMaxInflation: %s / %s", nextMin, nextMax))
	}

	return
}

// NextInflationPower returns the new inflation power.
func (m Minter) NextInflationPower(params Params, bondedRatio, lockedRatio sdk.Dec) sdk.Dec {
	var (
		shoulderPoint     = sdk.NewDecWithPrec(8, 1) // 0.8
		bondedLockedRatio = params.InfPwrBondedLockedRatio
	)

	// Sanity check
	if err := CheckRatioVariable("bondedRatio", bondedRatio); err != nil {
		panic(err)
	}
	if err := CheckRatioVariable("lockedRatio", lockedRatio); err != nil {
		panic(err)
	}

	// function f(u)
	bondedShoulder := func() sdk.Dec {
		if bondedRatio.LT(shoulderPoint) {
			// 0 <= BondedRatio < 0.8: 1 - BondedRatio / 0.8
			return sdk.OneDec().Sub(bondedRatio.Quo(shoulderPoint))
		} else if bondedRatio.Equal(shoulderPoint) {
			// 0.8: 0.0
			return sdk.ZeroDec()
		}
		// 0.8 < BondedRatio <= 1: 3 * (BondedRatio - 0.8)
		return sdk.NewDecWithPrec(3, 0).Mul(bondedRatio.Sub(shoulderPoint))
	}

	// function g(u)
	lockedShoulder := func() sdk.Dec {
		if lockedRatio.LT(shoulderPoint) {
			// 0 <= LockedRatio < 0.8: LockedRatio / 0.8
			return lockedRatio.Quo(shoulderPoint)
		} else if bondedRatio.Equal(shoulderPoint) {
			// 0.8: 1.0
			return sdk.OneDec()
		}
		// 0.8 < LockedRatio <= 1: 1 - 3 * (LockedRatio - 0.8)
		return sdk.OneDec().Sub(sdk.NewDecWithPrec(3, 0).Mul(lockedRatio.Sub(shoulderPoint)))
	}

	// InflationPower = InfPwrBondedLockedRatio * BondedShoulder + (1 - InfPwrBondedLockedRatio) * LockedShoulder
	infPowerBonded := bondedLockedRatio.Mul(bondedShoulder())
	infPowerLocked := sdk.OneDec().Sub(bondedLockedRatio).Mul(lockedShoulder())
	infPower := infPowerBonded.Add(infPowerLocked)

	// sanity check (TODO: remove later)
	if err := CheckRatioVariable("infPower", infPower); err != nil {
		panic(err)
	}

	return infPower
}

// NextInflationRate returns the new inflation rate capped to [min, max].
// ActualInflation = MinInflation + (MaxInflation - MinInflation) * InflationPower.
func (m Minter) NextInflationRate(params Params, inflationPower sdk.Dec) sdk.Dec {
	// sanity input check
	if params.InflationMin.GT(params.InflationMax) {
		panic(fmt.Errorf("minInflation GT maxInflation: %s / %s", params.InflationMin, params.InflationMax))
	}
	inflation := params.InflationMin.Add(params.InflationMax.Sub(params.InflationMin).Mul(inflationPower))
	// sanity check (TODO: remove later)
	if err := CheckRatioVariable("inflation", inflation); err != nil {
		panic(err)
	}
	if inflation.LT(params.InflationMin) {
		// that should not happen, leave this for now
		inflation = params.InflationMin
	}
	if inflation.GT(params.InflationMax) {
		inflation = params.InflationMax
	}

	return inflation
}

// NextFoundationInflationRate returns the new foundation inflation rate used for FoundationPool tokens allocation.
// FoundationInflation = min( MaxInflation - ActualInflation, ActualInflation * FoundationAllocationRatio ).
func (m Minter) NextFoundationInflationRate(params Params) sdk.Dec {
	value1 := params.InflationMax.Sub(m.Inflation)
	value2 := m.Inflation.Mul(params.FoundationAllocationRatio)

	if value1.LT(value2) {
		return value1
	}

	return value2
}

// NextAnnualProvisions returns the annual provisions based on current total supply and inflation rate.
func (m Minter) NextAnnualProvisions(_ Params, totalSupply sdk.Int) (inflation, foundationInflation sdk.Dec) {
	return m.Inflation.MulInt(totalSupply), m.FoundationInflation.MulInt(totalSupply)
}

// BlockProvision returns the provisions for a block based on the annual provisions rates.
func (m Minter) BlockProvision(params Params) sdk.Coin {
	// sanity check
	if m.BlocksPerYear == 0 {
		panic("blocksPerYear iz zero")
	}

	totalProvision := m.Provisions.Add(m.FoundationProvisions)
	provisionAmt := totalProvision.QuoInt64(int64(m.BlocksPerYear))

	return sdk.NewCoin(params.MintDenom, provisionAmt.TruncateInt())
}

func (m Minter) String() string {
	return fmt.Sprintf(`Minter:
  Inflation:                %s
  Foundation inflation:     %s
  Provision:                %s
  Foundation provision:     %s
  BlocksPerYear estimation: %d`,
		m.Inflation, m.FoundationInflation,
		m.Provisions, m.FoundationProvisions,
		m.BlocksPerYear,
	)
}
