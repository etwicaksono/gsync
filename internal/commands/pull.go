package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/spf13/cobra"
)

var forcePull bool
var pullCmd = &cobra.Command{
	Use:   "pull {transaction|credit|debt|transfer|account|category|all}",
	Short: "Pull data from sheet to db",
	Long:  `Pull data from sheet to db`,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(2)
		}

		target := args[0]

		authData := entity.AuthData{Name: "cli"}
		ctx := util.SetAuthData(cmd.Context(), authData)

		// Define the callable signature and registry
		type pullFn func(context.Context, bool) error
		registry := map[string]pullFn{
			"transaction": func(ctx context.Context, force bool) error { return transactionApp.PullTransaction(ctx, force) },
			"transfer":    func(ctx context.Context, force bool) error { return transferApp.PullTransfer(ctx, force) },
			"credit":      func(ctx context.Context, force bool) error { return creditApp.PullCredit(ctx, force) },
			"debt":        func(ctx context.Context, force bool) error { return debtApp.PullDebt(ctx, force) },
			"account":     func(ctx context.Context, force bool) error { return accountApp.PullAccount(ctx, force) },
			"category":    func(ctx context.Context, force bool) error { return categoryApp.PullCategory(ctx, force) },
		}

		// Helper to run one target with nice error context
		runOne := func(name string) error {
			fn, ok := registry[name]
			if !ok {
				return fmt.Errorf("invalid target %q; available: transaction, transfer, credit, debt, category, all", name)
			}
			if err := fn(ctx, forcePull); err != nil {
				return fmt.Errorf("error in %s: %w", name, err)
			}
			return nil
		}

		// Handle "all" or a single target
		if target == "all" {
			order := []string{"transaction", "transfer", "credit", "debt", "account", "category"}
			for _, name := range order {
				if err := runOne(name); err != nil {
					return err
				}
			}
			return nil
		}

		return runOne(target)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.ValidArgs = []string{"transaction", "transfer", "credit", "debt", "category", "all"}
	pullCmd.Flags().BoolVarP(&forcePull, "force", "f", false, "truncate data first")
}
