package distribution

import (
	"github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// nolint

const (
	ModuleName                          = types.ModuleName
	RewardsBankPoolName                 = types.RewardsBankPoolName
	StoreKey                            = types.StoreKey
	RouterKey                           = types.RouterKey
	QuerierRoute                        = types.QuerierRoute
	ProposalTypePublicTreasuryPoolSpend = types.ProposalTypePublicTreasuryPoolSpend
	QueryParams                         = types.QueryParams
	QueryValidatorOutstandingRewards    = types.QueryValidatorOutstandingRewards
	QueryValidatorCommission            = types.QueryValidatorCommission
	QueryValidatorSlashes               = types.QueryValidatorSlashes
	QueryDelegationRewards              = types.QueryDelegationRewards
	QueryDelegatorTotalRewards          = types.QueryDelegatorTotalRewards
	QueryDelegatorValidators            = types.QueryDelegatorValidators
	QueryWithdrawAddr                   = types.QueryWithdrawAddr
	QueryPool                           = types.QueryPool
	QueryLockedRewardsState             = types.QueryLockedRewardsState
	DefaultParamspace                   = types.DefaultParamspace
	TypeMsgFundPublicTreasuryPool       = types.TypeMsgFundPublicTreasuryPool
)

var (
	// functions aliases
	RegisterInvariants                         = keeper.RegisterInvariants
	AllInvariants                              = keeper.AllInvariants
	NonNegativeOutstandingInvariant            = keeper.NonNegativeOutstandingInvariant
	CanWithdrawInvariant                       = keeper.CanWithdrawInvariant
	ReferenceCountInvariant                    = keeper.ReferenceCountInvariant
	ModuleAccountInvariant                     = keeper.ModuleAccountInvariant
	NewKeeper                                  = keeper.NewKeeper
	GetValidatorOutstandingRewardsAddress      = types.GetValidatorOutstandingRewardsAddress
	GetDelegatorWithdrawInfoAddress            = types.GetDelegatorWithdrawInfoAddress
	GetDelegatorStartingInfoAddresses          = types.GetDelegatorStartingInfoAddresses
	GetValidatorHistoricalRewardsAddressPeriod = types.GetValidatorHistoricalRewardsAddressPeriod
	GetValidatorCurrentRewardsAddress          = types.GetValidatorCurrentRewardsAddress
	GetValidatorAccumulatedCommissionAddress   = types.GetValidatorAccumulatedCommissionAddress
	GetValidatorSlashEventAddressHeight        = types.GetValidatorSlashEventAddressHeight
	GetValidatorOutstandingRewardsKey          = types.GetValidatorOutstandingRewardsKey
	GetDelegatorWithdrawAddrKey                = types.GetDelegatorWithdrawAddrKey
	GetDelegatorStartingInfoKey                = types.GetDelegatorStartingInfoKey
	GetValidatorHistoricalRewardsPrefix        = types.GetValidatorHistoricalRewardsPrefix
	GetValidatorHistoricalRewardsKey           = types.GetValidatorHistoricalRewardsKey
	GetValidatorCurrentRewardsKey              = types.GetValidatorCurrentRewardsKey
	GetValidatorAccumulatedCommissionKey       = types.GetValidatorAccumulatedCommissionKey
	GetValidatorSlashEventPrefix               = types.GetValidatorSlashEventPrefix
	GetValidatorSlashEventKeyPrefix            = types.GetValidatorSlashEventKeyPrefix
	GetValidatorSlashEventKey                  = types.GetValidatorSlashEventKey
	HandlePublicTreasuryPoolSpendProposal      = keeper.HandlePublicTreasuryPoolSpendProposal
	NewQuerier                                 = keeper.NewQuerier
	MakeTestCodec                              = keeper.MakeTestCodec
	CreateTestInputDefault                     = keeper.CreateTestInputDefault
	CreateTestInputAdvanced                    = keeper.CreateTestInputAdvanced
	ParamKeyTable                              = types.ParamKeyTable
	DefaultParams                              = types.DefaultParams
	RegisterCodec                              = types.RegisterCodec
	NewDelegatorStartingInfo                   = types.NewDelegatorStartingInfo
	ErrEmptyDelegatorAddr                      = types.ErrEmptyDelegatorAddr
	ErrEmptyWithdrawAddr                       = types.ErrEmptyWithdrawAddr
	ErrEmptyValidatorAddr                      = types.ErrEmptyValidatorAddr
	ErrEmptyDelegationDistInfo                 = types.ErrEmptyDelegationDistInfo
	ErrNoValidatorDistInfo                     = types.ErrNoValidatorDistInfo
	ErrNoValidatorExists                       = types.ErrNoValidatorExists
	ErrNoDelegationExists                      = types.ErrNoDelegationExists
	ErrNoValidatorCommission                   = types.ErrNoValidatorCommission
	ErrSetWithdrawAddrDisabled                 = types.ErrSetWithdrawAddrDisabled
	ErrBadDistribution                         = types.ErrBadDistribution
	ErrInvalidProposalAmount                   = types.ErrInvalidProposalAmount
	ErrEmptyProposalRecipient                  = types.ErrEmptyProposalRecipient
	ErrWithdrawLocked                          = types.ErrWithdrawLocked
	ErrInvalidLockOperation                    = types.ErrInvalidLockOperation
	InitialRewardPools                         = types.InitialRewardPools
	NewGenesisState                            = types.NewGenesisState
	DefaultGenesisState                        = types.DefaultGenesisState
	ValidateGenesis                            = types.ValidateGenesis
	NewMsgSetWithdrawAddress                   = types.NewMsgSetWithdrawAddress
	NewMsgWithdrawDelegatorReward              = types.NewMsgWithdrawDelegatorReward
	NewMsgWithdrawValidatorCommission          = types.NewMsgWithdrawValidatorCommission
	NewMsgFundPublicTreasuryPool               = types.NewMsgFundPublicTreasuryPool
	NewMsgWithdrawFoundationPool               = types.NewMsgWithdrawFoundationPool
	NewPublicTreasuryPoolSpendProposal         = types.NewPublicTreasuryPoolSpendProposal
	NewTaxParamsUpdateProposal                 = types.NewTaxParamsUpdateProposal
	NewQueryValidatorOutstandingRewardsParams  = types.NewQueryValidatorOutstandingRewardsParams
	NewQueryValidatorCommissionParams          = types.NewQueryValidatorCommissionParams
	NewQueryValidatorSlashesParams             = types.NewQueryValidatorSlashesParams
	NewQueryDelegationRewardsParams            = types.NewQueryDelegationRewardsParams
	NewQueryDelegatorParams                    = types.NewQueryDelegatorParams
	NewQueryDelegatorWithdrawAddrParams        = types.NewQueryDelegatorWithdrawAddrParams
	NewQueryDelegatorTotalRewardsResponse      = types.NewQueryDelegatorTotalRewardsResponse
	NewDelegationDelegatorReward               = types.NewDelegationDelegatorReward
	NewValidatorHistoricalRewards              = types.NewValidatorHistoricalRewards
	NewValidatorCurrentRewards                 = types.NewValidatorCurrentRewards
	InitialValidatorAccumulatedCommission      = types.InitialValidatorAccumulatedCommission
	NewValidatorSlashEvent                     = types.NewValidatorSlashEvent

	// variable aliases
	RewardPoolsKey                       = types.RewardPoolsKey
	ProposerKey                          = types.ProposerKey
	ValidatorOutstandingRewardsPrefix    = types.ValidatorOutstandingRewardsPrefix
	DelegatorWithdrawAddrPrefix          = types.DelegatorWithdrawAddrPrefix
	DelegatorStartingInfoPrefix          = types.DelegatorStartingInfoPrefix
	ValidatorHistoricalRewardsPrefix     = types.ValidatorHistoricalRewardsPrefix
	ValidatorCurrentRewardsPrefix        = types.ValidatorCurrentRewardsPrefix
	ValidatorAccumulatedCommissionPrefix = types.ValidatorAccumulatedCommissionPrefix
	ValidatorSlashEventPrefix            = types.ValidatorSlashEventPrefix
	ParamKeyValidatorsPoolTax            = types.ParamKeyValidatorsPoolTax
	ParamKeyLiquidityProvidersPoolTax    = types.ParamKeyLiquidityProvidersPoolTax
	ParamKeyPublicTreasuryPoolTax        = types.ParamKeyPublicTreasuryPoolTax
	ParamKeyHARPTax                      = types.ParamKeyHARPTax
	ParamKeyBaseProposerReward           = types.ParamKeyBaseProposerReward
	ParamKeyBonusProposerReward          = types.ParamKeyBonusProposerReward
	ParamKeyWithdrawAddrEnabled          = types.ParamKeyWithdrawAddrEnabled
	ParamKeyFoundationNominees           = types.ParamKeyFoundationNominees
	ModuleCdc                            = types.ModuleCdc
	EventTypeSetWithdrawAddress          = types.EventTypeSetWithdrawAddress
	EventTypeRewards                     = types.EventTypeRewards
	EventTypeCommission                  = types.EventTypeCommission
	EventTypeWithdrawRewards             = types.EventTypeWithdrawRewards
	EventTypeWithdrawCommission          = types.EventTypeWithdrawCommission
	EventTypeProposerReward              = types.EventTypeProposerReward
	AttributeKeyWithdrawAddress          = types.AttributeKeyWithdrawAddress
	AttributeKeyValidator                = types.AttributeKeyValidator
	AttributeValueCategory               = types.AttributeValueCategory
	LiquidityProvidersPoolName           = types.LiquidityProvidersPoolName
	FoundationPoolName                   = types.FoundationPoolName
	PublicTreasuryPoolName               = types.PublicTreasuryPoolName
	HARPName                             = types.HARPName
	PublicTreasurySpendProposalHandler   = client.PublicTreasurySpendProposalHandler
	TaxParamsUpdateProposalHandler       = client.TaxParamsUpdateProposalHandler
)

type (
	Hooks                                  = keeper.Hooks
	Keeper                                 = keeper.Keeper
	DelegatorStartingInfo                  = types.DelegatorStartingInfo
	RewardPoolName                         = types.RewardPoolName
	RewardPools                            = types.RewardPools
	DelegatorWithdrawInfo                  = types.DelegatorWithdrawInfo
	ValidatorOutstandingRewardsRecord      = types.ValidatorOutstandingRewardsRecord
	ValidatorAccumulatedCommissionRecord   = types.ValidatorAccumulatedCommissionRecord
	ValidatorHistoricalRewardsRecord       = types.ValidatorHistoricalRewardsRecord
	ValidatorCurrentRewardsRecord          = types.ValidatorCurrentRewardsRecord
	DelegatorStartingInfoRecord            = types.DelegatorStartingInfoRecord
	ValidatorSlashEventRecord              = types.ValidatorSlashEventRecord
	Params                                 = types.Params
	GenesisState                           = types.GenesisState
	MsgSetWithdrawAddress                  = types.MsgSetWithdrawAddress
	MsgWithdrawDelegatorReward             = types.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorCommission         = types.MsgWithdrawValidatorCommission
	MsgWithdrawFoundationPool              = types.MsgWithdrawFoundationPool
	PublicTreasuryPoolSpendProposal        = types.PublicTreasuryPoolSpendProposal
	TaxParamsUpdateProposal                = types.TaxParamsUpdateProposal
	QueryValidatorOutstandingRewardsParams = types.QueryValidatorOutstandingRewardsParams
	QueryValidatorCommissionParams         = types.QueryValidatorCommissionParams
	QueryValidatorSlashesParams            = types.QueryValidatorSlashesParams
	QueryDelegationRewardsParams           = types.QueryDelegationRewardsParams
	QueryDelegatorParams                   = types.QueryDelegatorParams
	QueryDelegatorWithdrawAddrParams       = types.QueryDelegatorWithdrawAddrParams
	QueryDelegatorTotalRewardsResponse     = types.QueryDelegatorTotalRewardsResponse
	DelegationDelegatorReward              = types.DelegationDelegatorReward
	ValidatorHistoricalRewards             = types.ValidatorHistoricalRewards
	ValidatorCurrentRewards                = types.ValidatorCurrentRewards
	ValidatorAccumulatedCommission         = types.ValidatorAccumulatedCommission
	ValidatorSlashEvent                    = types.ValidatorSlashEvent
	ValidatorSlashEvents                   = types.ValidatorSlashEvents
	ValidatorOutstandingRewards            = types.ValidatorOutstandingRewards
	ABCIVote                               = types.ABCIVote
	ABCIVotes                              = types.ABCIVotes
)
