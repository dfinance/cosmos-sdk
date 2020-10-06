// nolint
package cli

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/distribution/client/common"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var (
	flagOnlyFromValidator = "only-from-validator"
	flagIsValidator       = "is-validator"
	flagCommission        = "commission"
	flagMaxMessagesPerTx  = "max-msgs"
)

const (
	MaxMessagesPerTxDefault = 5
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(storeKey string, cdc *codec.Codec) *cobra.Command {
	distTxCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Distribution transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	distTxCmd.AddCommand(flags.PostCommands(
		GetCmdWithdrawRewards(cdc),
		GetCmdSetWithdrawAddr(cdc),
		GetCmdWithdrawAllRewards(cdc, storeKey),
		GetCmdLockValidatorRewards(cdc),
		GetCmdDisableValidatorLockedRewardsAutoRenewal(cdc),
		GetCmdFundPublicTreasuryPool(cdc),
		GetCmdFoundationPoolWithdraw(cdc),
		GetChangeFoundationAllocationRatioTxCmd(cdc),
		GetChangeMintStakingTotalSupplyShiftParamTxCmd(cdc),
	)...)

	return distTxCmd
}

type generateOrBroadcastFunc func(context.CLIContext, auth.TxBuilder, []sdk.Msg) error

func splitAndApply(
	generateOrBroadcast generateOrBroadcastFunc,
	cliCtx context.CLIContext,
	txBldr auth.TxBuilder,
	msgs []sdk.Msg,
	chunkSize int,
) error {

	if chunkSize == 0 {
		return generateOrBroadcast(cliCtx, txBldr, msgs)
	}

	// split messages into slices of length chunkSize
	totalMessages := len(msgs)
	for i := 0; i < len(msgs); i += chunkSize {

		sliceEnd := i + chunkSize
		if sliceEnd > totalMessages {
			sliceEnd = totalMessages
		}

		msgChunk := msgs[i:sliceEnd]
		if err := generateOrBroadcast(cliCtx, txBldr, msgChunk); err != nil {
			return err
		}
	}

	return nil
}

// command to withdraw rewards
func GetCmdWithdrawRewards(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-rewards [validator-addr]",
		Short: "Withdraw rewards from a given delegation address, and optionally withdraw validator commission if the delegation address given is a validator operator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw rewards from a given delegation address,
and optionally withdraw validator commission if the delegation address given is a validator operator.

Example:
$ %s tx distribution withdraw-rewards cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey
$ %s tx distribution withdraw-rewards cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj --from mykey --commission
`,
				version.ClientName, version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			delAddr := cliCtx.GetFromAddress()
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msgs := []sdk.Msg{types.NewMsgWithdrawDelegatorReward(delAddr, valAddr)}
			if viper.GetBool(flagCommission) {
				msgs = append(msgs, types.NewMsgWithdrawValidatorCommission(valAddr))
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, msgs)
		},
	}
	cmd.Flags().Bool(flagCommission, false, "also withdraw validator's commission")
	return cmd
}

// command to withdraw all rewards
func GetCmdWithdrawAllRewards(cdc *codec.Codec, queryRoute string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-all-rewards",
		Short: "withdraw all delegations rewards for a delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Withdraw all rewards for a single delegator.

Example:
$ %s tx distribution withdraw-all-rewards --from mykey
`,
				version.ClientName,
			),
		),
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			delAddr := cliCtx.GetFromAddress()

			// The transaction cannot be generated offline since it requires a query
			// to get all the validators.
			if cliCtx.GenerateOnly {
				return fmt.Errorf("command disabled with the provided flag: %s", flags.FlagGenerateOnly)
			}

			msgs, err := common.WithdrawAllDelegatorRewards(cliCtx, queryRoute, delAddr)
			if err != nil {
				return err
			}

			chunkSize := viper.GetInt(flagMaxMessagesPerTx)
			return splitAndApply(utils.GenerateOrBroadcastMsgs, cliCtx, txBldr, msgs, chunkSize)
		},
	}

	cmd.Flags().Int(flagMaxMessagesPerTx, MaxMessagesPerTxDefault, "Limit the number of messages per tx (0 for unlimited)")
	return cmd
}

// command to replace a delegator's withdrawal address
func GetCmdSetWithdrawAddr(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "set-withdraw-addr [withdraw-addr]",
		Short: "change the default withdraw address for rewards associated with an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set the withdraw address for rewards associated with a delegator address.

Example:
$ %s tx distribution set-withdraw-addr cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p --from mykey
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			delAddr := cliCtx.GetFromAddress()
			withdrawAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgSetWithdrawAddress(delAddr, withdrawAddr)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdFundPublicTreasuryPool returns a command implementation that supports directly funding the public treasury pool.
func GetCmdFundPublicTreasuryPool(cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "fund-public-treasury-pool [amount]",
		Args:  cobra.ExactArgs(1),
		Short: "Funds the public treasury pool with the specified amount",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Funds the public treasury pool with the specified amount

Example:
$ %s tx distribution fund-public-treasury-pool 100uatom --from mykey
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			depositorAddr := cliCtx.GetFromAddress()
			amount, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundPublicTreasuryPool(amount, depositorAddr)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}
}

