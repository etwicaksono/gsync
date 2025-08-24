package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var transactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Sync Transaction sheet to db",
	Long:  `Sync Transaction sheet to db`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		err = transactionApp.Sync(cmd.Context())
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
