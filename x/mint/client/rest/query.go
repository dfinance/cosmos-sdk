package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/mint/internal/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(
		"/minting/parameters",
		queryParamsHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/minting/inflation",
		queryInflationHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/minting/annual-provisions",
		queryAnnualProvisionsHandlerFn(cliCtx),
	).Methods("GET")

	r.HandleFunc(
		"/minting/minter-extended",
		queryMinterExtendedHandlerFn(cliCtx),
	).Methods("GET")
}

// queryParamsHandlerFn godoc
// @Tags Mint
// @Summary Minting module parameters
// @Description Minting module parameters
// @ID mintingGetParams
// @Accept  json
// @Produce json
// @Success 200 {object} QueryMinterParamsResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/parameters [get]
func queryParamsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParameters)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// queryInflationHandlerFn godoc
// @Tags Mint
// @Summary Current minting inflation value
// @Description Current minting inflation value
// @ID mintingGetInflation
// @Accept  json
// @Produce json
// @Success 200 {object} QueryDecResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/inflation [get]
func queryInflationHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryInflation)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// queryAnnualProvisionsHandlerFn godoc
// @Tags Mint
// @Summary Current minting annual provisions value
// @Description Current minting annual provisions value
// @ID mintingGetAnnualProvisions
// @Accept  json
// @Produce json
// @Success 200 {object} QueryDecResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/annual-provisions [get]
func queryAnnualProvisionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAnnualProvisions)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// queryMinterExtendedHandlerFn godoc
// @Tags Mint
// @Summary Current minting extended state
// @Description Current minting extended state
// @ID mintingGetExtendedState
// @Accept  json
// @Produce json
// @Success 200 {object} QueryMinterExtendedResp
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /minting/minter-extended [get]
func queryMinterExtendedHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryMinterExtended)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, height, err := cliCtx.QueryWithData(route, nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