// GetCmdFoundationPoolWithdraw implements the command to transfer funds from the foundation-pool to wallet/other pool.
func GetCmdFoundationPoolWithdraw(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "foundation-pool-withdraw [amount] [recipient]",
		Args:  cobra.ExactArgs(2),
		Short: "Transfers specified amount of foundation pool funds to the recipient (wallet address / other pool name)",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Transfers specified amount of foundation pool funds to the recipient (wallet address / other pool name)

Example:
$ %s tx distribution foundation-pool-withdraw 100uatom cosmos1s5afhd6gxevu37mkqcvvsj8qeylhn0rz46zdlq --from nomineeKey
$ %s tx distribution foundation-pool-withdraw 100uatom [LiquidityProvidersPool|PublicTreasuryPool|HARP] --from nomineeKey
`,
				version.ClientName, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			nomineeAddr := cliCtx.GetFromAddress()

			amount, err := sdk.ParseCoins(args[0])
			if err != nil {
				return err
			}

			recipientAddr, recipientPoolName := sdk.AccAddress{}, types.RewardPoolName("")
			if poolName := types.RewardPoolName(args[1]); poolName.IsValid() {
				recipientPoolName = poolName
			} else {
				accAddr, err := sdk.AccAddressFromBech32(args[1])
				if err != nil {
					return err
				}
				recipientAddr = accAddr
			}

			msg := types.NewMsgWithdrawFoundationPool(nomineeAddr, recipientAddr, recipientPoolName, amount)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdLockValidatorRewards implements the command to lock validators rewards.
func GetCmdLockValidatorRewards(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-rewards",
		Short: "Lock delegators rewards withdraw",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Lock delegators rewards withdraw for params defined period of time.
Lock increases validator's' rewards distribution power.
Tx target validator is defined by the '--from' argument (only the operator can emit this operation).
Lock is auto-renewed by default.

Example:
$ %s tx distribution lock-rewards --from=<key_or_address>
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			from := cliCtx.GetFromAddress()

			msg := types.NewMsgLockValidatorRewards(from, sdk.ValAddress(from))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdDisableValidatorLockedRewardsAutoRenewal implements the command to disable validators locked rewards auto-renewal.
func GetCmdDisableValidatorLockedRewardsAutoRenewal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disable-auto-renewal",
		Short: "Disable validator's locked rewards auto-renewal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Disable validator's locked rewards auto-renewal.

Example:
$ %s tx distribution disable-auto-renewal --from=<key_or_address>
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			from := cliCtx.GetFromAddress()

			msg := types.NewMsgDisableLockedRewardsAutoRenewal(from, sdk.ValAddress(from))
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdSubmitPublicTreasurySpendProposal implements the command to submit a public-treasury-pool-spend proposal.
func GetCmdSubmitPublicTreasurySpendProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-treasury-pool-spend [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a public treasury pool spend proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a public treasury pool spend proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal public-treasury-pool-spend <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "PublicTreasury Pool Spend",
  "description": "Pay me some Atoms!",
  "recipient": "cosmos1s5afhd6gxevu37mkqcvvsj8qeylhn0rz46zdlq",
  "amount": [
    {
      "denom": "stake",
      "amount": "10000"
    }
  ],
  "deposit": [
    {
      "denom": "stake",
      "amount": "10000"
    }
  ]
}
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			proposal, err := ParsePublicTreasuryPoolSpendProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewPublicTreasuryPoolSpendProposal(proposal.Title, proposal.Description, proposal.Recipient, proposal.Amount)

			msg := gov.NewMsgSubmitProposal(content, proposal.Deposit, from)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetCmdSubmitTaxParamsUpdateProposal implements the command to submit a tax-params-update proposal.
func GetCmdSubmitTaxParamsUpdateProposal(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tax-params-update [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a tax params update proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a tax params update proposal along with an initial deposit.
The proposal details must be supplied via a JSON file.

Example:
$ %s tx gov submit-proposal tax-params-update <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "title": "TaxParams Update",
  "description": "Lower the PublicTreasury tax",
  "validators_pool_tax": "0.45",
  "liquidity_providers_pool_tax": "0.45",
  "public_treasury_pool_tax": "0.10",
  "harp_tax": "0.0",
  "deposit": [
    {
      "denom": "stake",
      "amount": "10000"
    }
  ]
}
`, version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(utils.GetTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			proposal, err := ParseTaxParamsUpdateProposalJSON(cdc, args[0])
			if err != nil {
				return err
			}

			from := cliCtx.GetFromAddress()
			content := types.NewTaxParamsUpdateProposal(
				proposal.Title, proposal.Description,
				proposal.ValidatorsPoolTax,
				proposal.LiquidityProvidersPoolTax,
				proposal.PublicTreasuryPoolTax,
				proposal.HARPTax,
			)

			msg := gov.NewMsgSubmitProposal(content, proposal.Deposit, from)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetChangeFoundationAllocationRatioTxCmd will create a send tx and sign it with the given key.
func GetChangeFoundationAllocationRatioTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-foundation-allocation-ratio [ratio]",
		Short: "Change FoundationAllocationRatio param",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// parse ratio
			ratio, err := sdk.NewDecFromStr(args[0])
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgSetFoundationAllocationRatio(cliCtx.GetFromAddress(), ratio)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}

// GetChangeMintStakingTotalSupplyShiftParamTxCmd will create a send tx and sign it with the given key.
func GetChangeMintStakingTotalSupplyShiftParamTxCmd(cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-mint-stakingtotalsupplyshift-value [value]",
		Short: "Change mint module StakingTotalSupplyShift param",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inBuf := bufio.NewReader(cmd.InOrStdin())
			txBldr := auth.NewTxBuilderFromCLI(inBuf).WithTxEncoder(auth.DefaultTxEncoder(cdc))
			cliCtx := context.NewCLIContextWithInput(inBuf).WithCodec(cdc)

			// parse value
			value, ok := sdk.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("invalid Int value")
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := types.NewMsgSetStakingTotalSupplyShift(cliCtx.GetFromAddress(), value)
			return utils.GenerateOrBroadcastMsgs(cliCtx, txBldr, []sdk.Msg{msg})
		},
	}

	return cmd
}
