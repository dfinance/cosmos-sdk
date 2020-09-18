package mint

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/mint/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

const (
	ModuleName                              = types.ModuleName
	DefaultParamspace                       = types.DefaultParamspace
	StoreKey                                = types.StoreKey
	QuerierRoute                            = types.QuerierRoute
	QueryParameters                         = types.QueryParameters
	QueryInflation                          = types.QueryInflation
	QueryAnnualProvisions                   = types.QueryAnnualProvisions
	QueryBlocksPerYear                      = types.QueryBlocksPerYear
	ChangeFoundationAllocationRatioTTL      = types.ChangeFoundationAllocationRatioTTL
	ChangeFoundationAllocationRatioMaxValue = types.ChangeFoundationAllocationRatioMaxValue
	ChangeFoundationAllocationRatioMinValue = types.ChangeFoundationAllocationRatioMinValue
)

var (
	// functions aliases
	NewKeeper            = keeper.NewKeeper
	NewQuerier           = keeper.NewQuerier
	NewGenesisState      = types.NewGenesisState
	DefaultGenesisState  = types.DefaultGenesisState
	ValidateGenesis      = types.ValidateGenesis
	RegisterCodec        = types.RegisterCodec
	NewMinter            = types.NewMinter
	InitialMinter        = types.InitialMinter
	DefaultInitialMinter = types.DefaultInitialMinter
	ValidateMinter       = types.ValidateMinter
	ParamKeyTable        = types.ParamKeyTable
	NewParams            = types.NewParams
	DefaultParams        = types.DefaultParams

	// variable aliases
	ErrWrongFoundationAllocationRatio = types.ErrWrongFoundationAllocationRatio
	ErrExceededTimeLimit              = types.ErrExceededTimeLimit
	ModuleCdc                         = types.ModuleCdc
	MinterKey                         = types.MinterKey
	KeyMintDenom                      = types.KeyMintDenom
	KeyInflationMax                   = types.KeyInflationMax
	KeyInflationMin                   = types.KeyInflationMin
	KeyFeeBurningRatio                = types.KeyFeeBurningRatio
	KeyInfPwrBondedLockedRatio        = types.KeyInfPwrBondedLockedRatio
	KeyFoundationAllocationRatio      = types.KeyFoundationAllocationRatio
	KeyAvgBlockTimeWindow             = types.KeyAvgBlockTimeWindow
)

type (
	Keeper       = keeper.Keeper
	GenesisState = types.GenesisState
	Minter       = types.Minter
	Params       = types.Params
)
