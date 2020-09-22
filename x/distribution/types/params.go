package types

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	// default paramspace for params keeper
	DefaultParamspace = ModuleName
)

// Parameter keys
var (
	ParamKeyValidatorsPoolTax         = []byte("ValidatorsPoolTax")
	ParamKeyLiquidityProvidersPoolTax = []byte("LiquidityProvidersPoolTax")
	ParamKeyPublicTreasuryPoolTax     = []byte("PublicTreasuryPoolTax")
	ParamKeyHARPTax                   = []byte("HARPTax")
	//
	ParamKeyPublicTreasuryPoolCapacity = []byte("PublicTreasuryPoolCapacity")
	//
	ParamKeyBaseProposerReward  = []byte("BaseProposerReward")
	ParamKeyBonusProposerReward = []byte("BonusProposerReward")
	//
	ParamKeyWithdrawAddrEnabled = []byte("WithdrawAddrEnabled")
	ParamKeyFoundationNominees  = []byte("FoundationNominees")
)

// Params defines the set of distribution parameters.
type Params struct {
	// Rewards distribution ratio for ValidatorsPool
	ValidatorsPoolTax sdk.Dec `json:"validators_pool_tax" yaml:"validators_pool_tax"`
	// Rewards distribution ratio for LiquidityProvidersPool
	LiquidityProvidersPoolTax sdk.Dec `json:"liquidity_providers_pool_tax" yaml:"liquidity_providers_pool_tax"`
	// Rewards distribution ratio for PublicTreasuryPool
	PublicTreasuryPoolTax sdk.Dec `json:"public_treasury_pool_tax" yaml:"public_treasury_pool_tax"`
	// Rewards distribution ratio for HARP
	HARPTax sdk.Dec `json:"harp_tax" yaml:"harp_tax"`

	// PublicTreasuryPool max amount limit
	PublicTreasuryPoolCapacity sdk.Int `json:"public_treasury_pool_capacity"`

	// Block proposer base reward ratio
	BaseProposerReward sdk.Dec `json:"base_proposer_reward" yaml:"base_proposer_reward"`
	// Block proposer bonus reward ratio
	BonusProposerReward sdk.Dec `json:"bonus_proposer_reward" yaml:"bonus_proposer_reward"`

	//
	WithdrawAddrEnabled bool             `json:"withdraw_addr_enabled" yaml:"withdraw_addr_enabled"`
	FoundationNominees  []sdk.AccAddress `json:"foundation_nominees" yaml:"foundation_nominees"`
}

// ParamKeyTable returns the parameter key table.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns default distribution parameters
func DefaultParams() Params {
	return Params{
		ValidatorsPoolTax:         sdk.NewDecWithPrec(4825, 4), // 48.25%
		LiquidityProvidersPoolTax: sdk.NewDecWithPrec(4825, 4), // 48.25%
		PublicTreasuryPoolTax:     sdk.NewDecWithPrec(15, 3),   // 1.5%
		HARPTax:                   sdk.NewDecWithPrec(2, 2),    // 2%
		//
		PublicTreasuryPoolCapacity: sdk.NewInt(250000), // 250K (doesn't include currency decimals)
		//
		BaseProposerReward:  sdk.NewDecWithPrec(1, 2), // 1%
		BonusProposerReward: sdk.NewDecWithPrec(4, 2), // 4%
		//
		WithdrawAddrEnabled: true,
		FoundationNominees:  make([]sdk.AccAddress, 0),
	}
}

