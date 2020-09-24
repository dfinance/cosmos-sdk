package mint

// nolint

import (
	"github.com/cosmos/cosmos-sdk/x/mint/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

const (
	ModuleName            = types.ModuleName
	DefaultParamspace     = types.DefaultParamspace
	StoreKey              = types.StoreKey
	QuerierRoute          = types.QuerierRoute
	QueryParameters       = types.QueryParameters
	QueryInflation        = types.QueryInflation
	QueryAnnualProvisions = types.QueryAnnualProvisions
	QueryBlocksPerYear    = types.QueryBlocksPerYear
)

var (
	// functions aliases
	NewKeeper            = keeper.NewKeeper
	NewQuerier           = keeper.NewQuerier
	NewGenesisState      = types.NewGenesisState
	DefaultGenesisState  = types.DefaultGenesisState
	ValidateGenesis      = types.ValidateGenesis
	NewMinter            = types.NewMinter
	InitialMinter        = types.InitialMinter
	DefaultInitialMinter = types.DefaultInitialMinter
	ValidateMinter       = types.ValidateMinter
	ParamKeyTable        = types.ParamKeyTable
	NewParams            = types.NewParams
	DefaultParams        = types.DefaultParams

	// variable aliases
	ModuleCdc                         = types.ModuleCdc
	MinterKey                         = types.MinterKey
	KeyMintDenom                      = types.KeyMintDenom
	KeyInflationMax                   = types.KeyInflationMax
	KeyInflationMin                   = types.KeyInflationMin
	KeyFeeBurningRatio                = types.KeyFeeBurningRatio
	KeyInfPwrBondedLockedRatio        = types.KeyInfPwrBondedLockedRatio
	KeyFoundationAllocationRatio      = types.KeyFoundationAllocationRatio
	KeyAvgBlockTimeWindow             = types.KeyAvgBlockTimeWindow
	FoundationAllocationRatioMaxValue = types.FoundationAllocationRatioMaxValue
)

type (
	Keeper         = keeper.Keeper
	GenesisState   = types.GenesisState
	Minter         = types.Minter
	Params         = types.Params
	BlockDurFilter = types.BlockDurFilter
)
