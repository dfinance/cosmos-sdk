package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	FlagStatus = "status"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	distQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the distribution module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distQueryCmd.AddCommand(flags.GetCommands(
		GetCmdQueryParams(queryRoute, cdc),
		GetCmdQueryValidatorOutstandingRewards(queryRoute, cdc),
		GetCmdQueryValidatorCommission(queryRoute, cdc),
		GetCmdQueryValidatorSlashes(queryRoute, cdc),
		GetCmdQueryDelegatorRewards(queryRoute, cdc),
		GetCmdQueryPool(queryRoute, cdc),
		GetCmdQueryValidatorExtended(queryRoute, cdc),
		GetCmdQueryValidatorsExtended(queryRoute, cdc),
	)...)

	return distQueryCmd
}

// GetCmdQueryParams implements the query params command.
func GetCmdQueryParams(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query distribution params",
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
			res, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			if err := cdc.UnmarshalJSON(res, &params); err != nil {
				return fmt.Errorf("failed to unmarshal params: %w", err)
			}

			return cliCtx.PrintOutput(params)
		},
	}
}

// GetCmdQueryValidatorOutstandingRewards implements the query validator outstanding rewards command.
func GetCmdQueryValidatorOutstandingRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator-outstanding-rewards [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution outstanding (un-withdrawn) rewards for a validator and all their delegations",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query distribution outstanding (un-withdrawn) rewards
for a validator and all their delegations.

Example:
$ %s query distribution validator-outstanding-rewards cosmosvaloper1lwjmdnks33xwnmfayc64ycprww49n33mtm92ne
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryValidatorOutstandingRewardsParams(valAddr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			resp, _, err := cliCtx.QueryWithData(
				fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorOutstandingRewards),
				bz,
			)
			if err != nil {
				return err
			}

			var outstandingRewards types.ValidatorOutstandingRewards
			if err := cdc.UnmarshalJSON(resp, &outstandingRewards); err != nil {
				return err
			}

			return cliCtx.PrintOutput(outstandingRewards)
		},
	}
}

// GetCmdQueryValidatorCommission implements the query validator commission command.
func GetCmdQueryValidatorCommission(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "commission [validator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query distribution validator commission",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query validator commission rewards from delegators to that validator.

Example:
$ %s query distribution commission cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := common.QueryValidatorCommission(cliCtx, queryRoute, validatorAddr)
			if err != nil {
				return err
			}

			var valCom types.ValidatorAccumulatedCommission
			cdc.MustUnmarshalJSON(res, &valCom)
			return cliCtx.PrintOutput(valCom)
		},
	}
}

// GetCmdQueryValidatorSlashes implements the query validator slashes command.
func GetCmdQueryValidatorSlashes(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "slashes [validator] [start-height] [end-height]",
		Args:  cobra.ExactArgs(3),
		Short: "Query distribution validator slashes",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all slashes of a validator for a given block range.

Example:
$ %s query distribution slashes cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj 0 100
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("start-height %s not a valid uint, please input a valid start-height", args[1])
			}

			endHeight, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("end-height %s not a valid uint, please input a valid end-height", args[2])
			}

			params := types.NewQueryValidatorSlashesParams(validatorAddr, startHeight, endHeight)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validator_slashes", queryRoute), bz)
			if err != nil {
				return err
			}

			var slashes types.ValidatorSlashEvents
			cdc.MustUnmarshalJSON(res, &slashes)
			return cliCtx.PrintOutput(slashes)
		},
	}
}

// GetCmdQueryDelegatorRewards implements the query delegator rewards command.
func GetCmdQueryDelegatorRewards(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "rewards [delegator-addr] [<validator-addr>]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query all distribution delegator rewards or rewards from a particular validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all rewards earned by a delegator, optionally restrict to rewards from a single validator.

Example:
$ %s query distribution rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
$ %s query distribution rewards cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			// query for rewards from a particular delegation
			if len(args) == 2 {
				resp, _, err := common.QueryDelegationRewards(cliCtx, queryRoute, args[0], args[1])
				if err != nil {
					return err
				}

				var result sdk.DecCoins
				if err = cdc.UnmarshalJSON(resp, &result); err != nil {
					return fmt.Errorf("failed to unmarshal response: %w", err)
				}

				return cliCtx.PrintOutput(result)
			}

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryDelegatorParams(delegatorAddr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			// query for delegator total rewards
			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorTotalRewards)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var result types.QueryDelegatorTotalRewardsResponse
			if err = cdc.UnmarshalJSON(res, &result); err != nil {
				return fmt.Errorf("failed to unmarshal response: %w", err)
			}

			return cliCtx.PrintOutput(result)
		},
	}
}

// GetCmdQueryPool returns the command for fetching specified pool info.
func GetCmdQueryPool(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool",
		Args:  cobra.ExactArgs(1),
		Short: "Query the amount of coins in the specified pool",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the amount of coins in the specified pool.

Example:
$ %s query distribution pool PublicTreasuryPool
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", queryRoute, types.QueryPool, args[0]), nil)
			if err != nil {
				return err
			}

			var result sdk.DecCoins
			cdc.MustUnmarshalJSON(res, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}

// GetCmdQueryValidatorExtended implements the validatorExtended query command.
func GetCmdQueryValidatorExtended(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator with extended info",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query extended details about an individual validator.

Example:
$ %s query distribution validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := types.NewQueryValidatorParams(addr)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorExtended), bz)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no validator found with address %s", addr)
			}

			var result types.ValidatorResp
			cdc.MustUnmarshalJSON(res, &result)
			return cliCtx.PrintOutput(result)
		},
	}
}

// GetCmdQueryValidatorsExtended implements the query all validatorExtended objects command.
func GetCmdQueryValidatorsExtended(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators with extended info",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query extended details about all validators on a network.

Example:
$ %s query distribution validators --status=Unbonded --page=1 --limit=50
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			statusFilter := viper.GetString(FlagStatus)
			if statusFilter == "" {
				statusFilter = sdk.BondStatusBonded
			}

			params := types.NewQueryValidatorsParams(
				viper.GetInt(flags.FlagPage),
				viper.GetInt(flags.FlagLimit),
				statusFilter,
			)
			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return fmt.Errorf("failed to marshal params: %w", err)
			}

			res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorsExtended), bz)
			if err != nil {
				return err
			}

			var result []types.ValidatorResp
			cdc.MustUnmarshalJSON(res, &result)
			return cliCtx.PrintOutput(result)
		},
	}
	cmd.Flags().String(FlagStatus, "", "(optional) filter validators by validator status, status: Bonded/Unbonding/Unbonded")
	cmd.Flags().Int(flags.FlagPage, 1, "(optional) query a specific page of paginated results")
	cmd.Flags().Int(flags.FlagLimit, 50, "(optional) query number of results returned per page")

	return cmd
}
