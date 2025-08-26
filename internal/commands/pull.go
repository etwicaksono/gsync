package commands

import (
	"fmt"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/spf13/cobra"
)

var forcePull bool
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull data from sheet to db",
	Long:  `Pull data from sheet to db`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			_, _ = fmt.Fprintln(os.Stderr, "target is required, available are: 'transaction', 'credit', 'debt', 'transfer', and 'all'")
			os.Exit(2)
		}

		authData := entity.AuthData{Name: "cli"}
		ctx := util.SetAuthData(cmd.Context(), authData)

		var funcCall string
		switch args[0] {
		case "transaction":
			{
				err = transactionApp.PullTransaction(ctx, forcePull)
				funcCall = "transactionApp.PullTransaction"
			}
		case "transfer":
			{
				err = transferApp.PullTransfer(ctx, forcePull)
				funcCall = "transferApp.PullTransfer"
			}
		default:
			{
				_, _ = fmt.Fprintln(os.Stderr, "invalid argument, available are: 'transaction', 'credit', 'debt', 'transfer', and 'all'")
				os.Exit(2)
			}
		}
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf("error on %s:", funcCall), err)
			os.Exit(2)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().BoolVarP(&forcePull, "force", "f", false, "truncate data first")
}
