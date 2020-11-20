package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	// Get all delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations",
		delegatorDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all unbonding delegations from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations",
		delegatorUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all staking txs (i.e msgs) from a delegator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/txs",
		delegatorTxsHandlerFn(cliCtx),
	).Methods("GET")

	// Query all validators that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators",
		delegatorValidatorsHandlerFn(cliCtx),
	).Methods("GET")

	// Query a validator that a delegator is bonded to
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/validators/{validatorAddr}",
		delegatorValidatorHandlerFn(cliCtx),
	).Methods("GET")

	// Query a delegation between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/delegations/{validatorAddr}",
		delegationHandlerFn(cliCtx),
	).Methods("GET")

	// Query all unbonding delegations between a delegator and a validator
	r.HandleFunc(
		"/staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr}",
		unbondingDelegationHandlerFn(cliCtx),
	).Methods("GET")

	// Query redelegations (filters in query params)
	r.HandleFunc(
		"/staking/redelegations",
		redelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all validators
	r.HandleFunc(
		"/staking/validators",
		validatorsHandlerFn(cliCtx),
	).Methods("GET")

	// Get a single validator info
	r.HandleFunc(
		"/staking/validators/{validatorAddr}",
		validatorHandlerFn(cliCtx),
	).Methods("GET")

	// Get all delegations to a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/delegations",
		validatorDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get all unbonding delegations from a validator
	r.HandleFunc(
		"/staking/validators/{validatorAddr}/unbonding_delegations",
		validatorUnbondingDelegationsHandlerFn(cliCtx),
	).Methods("GET")

	// Get HistoricalInfo at a given height
	r.HandleFunc(
		"/staking/historical_info/{height}",
		historicalInfoHandlerFn(cliCtx),
	).Methods("GET")

	// Get the current state of the staking pool
	r.HandleFunc(
		"/staking/pool",
		poolHandlerFn(cliCtx),
	).Methods("GET")

	// Get the current staking parameter values
	r.HandleFunc(
		"/staking/parameters",
		paramsHandlerFn(cliCtx),
	).Methods("GET")

}

// delegatorDelegationsHandlerFn godoc
// @Tags Staking
// @Summary Get all delegations from a delegator
// @Description Get all delegations from a delegator
// @ID stakingGetDelegatorDelegations
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {object} QueryDelegationsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations [get]
func delegatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorDelegations))
}

// delegatorUnbondingDelegationsHandlerFn godoc
// @Tags Staking
// @Summary Get all unbonding delegations from a delegator
// @Description Get all unbonding delegations from a delegator
// @ID stakingGetDelegatorUnbondingDelegations
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {object} QueryUnbondingDelegationsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/unbonding_delegations [get]
func delegatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorUnbondingDelegations))
}

// delegatorTxsHandlerFn godoc
// @Tags Staking
// @Summary Query all staking txs (msgs) from a delegator
// @Description Query all staking txs (msgs) from a delegator
// @ID stakingGetDelegatorTxs
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Param type query string false "Unbonding types via space: bond unbond redelegate"
// @Success 200 {object} []types.SearchTxsResult
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/txs [get]
func delegatorTxsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var typesQuerySlice []string
		vars := mux.Vars(r)
		delegatorAddr := vars["delegatorAddr"]

		_, err := sdk.AccAddressFromBech32(delegatorAddr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		typesQuery := r.URL.Query().Get("type")
		trimmedQuery := strings.TrimSpace(typesQuery)
		if len(trimmedQuery) != 0 {
			typesQuerySlice = strings.Split(trimmedQuery, " ")
		}

		noQuery := len(typesQuerySlice) == 0
		isBondTx := contains(typesQuerySlice, "bond")
		isUnbondTx := contains(typesQuerySlice, "unbond")
		isRedTx := contains(typesQuerySlice, "redelegate")

		var (
			txs     []*sdk.SearchTxsResult
			actions []string
		)

		switch {
		case isBondTx:
			actions = append(actions, types.MsgDelegate{}.Type())

		case isUnbondTx:
			actions = append(actions, types.MsgUndelegate{}.Type())

		case isRedTx:
			actions = append(actions, types.MsgBeginRedelegate{}.Type())

		case noQuery:
			actions = append(actions, types.MsgDelegate{}.Type())
			actions = append(actions, types.MsgUndelegate{}.Type())
			actions = append(actions, types.MsgBeginRedelegate{}.Type())

		default:
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, action := range actions {
			foundTxs, errQuery := queryTxs(cliCtx, action, delegatorAddr)
			if errQuery != nil {
				rest.WriteErrorResponse(w, http.StatusInternalServerError, errQuery.Error())
			}
			txs = append(txs, foundTxs)
		}

		res, err := cliCtx.Codec.MarshalJSON(txs)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponseBare(w, cliCtx, res)
	}
}

// unbondingDelegationHandlerFn godoc
// @Tags Staking
// @Summary Query all unbonding delegations between a delegator and a validator
// @Description Query all unbonding delegations between a delegator and a validator
// @ID stakingGetUnbondingDelegation
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Param validatorAddr path string true "Bech32 OperatorAddress of validator"
// @Success 200 {object} QueryUnbondingDelegationResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/unbonding_delegations/{validatorAddr} [get]
func unbondingDelegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryUnbondingDelegation))
}

