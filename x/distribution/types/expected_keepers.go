package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingexported "github.com/cosmos/cosmos-sdk/x/staking/exported"
	supplyexported "github.com/cosmos/cosmos-sdk/x/supply/exported"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authexported.Account
}

// StakingKeeper expected staking keeper (noalias)
type StakingKeeper interface {
	// Iterate through validators by operator address, execute func for each validator
	IterateValidators(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// Iterate through bonded validators by operator address, execute func for each validator
	IterateBondedValidatorsByPower(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// Iterate through the consensus validator set of the last block by operator address, execute func for each validator
	IterateLastValidators(sdk.Context,
		func(index int64, validator stakingexported.ValidatorI) (stop bool))

	// Get a particular validator by operator address
	Validator(sdk.Context, sdk.ValAddress) stakingexported.ValidatorI
	// Get a particular validator by consensus address
	ValidatorByConsAddr(sdk.Context, sdk.ConsAddress) stakingexported.ValidatorI
	// Get a particular validator by operator address (direct call)
	GetValidator(ctx sdk.Context, addr sdk.ValAddress) (staking.Validator, bool)
	// Get all validators
	GetAllValidators(ctx sdk.Context) []staking.Validator
	// Get a particular validator delegations state
	GetValidatorStakingState(ctx sdk.Context, addr sdk.ValAddress) staking.ValidatorStakingState

	// slash the validator and delegators of the validator, specifying offence height, offence power, and slash fraction
	Slash(sdk.Context, sdk.ConsAddress, int64, int64, sdk.Dec)
	Jail(sdk.Context, sdk.ConsAddress)   // jail a validator
	Unjail(sdk.Context, sdk.ConsAddress) // unjail a validator

	// Delegation allows for getting a particular delegation for a given validator
	// and delegator outside the scope of the staking module.
	Delegation(sdk.Context, sdk.AccAddress, sdk.ValAddress) stakingexported.DelegationI

	// MaxValidators returns the maximum amount of bonded validators
	MaxValidators(sdk.Context) uint16

	// LPDistrRatio coefficient
	LPDistrRatio(sdk.Context) sdk.Dec

	// MaxDelegationsRatio coefficient
	MaxDelegationsRatio(ctx sdk.Context) sdk.Dec

	// Staking BondedPool current supply
	TotalBondedTokens(sdk.Context) sdk.Int

	IterateDelegations(ctx sdk.Context, delegator sdk.AccAddress,
		fn func(index int64, delegation stakingexported.DelegationI) (stop bool))

	GetLastTotalPower(ctx sdk.Context) sdk.Int
	GetLastValidatorPower(ctx sdk.Context, valAddr sdk.ValAddress) int64

	GetAllSDKDelegations(ctx sdk.Context) []staking.Delegation
}

// StakingHooks event hooks for staking validator object (noalias)
type StakingHooks interface {
	AfterValidatorCreated(ctx sdk.Context, valAddr sdk.ValAddress)                           // Must be called when a validator is created
	AfterValidatorRemoved(ctx sdk.Context, consAddr sdk.ConsAddress, valAddr sdk.ValAddress) // Must be called when a validator is deleted

	BeforeDelegationCreated(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)        // Must be called when a delegation is created
	BeforeDelegationSharesModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress) // Must be called when a delegation's shares are modified
	AfterDelegationModified(ctx sdk.Context, delAddr sdk.AccAddress, valAddr sdk.ValAddress)
	BeforeValidatorSlashed(ctx sdk.Context, valAddr sdk.ValAddress, fraction sdk.Dec)
}

// SupplyKeeper defines the expected supply Keeper (noalias)
type SupplyKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, name string) supplyexported.ModuleAccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(sdk.Context, supplyexported.ModuleAccountI)

	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}
