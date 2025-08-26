package commands

import (
	"fmt"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/spf13/cobra"
)

var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Sync Transaction sheet to db",
	Long:  `Sync Transaction sheet to db`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		authData := entity.AuthData{Name: "cli"}
		ctx := util.SetAuthData(cmd.Context(), authData)
		err = transactionApp.Sync(ctx)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error on transactionApp.Sync:", err)
			os.Exit(2)
		}
		return
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