func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// ParamSetPairs returns the parameter set pairs.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		params.NewParamSetPair(ParamKeyValidatorsPoolTax, &p.ValidatorsPoolTax, validateValidatorsPoolTax),
		params.NewParamSetPair(ParamKeyLiquidityProvidersPoolTax, &p.LiquidityProvidersPoolTax, validateLiquidityProvidersPoolTax),
		params.NewParamSetPair(ParamKeyPublicTreasuryPoolTax, &p.PublicTreasuryPoolTax, validatePublicTreasuryPoolTax),
		params.NewParamSetPair(ParamKeyHARPTax, &p.HARPTax, validateParamKeyHARPTax),
		params.NewParamSetPair(ParamKeyPublicTreasuryPoolCapacity, &p.PublicTreasuryPoolCapacity, validatePublicTreasuryPoolCapacity),
		params.NewParamSetPair(ParamKeyBaseProposerReward, &p.BaseProposerReward, validateBaseProposerReward),
		params.NewParamSetPair(ParamKeyBonusProposerReward, &p.BonusProposerReward, validateBonusProposerReward),
		params.NewParamSetPair(ParamKeyWithdrawAddrEnabled, &p.WithdrawAddrEnabled, validateWithdrawAddrEnabled),
		params.NewParamSetPair(ParamKeyFoundationNominees, &p.FoundationNominees, validateFoundationNominees),
	}
}

// ValidateBasic performs basic validation on distribution parameters.
func (p Params) ValidateBasic() error {
	if err := validateValidatorsPoolTax(p.ValidatorsPoolTax); err != nil {
		return err
	}
	if err := validateLiquidityProvidersPoolTax(p.LiquidityProvidersPoolTax); err != nil {
		return err
	}
	if err := validatePublicTreasuryPoolTax(p.PublicTreasuryPoolTax); err != nil {
		return err
	}
	if err := validateParamKeyHARPTax(p.HARPTax); err != nil {
		return err
	}
	if err := validatePublicTreasuryPoolCapacity(p.PublicTreasuryPoolCapacity); err != nil {
		return err
	}
	if err := validateBaseProposerReward(p.BaseProposerReward); err != nil {
		return err
	}
	if err := validateBonusProposerReward(p.BonusProposerReward); err != nil {
		return err
	}
	if err := validateWithdrawAddrEnabled(p.WithdrawAddrEnabled); err != nil {
		return err
	}
	if err := validateFoundationNominees(p.FoundationNominees); err != nil {
		return err
	}

	if v := p.ValidatorsPoolTax.Add(p.LiquidityProvidersPoolTax).Add(p.PublicTreasuryPoolTax).Add(p.HARPTax); !v.Equal(sdk.OneDec()) {
		return fmt.Errorf("sum of all pool taxes must be 1.0: %s", v)
	}

	if v := p.BaseProposerReward.Add(p.BonusProposerReward); v.GT(sdk.OneDec()) {
		return fmt.Errorf("sum of base and bonus proposer reward cannot greater than one: %s", v)
	}

	return nil
}

func validateValidatorsPoolTax(i interface{}) error {
	return CheckRatioVariable("validators pool tax", i)
}

func validateLiquidityProvidersPoolTax(i interface{}) error {
	return CheckRatioVariable("liquidity providers pool tax", i)
}

func validatePublicTreasuryPoolTax(i interface{}) error {
	return CheckRatioVariable("public treasury pool tax", i)
}

func validateParamKeyHARPTax(i interface{}) error {
	return CheckRatioVariable("HARP tax", i)
}

func validatePublicTreasuryPoolCapacity(i interface{}) error {
	const paramName = "public treasury pool capacity"

	v, ok := i.(sdk.Int)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	if v.IsNegative() {
		return fmt.Errorf("%s: must be positive: %s", paramName, v)
	}

	if v.IsZero() {
		return fmt.Errorf("%s: cannot be zero: %s", paramName, v)
	}

	return nil
}

func validateBaseProposerReward(i interface{}) error {
	return CheckRatioVariable("base proposer reward", i)
}

func validateBonusProposerReward(i interface{}) error {
	return CheckRatioVariable("bonus proposer reward", i)
}

func validateWithdrawAddrEnabled(i interface{}) error {
	const paramName = "withdraw address enabled"

	_, ok := i.(bool)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	return nil
}

func validateFoundationNominees(i interface{}) error {
	const paramName = "foundation nominees"

	v, ok := i.([]sdk.AccAddress)
	if !ok {
		return fmt.Errorf("%s: invalid parameter type: %T", paramName, i)
	}

	for i, vv := range v {
		if vv.Empty() {
			return fmt.Errorf("%s: address [%d]: empty", paramName, i)
		}
	}

	return nil
}
