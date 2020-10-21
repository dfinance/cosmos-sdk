package types

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Staking params default values
const (
	// DefaultUnbondingTime reflects three weeks in seconds as the default
	// unbonding time.
	// TODO: Justify our choice of default here.
	DefaultUnbondingTime = time.Hour * 24 * 7 * 3
	MaxUnbondingTime     = time.Hour * 24 * 7 * 4 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint16 = 100

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint16 = 7

	// DefaultHistorical entries is 0 since it must only be non-zero for
	// IBC connected chains
	DefaultHistoricalEntries uint16 = 0

	// Default validator.MinSelfDelegation level
	DefaultMinSelfDelegationLvl = 10000

	// Default max self-delegation level
	// Value is set high for tests to pass, should be defined by the client app
	DefaultMaxSelfDelegationLvl = 1000000000000

	// Default max delegations ratio (10.0)
	DefaultMaxDelegationsRatioBase      = 10
	DefaultMaxDelegationsRatioPrecision = 0

	// Default duration for validator.ScheduledToUnbond flag to be raised up
	// After the period is over, force validator unbond is performed
	DefaultScheduledUnbondTime time.Duration = time.Hour * 24 * 7
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime            = []byte("UnbondingTime")
	KeyMaxValidators            = []byte("MaxValidators")
	KeyMaxEntries               = []byte("KeyMaxEntries")
	KeyBondDenom                = []byte("BondDenom")
	KeyLPDenom                  = []byte("LPDenom")
	KeyLPDistrRatio             = []byte("LPDistrRatio")
	KeyHistoricalEntries        = []byte("HistoricalEntries")
	KeyMinSelfDelegationLvl     = []byte("MinSelfDelegationLvl")
	KeyMaxSelfDelegationLvl     = []byte("MaxSelfDelegationLvl")
	KeyMaxDelegationsRatio      = []byte("MaxDelegationsRatio")
	KeyScheduledUnbondDelayTime = []byte("ScheduledUnbondDelayTime")
)

var _ params.ParamSet = (*Params)(nil)

// Params defines the high level settings for staking
type Params struct {
	// Time duration of unbonding
	UnbondingTime time.Duration `json:"unbonding_time" yaml:"unbonding_time"`
	// Maximum number of validators (max uint16 = 65535)
	MaxValidators uint16 `json:"max_validators" yaml:"max_validators"`
	// Max entries for either unbonding delegation or redelegation (per pair/trio)
	MaxEntries uint16 `json:"max_entries" yaml:"max_entries"`
	// Number of historical entries to persist
	HistoricalEntries uint16 `json:"historical_entries" yaml:"historical_entries"`
	// Bondable coin denomination
	BondDenom string `json:"bond_denom" yaml:"bond_denom" example:"stake"`
	// Liquidity coin denomination
	LPDenom string `json:"lp_denom" yaml:"lp_denom" example:"liqd"`
	// Gov voting and distribution ratio between bonding tokens and liquidity tokens (BTokens + LPDistrRatio * LPTokens)
	LPDistrRatio sdk.Dec `json:"lp_distr_ratio" yaml:"lp_distr_ratio" swaggertype:"string" format:"number" example:"0.123"`
	// Min self delegation level for validator creation
	MinSelfDelegationLvl sdk.Int `json:"min_self_delegation_lvl" yaml:"min_self_delegation_lvl" swaggertype:"string" format:"integer" example:"100"`
	// Max self delegation level for self-delegation increment
	MaxSelfDelegationLvl sdk.Int `json:"max_self_delegation_lvl" yaml:"max_self_delegation_lvl" swaggertype:"string" format:"integer" example:"100"`
	// Max delegations per validator is limited by (CurrentSelfDelegation * KeyMaxDelegationsRatio)
	MaxDelegationsRatio sdk.Dec `json:"max_delegations_ratio" yaml:"max_delegations_ratio" swaggertype:"string" format:"number" example:"0.123"`
	// Time duration of validator.ScheduledToUnbond to be raised up before forced unbonding is done
	ScheduledUnbondDelayTime time.Duration `json:"scheduled_unbond_delay" yaml:"scheduled_unbond_delay"`
} //@name StakingParams

