package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	// Get the total rewards balance from all delegations
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards",
		delegatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Query a delegation reward
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/rewards/{validatorAddr}",
		delegationRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the rewards withdrawal address
	r.HandleFunc(
		"/distribution/delegators/{delegatorAddr}/withdraw_address",
		delegatorWithdrawalAddrHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Validator distribution information
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}",
		validatorInfoHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Commission and self-delegation rewards of a single a validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/rewards",
		validatorRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Outstanding rewards of a single validator
	r.HandleFunc(
		"/distribution/validators/{validatorAddr}/outstanding_rewards",
		outstandingRewardsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the current distribution parameter values
	r.HandleFunc(
		"/distribution/parameters",
		paramsHandlerFn(cliCtx, queryRoute),
	).Methods("GET")

	// Get the amount held in the specified pool
	r.HandleFunc(
		"/distribution/pool/{poolName}",
		poolHandler(cliCtx, queryRoute),
	).Methods("GET")

}

// delegatorRewardsHandlerFn godoc
// @Tags Distribution
// @Summary Get the total rewards balance from all delegations
// @Description Get the sum of all the rewards earned by delegations by a single delegator
// @ID distributionGetDelegatorRewards
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {object} []types.DelegationDelegatorReward
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/rewards [get]
func delegatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		params := types.NewQueryDelegatorParams(delegatorAddr)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to marshal params: %s", err))
			return
		}

		route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorTotalRewards)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// delegationRewardsHandlerFn godoc
// @Tags Distribution
// @Summary Query a delegation reward
// @Description Query a single delegation reward by a delegator
// @ID distributionGetDelegationRewards
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {object} []types.DecCoin
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/rewards/{validatorAddr} [get]
func delegationRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		delAddr := mux.Vars(r)["delegatorAddr"]
		valAddr := mux.Vars(r)["validatorAddr"]

		// query for rewards from a particular delegation
		res, height, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, delAddr, valAddr)
		if !ok {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// delegatorWithdrawalAddrHandlerFn godoc
// @Tags Distribution
// @Summary Get the rewards withdrawal address
// @Description Get the delegations' rewards withdrawal address. This is the address in which the user will receive the reward funds
// @ID distributionGetDelegatorWithdrawalAddr
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {string} Token "Bech32 AccAddress of the rewards withdrawal address"
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/delegators/{delegatorAddr}/withdraw_address [get]
func delegatorWithdrawalAddrHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delegatorAddr, ok := checkDelegatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bz := cliCtx.Codec.MustMarshalJSON(types.NewQueryDelegatorWithdrawAddrParams(delegatorAddr))
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryWithdrawAddr), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// ValidatorDistInfo defines the properties of
// validator distribution information response.
type ValidatorDistInfo struct {
	OperatorAddress     sdk.AccAddress                       `json:"operator_address" yaml:"operator_address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
	ValidatorCommission types.ValidatorAccumulatedCommission `json:"validator_commission" yaml:"validator_commission"`
	SelfBondRewards     sdk.DecCoins                         `json:"self_bond_rewards" yaml:"self_bond_rewards"`
}

// NewValidatorDistInfo creates a new instance of ValidatorDistInfo.
func NewValidatorDistInfo(operatorAddr sdk.AccAddress, rewards sdk.DecCoins,
	commission types.ValidatorAccumulatedCommission) ValidatorDistInfo {
	return ValidatorDistInfo{
		OperatorAddress:     operatorAddr,
		SelfBondRewards:     rewards,
		ValidatorCommission: commission,
	}
}

// validatorInfoHandlerFn godoc
// @Tags Distribution
// @Summary Validator distribution information
// @Description Query the distribution information of a single validator
// @ID distributionGetValidatorInfo
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 OperatorAddress of validator"
// @Success 200 {object} SwaggerValidatorDistInfo
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr} [get]
func validatorInfoHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// query commission
		bz, err := common.QueryValidatorCommission(cliCtx, queryRoute, valAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var commission types.ValidatorAccumulatedCommission
		if err := cliCtx.Codec.UnmarshalJSON(bz, &commission); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// self bond rewards
		delAddr := sdk.AccAddress(valAddr)
		bz, height, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, delAddr.String(), valAddr.String())
		if !ok {
			return
		}

		var rewards sdk.DecCoins
		if err := cliCtx.Codec.UnmarshalJSON(bz, &rewards); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		bz, err = cliCtx.Codec.MarshalJSON(NewValidatorDistInfo(delAddr, rewards, commission))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, bz)
	}
}

// validatorRewardsHandlerFn godoc
// @Tags Distribution
// @Summary Commission and self-delegation rewards of a single validator
// @Description Query the commission and self-delegation rewards of validator
// @ID distributionGetValidatorRewards
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 OperatorAddress of validator"
// @Success 200 {object} types.DecCoins
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr}/rewards [get]
func validatorRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		valAddr := mux.Vars(r)["validatorAddr"]
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		delAddr := sdk.AccAddress(validatorAddr).String()
		bz, height, ok := checkResponseQueryDelegationRewards(w, cliCtx, queryRoute, delAddr, valAddr)
		if !ok {
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, bz)
	}
}

// paramsHandlerFn godoc
// @Tags Distribution
// @Summary Fee distribution parameters
// @Description Fee distribution parameters
// @ID distributionGetParams
// @Accept  json
// @Produce json
// @Success 200 {object} types.Params
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/parameters [get]
func paramsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// poolHandler godoc
// @Tags Distribution
// @Summary Get the amount held in the specified pool
// @Description Get the amount held in the specified pool
// @ID distributionPool
// @Accept  json
// @Produce json
// @Param poolName path string true "PoolName: LiquidityProvidersPool, FoundationPool, PublicTreasuryPool, HARP"
// @Success 200 {object} types.DecCoins
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/pool/{poolName} [get]
func poolHandler(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		poolName := mux.Vars(r)["poolName"]
		if poolName == "" {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "poolName: empty")
			return
		}

		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, types.QueryPool, poolName), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var result sdk.DecCoins
		if err := cliCtx.Codec.UnmarshalJSON(res, &result); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, result)
	}
}

// outstandingRewardsHandlerFn godoc
// @Tags Distribution
// @Summary Fee distribution outstanding rewards of a single validator
// @Description Fee distribution outstanding rewards of a single validator
// @ID distributionOutstandingRewards
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 OperatorAddress of validator"
// @Success 200 {object} types.DecCoins
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /distribution/validators/{validatorAddr}/outstanding_rewards [get]
func outstandingRewardsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		validatorAddr, ok := checkValidatorAddressVar(w, r)
		if !ok {
			return
		}

		cliCtx, ok = rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bin := cliCtx.Codec.MustMarshalJSON(types.NewQueryValidatorOutstandingRewardsParams(validatorAddr))
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorOutstandingRewards), bin)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func checkResponseQueryDelegationRewards(
	w http.ResponseWriter, cliCtx context.CLIContext, queryRoute, delAddr, valAddr string,
) (res []byte, height int64, ok bool) {

	res, height, err := common.QueryDelegationRewards(cliCtx, queryRoute, delAddr, valAddr)
	if err != nil {
		rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return nil, 0, false
	}

	return res, height, true
}
