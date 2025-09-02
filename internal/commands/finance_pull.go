package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/spf13/cobra"

	"golang.org/x/sync/errgroup"
)

var forceFinancePull bool
var financePullCmd = &cobra.Command{
	Use:   "pull {transaction|credit|debt|transfer|account|category|all}",
	Short: "Pull finance data from sheet to db",
	Long:  `Pull finance data from sheet to db`,
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
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullTransaction(ctx, force) },
				nameFn: "transaction",
			},
			"transfer": {
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullTransfer(ctx, force) },
				nameFn: "transfer",
			},
			"credit": {
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullCredit(ctx, force) },
				nameFn: "credit",
			},
			"debt": {
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullDebt(ctx, force) },
				nameFn: "debt",
			},
			"account": {
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullAccount(ctx, force) },
				nameFn: "account",
			},
			"category": {
				fn:     func(ctx context.Context, force bool) error { return financeApp.PullCategory(ctx, force) },
				nameFn: "category",
			},
		}

		// Helper to run one target with nice error context
		runOne := func(ctx context.Context, name string) error {
			pullFn, ok := registry[name]
			if !ok {
				return fmt.Errorf("invalid target %q; available: transaction, transfer, credit, debt, account, category, all", name)
			}
			fmt.Printf("Start pulling %s...\n", pullFn.nameFn)
			if err := pullFn.fn(ctx, forceFinancePull); err != nil {
				return fmt.Errorf("error in %s: %w", name, err)
			}
			fmt.Printf("Success pulling %s.\n", pullFn.nameFn)
			return nil
		}

		// Handle "all" or a single target
		if target == "all" {
			order := []string{"transaction", "transfer", "credit", "debt", "account", "category"}
			// Create an errgroup with the context
			g, ctx := errgroup.WithContext(ctx)
			for _, name := range order {
				g.Go(func() error { return runOne(ctx, name) })
			}
			if err := g.Wait(); err != nil {
				return fmt.Errorf("error in all: %w", err)
			}
			return nil
		}

		return runOne(ctx, target)
	},
}

func init() {
	financeCmd.AddCommand(financePullCmd)
	financePullCmd.ValidArgs = []string{"transaction", "transfer", "credit", "debt", "account", "category", "all"}
	financePullCmd.Flags().BoolVarP(&forceFinancePull, "force", "f", false, "truncate data first")
}