// NewParams creates a new Params instance
func NewParams(
	unbondingTime time.Duration,
	maxValidators, maxEntries, historicalEntries uint16,
	bondDenom, lpDenom string,
	lpDistrRatio sdk.Dec,
	minSelfDelegationLvl sdk.Int, maxSelfDelegationLvl sdk.Int,
	maxDelegationsRatio sdk.Dec,
	scheduledUnbondDelay time.Duration,
) Params {

	return Params{
		UnbondingTime:            unbondingTime,
		MaxValidators:            maxValidators,
		MaxEntries:               maxEntries,
		HistoricalEntries:        historicalEntries,
		BondDenom:                bondDenom,
		LPDenom:                  lpDenom,
		LPDistrRatio:             lpDistrRatio,
		MinSelfDelegationLvl:     minSelfDelegationLvl,
		MaxSelfDelegationLvl:     maxSelfDelegationLvl,
		MaxDelegationsRatio:      maxDelegationsRatio,
		ScheduledUnbondDelayTime: scheduledUnbondDelay,
	}
}

// Implements params.ParamSet
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(KeyUnbondingTime, &p.UnbondingTime, validateUnbondingTime),
		params.NewParamSetPair(KeyMaxValidators, &p.MaxValidators, validateMaxValidators),
		params.NewParamSetPair(KeyMaxEntries, &p.MaxEntries, validateMaxEntries),
		params.NewParamSetPair(KeyHistoricalEntries, &p.HistoricalEntries, validateHistoricalEntries),
		params.NewParamSetPair(KeyBondDenom, &p.BondDenom, validateBondDenom),
		params.NewParamSetPair(KeyLPDenom, &p.LPDenom, validateLPDenom),
		params.NewParamSetPair(KeyLPDistrRatio, &p.LPDistrRatio, validateLPDistrRatio),
		params.NewParamSetPair(KeyMinSelfDelegationLvl, &p.MinSelfDelegationLvl, validateMinSelfDelegationLvl),
		params.NewParamSetPair(KeyMaxSelfDelegationLvl, &p.MaxSelfDelegationLvl, validateMaxSelfDelegationLvl),
		params.NewParamSetPair(KeyMaxDelegationsRatio, &p.MaxDelegationsRatio, validateMaxDelegationsRatio),
		params.NewParamSetPair(KeyScheduledUnbondDelayTime, &p.ScheduledUnbondDelayTime, validateScheduledUnbondDelayTime),
	}
}

// Equal returns a boolean determining if two Param types are identical.
// TODO: This is slower than comparing struct fields directly
func (p Params) Equal(p2 Params) bool {
	bz1 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p)
	bz2 := ModuleCdc.MustMarshalBinaryLengthPrefixed(&p2)
	return bytes.Equal(bz1, bz2)
}

// DefaultParams returns a default set of parameters.
func DefaultParams() Params {
	return NewParams(
		DefaultUnbondingTime,
		DefaultMaxValidators,
		DefaultMaxEntries,
		DefaultHistoricalEntries,
		sdk.DefaultBondDenom,
		sdk.DefaultLiquidityDenom,
		sdk.NewDecWithPrec(1, 0),
		sdk.NewInt(DefaultMinSelfDelegationLvl),
		sdk.NewInt(DefaultMaxSelfDelegationLvl),
		sdk.NewDecWithPrec(DefaultMaxDelegationsRatioBase, DefaultMaxDelegationsRatioPrecision),
		DefaultScheduledUnbondTime,
	)
}

// String returns a human readable string representation of the parameters.
func (p Params) String() string {
	return fmt.Sprintf(`Params:
  Unbonding Time:         %s
  Max Validators:         %d
  Max Entries:            %d
  Historical Entries:     %d
  Bonded Coin Denom:      %s
  Liquidity Coin Denom:   %s
  LP Tokens Distr Ratio:  %s
  Min SelfDelegation lvl: %s
  Max SelfDelegation lvl: %s
  Max Delegations Ratio:  %s
  Scheduled Unbond Delay: %s
`,
		p.UnbondingTime, p.MaxValidators, p.MaxEntries, p.HistoricalEntries,
		p.BondDenom, p.LPDenom, p.LPDistrRatio,
		p.MinSelfDelegationLvl, p.MaxSelfDelegationLvl,
		p.MaxDelegationsRatio, p.ScheduledUnbondDelayTime,
	)
}

