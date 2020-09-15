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
	DefaultUnbondingTime time.Duration = time.Hour * 24 * 7 * 3

	// Default maximum number of bonded validators
	DefaultMaxValidators uint16 = 100

	// Default maximum entries in a UBD/RED pair
	DefaultMaxEntries uint16 = 7

	// DefaultHistorical entries is 0 since it must only be non-zero for
	// IBC connected chains
	DefaultHistoricalEntries uint16 = 0

	// Default validator.MinSelfDelegation level
	DefaultMinSelfDelegationLvl = 10000

	// Default max delegations ratio (10.0)
	DefaultMaxDelegationsRatioBase      = 10
	DefaultMaxDelegationsRatioPrecision = 0
)

// nolint - Keys for parameter access
var (
	KeyUnbondingTime        = []byte("UnbondingTime")
	KeyMaxValidators        = []byte("MaxValidators")
	KeyMaxEntries           = []byte("KeyMaxEntries")
	KeyBondDenom            = []byte("BondDenom")
	KeyHistoricalEntries    = []byte("HistoricalEntries")
	KeyMinSelfDelegationLvl = []byte("MinSelfDelegationLvl")
	KeyMaxDelegationsRatio  = []byte("MaxDelegationsRatio")
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
	BondDenom string `json:"bond_denom" yaml:"bond_denom"`
	// Min self delegation level for validator creation
	MinSelfDelegationLvl sdk.Int `json:"min_self_delegation_lvl" yaml:"min_self_delegation_lvl"`
	// Max delegations per validator is limited by (CurrentSelfDelegation * KeyMaxDelegationsRatio)
	MaxDelegationsRatio sdk.Dec `json:"max_delegations_ratio" yaml:"max_delegations_ratio"`
}

// NewParams creates a new Params instance
func NewParams(
	unbondingTime time.Duration,
	maxValidators, maxEntries, historicalEntries uint16,
	bondDenom string,
	minSelfDelegationLvl sdk.Int,
	maxDelegationsRatio sdk.Dec,
) Params {

	return Params{
		UnbondingTime:        unbondingTime,
		MaxValidators:        maxValidators,
		MaxEntries:           maxEntries,
		HistoricalEntries:    historicalEntries,
		BondDenom:            bondDenom,
		MinSelfDelegationLvl: minSelfDelegationLvl,
		MaxDelegationsRatio:  maxDelegationsRatio,
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
		params.NewParamSetPair(KeyMinSelfDelegationLvl, &p.MinSelfDelegationLvl, validateMinSelfDelegationLvl),
		params.NewParamSetPair(KeyMaxDelegationsRatio, &p.MaxDelegationsRatio, validateMaxDelegationsRatio),
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
		sdk.NewInt(DefaultMinSelfDelegationLvl),
		sdk.NewDecWithPrec(DefaultMaxDelegationsRatioBase, DefaultMaxDelegationsRatioPrecision),
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
  Min SelfDelegation lvl: %s`,
		p.UnbondingTime, p.MaxValidators, p.MaxEntries, p.HistoricalEntries, p.BondDenom, p.MinSelfDelegationLvl)
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
	if err := validateMinSelfDelegationLvl(p.MinSelfDelegationLvl); err != nil {
		return err
	}
	if err := validateMaxDelegationsRatio(p.MaxDelegationsRatio); err != nil {
		return err
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