// redelegationsHandlerFn godoc
// @Tags Staking
// @Summary Get all redelegations (filter by query params)
// @Description Get all redelegations (filter by query params)
// @ID stakingGetRedelegations
// @Accept  json
// @Produce json
// @Param delegator query string false "Bech32 AccAddress of Delegator"
// @Param validator_from query string false "Bech32 AccAddress of SrcValidator"
// @Param validator_to query string false "Bech32 AccAddress of DstValidator"
// @Success 200 {object} QueryRedelegationsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/redelegations [get]
func redelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var params types.QueryRedelegationParams

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		bechDelegatorAddr := r.URL.Query().Get("delegator")
		bechSrcValidatorAddr := r.URL.Query().Get("validator_from")
		bechDstValidatorAddr := r.URL.Query().Get("validator_to")

		if len(bechDelegatorAddr) != 0 {
			delegatorAddr, err := sdk.AccAddressFromBech32(bechDelegatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.DelegatorAddr = delegatorAddr
		}

		if len(bechSrcValidatorAddr) != 0 {
			srcValidatorAddr, err := sdk.ValAddressFromBech32(bechSrcValidatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.SrcValidatorAddr = srcValidatorAddr
		}

		if len(bechDstValidatorAddr) != 0 {
			dstValidatorAddr, err := sdk.ValAddressFromBech32(bechDstValidatorAddr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}
			params.DstValidatorAddr = dstValidatorAddr
		}

		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		res, height, err := cliCtx.QueryWithData("custom/staking/redelegations", bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// delegationHandlerFn godoc
// @Tags Staking
// @Summary Query the current delegation between a delegator and a validator
// @Description Query the current delegation between a delegator and a validator
// @ID stakingGetDelegaton
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Param validatorAddr path string true "Bech32 OperatorAddress of validator"
// @Success 200 {object} QueryDelegationResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/delegations/{validatorAddr} [get]
func delegationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryBonds(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegation))
}

// delegatorValidatorsHandlerFn godoc
// @Tags Staking
// @Summary Query all validators that a delegator is bonded to
// @Description Query all validators that a delegator is bonded to
// @ID stakingGetDelegatorValidators
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Success 200 {object} QueryValidatorsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/validators [get]
func delegatorValidatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorValidators))
}

// delegatorValidatorHandlerFn godoc
// @Tags Staking
// @Summary Query a validator that a delegator is bonded to
// @Description Query a validator that a delegator is bonded to
// @ID stakingGetDelegatorValidator
// @Accept  json
// @Produce json
// @Param delegatorAddr path string true "Bech32 AccAddress of Delegator"
// @Param validatorAddr path string true "Bech32 ValAddress of Delegator"
// @Success 200 {object} QueryValidatorResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/delegators/{delegatorAddr}/validators/{validatorAddr} [get]
func delegatorValidatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryDelegator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryDelegatorValidator))
}

// validatorsHandlerFn godoc
// @Tags Staking
// @Summary Get all validator candidates. By default it returns only the bonded validators
// @Description Get all validator candidates. By default it returns only the bonded validators
// @ID stakingGetValidators
// @Accept  json
// @Produce json
// @Param status query string false "The validator bond status. Must be either 'bonded', 'unbonded', or 'unbonding'"
// @Param page query string false "The page number"
// @Param limit query string false "The maximum number of items per page"
// @Success 200 {object} QueryValidatorsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators [get]
func validatorsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		status := r.FormValue("status")
		if status == "" {
			status = sdk.BondStatusBonded
		}

		params := types.NewQueryValidatorsParams(page, limit, status)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidators)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// validatorHandlerFn godoc
// @Tags Staking
// @Summary Query the information from a single validator
// @Description Query the information from a single validator
// @ID stakingGetValidator
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 ValAddress"
// @Success 200 {object} QueryValidatorResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr} [get]
func validatorHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidator))
}

// validatorDelegationsHandlerFn godoc
// @Tags Staking
// @Summary Get the current delegations for the validator
// @Description Get the current delegations for the validator
// @ID stakingGetValidatorDelegations
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 ValAddress"
// @Success 200 {object} QueryDelegationsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr}/delegations [get]
func validatorDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidatorDelegations))
}

// validatorUnbondingDelegationsHandlerFn godoc
// @Tags Staking
// @Summary Get the current unbonding information for the validator
// @Description Get the current unbonding information for the validator
// @ID stakingGetValidatorUnbondingDelegation
// @Accept  json
// @Produce json
// @Param validatorAddr path string true "Bech32 ValAddress"
// @Success 200 {object} QueryUnbondingDelegationsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/validators/{validatorAddr}/unbonding_delegations [get]
func validatorUnbondingDelegationsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return queryValidator(cliCtx, fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryValidatorUnbondingDelegations))
}

// historicalInfoHandlerFn godoc
// @Tags Staking
// @Summary Query historical info at a given height
// @Description Query historical info at a given height
// @ID stakingGetHistoricalInfo
// @Accept  json
// @Produce json
// @Param height path string true "block height"
// @Success 200 {object} QueryHistoricalInfoResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/historical_info/{height} [get]
func historicalInfoHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		heightStr := vars["height"]
		height, err := strconv.ParseInt(heightStr, 10, 64)
		if err != nil || height < 0 {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Must provide non-negative integer for height: %v", err))
			return
		}

		params := types.NewQueryHistoricalInfoParams(height)
		bz, err := cliCtx.Codec.MarshalJSON(params)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryHistoricalInfo)
		res, height, err := cliCtx.QueryWithData(route, bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// poolHandlerFn godoc
// @Tags Staking
// @Summary Get the current state of the staking pool
// @Description Get the current state of the staking pool
// @ID stakingGetPool
// @Accept  json
// @Produce json
// @Success 200 {object} QueryPoolResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/pool [get]
func poolHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryPool)
		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// paramsHandlerFn godoc
// @Tags Staking
// @Summary Get the current staking parameter values
// @Description Get the current staking parameter values
// @ID stakingGetParams
// @Accept  json
// @Produce json
// @Success 200 {object} QueryStakingParamsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /staking/parameters [get]
func paramsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParameters)
		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