// unmarshal the current staking params value from store key or panic
func MustUnmarshalParams(cdc *codec.Codec, value []byte) Params {
	params, err := UnmarshalParams(cdc, value)
	if err != nil {
		panic(err)
	}
	return params
}

// unmarshal the current staking params value from store key
func UnmarshalParams(cdc *codec.Codec, value []byte) (params Params, err error) {
	err = cdc.UnmarshalBinaryLengthPrefixed(value, &params)
	if err != nil {
		return
	}
	return
}

// validate a set of params
func (p Params) Validate() error {
	if err := validateUnbondingTime(p.UnbondingTime); err != nil {
		return err
	}
	if err := validateMaxValidators(p.MaxValidators); err != nil {
		return err
	}
	if err := validateMaxEntries(p.MaxEntries); err != nil {
		return err
	}
	if err := validateBondDenom(p.BondDenom); err != nil {
		return err
	}
	if err := validateLPDenom(p.LPDenom); err != nil {
		return err
	}
	if err := validateLPDistrRatio(p.LPDistrRatio); err != nil {
		return err
	}
	if err := validateMinSelfDelegationLvl(p.MinSelfDelegationLvl); err != nil {
		return err
	}
	if err := validateMaxSelfDelegationLvl(p.MaxSelfDelegationLvl); err != nil {
		return err
	}
	if err := validateMaxDelegationsRatio(p.MaxDelegationsRatio); err != nil {
		return err
	}
	if err := validateScheduledUnbondDelayTime(p.ScheduledUnbondDelayTime); err != nil {
		return err
	}

	if p.MaxSelfDelegationLvl.LT(p.MinSelfDelegationLvl) {
		return fmt.Errorf("max self-delegation level < min: %s < %s", p.MaxSelfDelegationLvl, p.MinSelfDelegationLvl)
	}

	return nil
}

func validateUnbondingTime(i interface{}) error {
	const paramName = "unbonding time"

	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v <= 0 {
		return fmt.Errorf("%s: must be positive: %d", paramName, v)
	}
	if v > MaxUnbondingTime {
		return fmt.Errorf("%s: must be LT %v: %v", paramName, MaxUnbondingTime, v)
	}

	return nil
}

func validateMaxValidators(i interface{}) error {
	const paramName = "max validators"

	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v == 0 {
		return fmt.Errorf("%s: must be positive: %d", paramName, v)
	}

	return nil
}

func validateMaxEntries(i interface{}) error {
	const paramName = "max entries"

	v, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v == 0 {
		return fmt.Errorf("%s: must be positive: %d", paramName, v)
	}

	return nil
}

func validateHistoricalEntries(i interface{}) error {
	const paramName = "historical entries"

	_, ok := i.(uint16)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	return nil
}

func validateBondDenom(i interface{}) error {
	const paramName = "bond denom"

	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s: cannot be blank", paramName)
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return fmt.Errorf("%s: validation: %v", paramName, err)
	}

	return nil
}

func validateLPDenom(i interface{}) error {
	const paramName = "liquidity denom"

	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("%s: cannot be blank", paramName)
	}
	if err := sdk.ValidateDenom(v); err != nil {
		return fmt.Errorf("%s: validation: %v", paramName, err)
	}

	return nil
}

func validateMinSelfDelegationLvl(i interface{}) error {
	const paramName = "min self-delegation level"

	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("%s: must be GT than zero: %s", paramName, v.String())
	}

	return nil
}

func validateMaxSelfDelegationLvl(i interface{}) error {
	const paramName = "max self-delegation level"

	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("%s: must be GT than zero: %s", paramName, v.String())
	}

	return nil
}

func validateMaxDelegationsRatio(i interface{}) error {
	const paramName = "max delegations ratio"

	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.LT(sdk.OneDec()) {
		return fmt.Errorf("%s: must be GTE than 1.0: %s", paramName, v.String())
	}

	return nil
}

func validateScheduledUnbondDelayTime(i interface{}) error {
	const paramName = "scheduled unbond delay time"

	v, ok := i.(time.Duration)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v <= 0 {
		return fmt.Errorf("%s: must be positive: %d", paramName, v)
	}

	return nil
}

func validateLPDistrRatio(i interface{}) error {
	const paramName = "liquidity tokens distribution ratio"

	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.IsNegative() {
		return fmt.Errorf("%s: must be GTE than 0.0: %s", paramName, v.String())
	}

	return nil
}
