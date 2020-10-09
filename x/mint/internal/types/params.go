package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom                 = []byte("MintDenom")
	KeyInflationMax              = []byte("InflationMax")
	KeyInflationMin              = []byte("InflationMin")
	KeyFeeBurningRatio           = []byte("FeeBurningRatio")
	KeyInfPwrBondedLockedRatio   = []byte("InfPwrBondedLockedRatio")
	KeyFoundationAllocationRatio = []byte("FoundationAllocRatio")
	KeyAvgBlockTimeWindow        = []byte("AvgBlockTimeWindow")
	KeyStakingTotalSupplyShift   = []byte("StakingTotalSupplyShift")
)

// mint parameters
type Params struct {
	// Type of coin to mint
	MintDenom string `json:"mint_denom" yaml:"mint_denom" example:"stake"`
	// Maximum inflation rate (annual)
	InflationMax sdk.Dec `json:"inflation_max" yaml:"inflation_max" swaggertype:"string" format:"number"  example:"0.123"`
	// Minimum inflation rate (annual)
	InflationMin sdk.Dec `json:"inflation_min" yaml:"inflation_min" swaggertype:"string" format:"number"  example:"0.123"`
	// % of fees burned (per block)
	FeeBurningRatio sdk.Dec `json:"fee_burning_ratio" yaml:"fee_burning_ratio" swaggertype:"string" format:"number"  example:"0.123"`
	// Bonded, locked shoulders relation for inflation power calculation
	InfPwrBondedLockedRatio sdk.Dec `json:"infpwr_bondedlocked_ratio" yaml:"infpwr_bondedlocked_ratio" swaggertype:"string" format:"number"  example:"0.123"`
	// Extra Foundation pool allocation inflation ratio
	FoundationAllocationRatio sdk.Dec `json:"foundation_allocation_ratio" yaml:"foundation_allocation_ratio" swaggertype:"string" format:"number"  example:"0.123"`
	// Avg block time filter window size
	AvgBlockTimeWindow uint16 `json:"avg_block_time_window" yaml:"avg_block_time_window"`
	// BondedRatio denominator (TotalSupply) shift coefficient
	StakingTotalSupplyShift sdk.Int `json:"staking_total_supply_shift" yaml:"staking_total_supply_shift" swaggertype:"string" format:"integer"  example:"100"`
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	inflationMax, inflationMin sdk.Dec,
	feeBurningRatio sdk.Dec,
	infPwrBondedLockedRatio, foundationAllocationRatio sdk.Dec,
	avgBlockTimeWindow uint16,
	stakingTotalSupplyShift sdk.Int,
) Params {
	return Params{
		MintDenom:                 mintDenom,
		InflationMax:              inflationMax,
		InflationMin:              inflationMin,
		FeeBurningRatio:           feeBurningRatio,
		InfPwrBondedLockedRatio:   infPwrBondedLockedRatio,
		FoundationAllocationRatio: foundationAllocationRatio,
		AvgBlockTimeWindow:        avgBlockTimeWindow,
		StakingTotalSupplyShift:   stakingTotalSupplyShift,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:                 sdk.DefaultBondDenom,
		InflationMax:              sdk.NewDecWithPrec(50, 2),   // 50%
		InflationMin:              sdk.NewDecWithPrec(1776, 4), // 17.76%
		FeeBurningRatio:           sdk.NewDecWithPrec(50, 2),   // 50%
		InfPwrBondedLockedRatio:   sdk.NewDecWithPrec(4, 1),    // 40%
		FoundationAllocationRatio: sdk.NewDecWithPrec(45, 2),   // 45%
		AvgBlockTimeWindow:        100,                         // 100 blocks
		StakingTotalSupplyShift:   sdk.ZeroInt(),               // no shift
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationMax(p.InflationMax); err != nil {
		return err
	}
	if err := validateInflationMin(p.InflationMin); err != nil {
		return err
	}
	if err := validateInfPwrBondedLockedRatio(p.InfPwrBondedLockedRatio); err != nil {
		return err
	}
	if err := validateFeeBurningRatio(p.FeeBurningRatio); err != nil {
		return err
	}
	if err := validateFoundationAllocationRatio(p.FoundationAllocationRatio); err != nil {
		return err
	}
	if err := validateAvgBlockTimeWindow(p.AvgBlockTimeWindow); err != nil {
		return err
	}
	if err := validateStakingTotalSupplyShift(p.StakingTotalSupplyShift); err != nil {
		return err
	}

	if p.InflationMax.LT(p.InflationMin) {
		return fmt.Errorf(
			"max inflation (%s) must be greater than or equal to min inflation (%s)",
			p.InflationMax, p.InflationMin,
		)
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf(`Minting Params:
  Mint Denom:                   %s
  Inflation Max:                %s
  Inflation Min:                %s
  Fee burning ratio:            %s
  InfPower bonded/locked ratio: %s
  Foundation allocation ratio:  %s
  Avg blocksPerYear Window:     %d
`,
		p.MintDenom,
		p.InflationMax,
		p.InflationMin,
		p.FeeBurningRatio,
		p.InfPwrBondedLockedRatio,
		p.FoundationAllocationRatio,
		p.AvgBlockTimeWindow,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		params.NewParamSetPair(KeyInflationMax, &p.InflationMax, validateInflationMax),
		params.NewParamSetPair(KeyInflationMin, &p.InflationMin, validateInflationMin),
		params.NewParamSetPair(KeyFeeBurningRatio, &p.FeeBurningRatio, validateFeeBurningRatio),
		params.NewParamSetPair(KeyInfPwrBondedLockedRatio, &p.InfPwrBondedLockedRatio, validateInfPwrBondedLockedRatio),
		params.NewParamSetPair(KeyFoundationAllocationRatio, &p.FoundationAllocationRatio, validateFoundationAllocationRatio),
		params.NewParamSetPair(KeyAvgBlockTimeWindow, &p.AvgBlockTimeWindow, validateAvgBlockTimeWindow),
		params.NewParamSetPair(KeyStakingTotalSupplyShift, &p.StakingTotalSupplyShift, validateStakingTotalSupplyShift),
	}
}

func validateMintDenom(i interface{}) error {
	const paramName = "mint denom"

	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s: cannot be blank", paramName)
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return fmt.Errorf("%s: %v", paramName, err)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	return CheckRatioVariable("max inflation", i)
}

func validateInflationMin(i interface{}) error {
	return CheckRatioVariable("min inflation", i)
}

func validateFeeBurningRatio(i interface{}) error {
	return CheckRatioVariable("fee burning ratio", i)
}

func validateInfPwrBondedLockedRatio(i interface{}) error {
	return CheckRatioVariable("inflation power bonded/locked ratio", i)
}

func validateFoundationAllocationRatio(i interface{}) error {
	const paramName = "foundation allocation ratio"

	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.IsNegative() {
		return fmt.Errorf("%s: cannot be nagative: %s", paramName, v)
	}

	if v.GT(FoundationAllocationRatioMaxValue) {
		return fmt.Errorf("%s: cannot be greater than %s: %s", paramName, FoundationAllocationRatioMaxValue, v)
	}

	return nil
}

func validateAvgBlockTimeWindow(i interface{}) error {
	const paramName = "avg blockTime window"

	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v < 2 {
		return fmt.Errorf("%s: must be GTE than 2: %d", paramName, v)
	}

	return nil
}

func validateStakingTotalSupplyShift(i interface{}) error {
	const paramName = "staking totalSupply shift"

	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.IsNegative() {
		return fmt.Errorf("%s: must be GTE than 0: %s", paramName, v.String())
	}

	return nil
}
