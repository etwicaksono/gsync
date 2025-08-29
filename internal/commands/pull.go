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
		type pullFn struct {
			fn     func(context.Context, bool) error
			nameFn string
		}
		registry := map[string]pullFn{
			"transaction": {
				fn:     func(ctx context.Context, force bool) error { return transactionApp.PullTransaction(ctx, force) },
				nameFn: "transaction",
			},
			"transfer": {
				fn:     func(ctx context.Context, force bool) error { return transferApp.PullTransfer(ctx, force) },
				nameFn: "transfer",
			},
			"credit": {
				fn:     func(ctx context.Context, force bool) error { return creditApp.PullCredit(ctx, force) },
				nameFn: "credit",
			},
			"debt": {
				fn:     func(ctx context.Context, force bool) error { return debtApp.PullDebt(ctx, force) },
				nameFn: "debt",
			},
			"account": {
				fn:     func(ctx context.Context, force bool) error { return accountApp.PullAccount(ctx, force) },
				nameFn: "account",
			},
			"category": {
				fn:     func(ctx context.Context, force bool) error { return categoryApp.PullCategory(ctx, force) },
				nameFn: "category",
			},
		}

		// Helper to run one target with nice error context
		runOne := func(name string) error {
			pullFn, ok := registry[name]
			if !ok {
				return fmt.Errorf("invalid target %q; available: transaction, transfer, credit, debt, account, category, all", name)
			}
			fmt.Printf("Start pulling %s...\n", pullFn.nameFn)
			if err := pullFn.fn(ctx, forcePull); err != nil {
				return fmt.Errorf("error in %s: %w", name, err)
			}
			fmt.Printf("Success pulling %s.\n", pullFn.nameFn)
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
	pullCmd.ValidArgs = []string{"transaction", "transfer", "credit", "debt", "account", "category", "all"}
	pullCmd.Flags().BoolVarP(&forcePull, "force", "f", false, "truncate data first")
}
