package types

import (
	"errors"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Parameter store keys
var (
	KeyMintDenom           = []byte("MintDenom")
	KeyInflationRateChange = []byte("InflationRateChange")
	KeyInflationMax        = []byte("InflationMax")
	KeyInflationMin        = []byte("InflationMin")
	KeyGoalBonded          = []byte("GoalBonded")
	KeyAvgBlockTimeWindow  = []byte("AvgBlockTimeWindow")
)

// mint parameters
type Params struct {
	MintDenom           string  `json:"mint_denom" yaml:"mint_denom"`                       // type of coin to mint
	InflationRateChange sdk.Dec `json:"inflation_rate_change" yaml:"inflation_rate_change"` // maximum annual change in inflation rate
	InflationMax        sdk.Dec `json:"inflation_max" yaml:"inflation_max"`                 // maximum inflation rate
	InflationMin        sdk.Dec `json:"inflation_min" yaml:"inflation_min"`                 // minimum inflation rate
	GoalBonded          sdk.Dec `json:"goal_bonded" yaml:"goal_bonded"`                     // goal of percent bonded atoms
	AvgBlockTimeWindow  uint16  `json:"avg_block_time_window" yaml:"avg_block_time_window"` // avg block time filter window size
}

// ParamTable for minting module.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

func NewParams(
	mintDenom string,
	inflationRateChange, inflationMax, inflationMin, goalBonded sdk.Dec,
	avgBlockTimeWindow uint16,
) Params {
	return Params{
		MintDenom:           mintDenom,
		InflationRateChange: inflationRateChange,
		InflationMax:        inflationMax,
		InflationMin:        inflationMin,
		GoalBonded:          goalBonded,
		AvgBlockTimeWindow:  avgBlockTimeWindow,
	}
}

// default minting module parameters
func DefaultParams() Params {
	return Params{
		MintDenom:           sdk.DefaultBondDenom,
		InflationRateChange: sdk.NewDecWithPrec(13, 2),
		InflationMax:        sdk.NewDecWithPrec(20, 2),
		InflationMin:        sdk.NewDecWithPrec(7, 2),
		GoalBonded:          sdk.NewDecWithPrec(67, 2),
		AvgBlockTimeWindow:  100,
	}
}

// validate params
func (p Params) Validate() error {
	if err := validateMintDenom(p.MintDenom); err != nil {
		return err
	}
	if err := validateInflationRateChange(p.InflationRateChange); err != nil {
		return err
	}
	if err := validateInflationMax(p.InflationMax); err != nil {
		return err
	}
	if err := validateInflationMin(p.InflationMin); err != nil {
		return err
	}
	if err := validateGoalBonded(p.GoalBonded); err != nil {
		return err
	}
	if err := validateAvgBlockTimeWindow(p.AvgBlockTimeWindow); err != nil {
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
  Mint Denom:               %s
  Inflation Rate Change:    %s
  Inflation Max:            %s
  Inflation Min:            %s
  Goal Bonded:              %s
  Avg blocksPerYear Window: %d
`,
		p.MintDenom, p.InflationRateChange, p.InflationMax,
		p.InflationMin, p.GoalBonded, p.AvgBlockTimeWindow,
	)
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyMintDenom, &p.MintDenom, validateMintDenom),
		params.NewParamSetPair(KeyInflationRateChange, &p.InflationRateChange, validateInflationRateChange),
		params.NewParamSetPair(KeyInflationMax, &p.InflationMax, validateInflationMax),
		params.NewParamSetPair(KeyInflationMin, &p.InflationMin, validateInflationMin),
		params.NewParamSetPair(KeyGoalBonded, &p.GoalBonded, validateGoalBonded),
		params.NewParamSetPair(KeyAvgBlockTimeWindow, &p.AvgBlockTimeWindow, validateAvgBlockTimeWindow),
	}
}

func validateMintDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("mint denom cannot be blank")
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateInflationRateChange(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("inflation rate change cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("inflation rate change too large: %s", v)
	}

	return nil
}

func validateInflationMax(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("max inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("max inflation too large: %s", v)
	}

	return nil
}

func validateInflationMin(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("min inflation cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("min inflation too large: %s", v)
	}

	return nil
}

func validateGoalBonded(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.IsNegative() {
		return fmt.Errorf("goal bonded cannot be negative: %s", v)
	}
	if v.GT(sdk.OneDec()) {
		return fmt.Errorf("goal bonded too large: %s", v)
	}

	return nil
}

func validateAvgBlockTimeWindow(i interface{}) error {
	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v < 2 {
		return fmt.Errorf("avg blockTime window must be GTE than 2: %d", v)
	}

	return nil
}
