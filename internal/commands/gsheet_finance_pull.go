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

const (
	transaction = "transaction"
	transfer    = "transfer"
	credit      = "credit"
	debt        = "debt"
	account     = "account"
	category    = "category"
)

var gsheetFinancePullArgs = []string{transaction, transfer, credit, debt, account, category}
var forcegsheetFinancePull bool
var gsheetFinancePullCmd = &cobra.Command{
	Use:       "pull {args}",
	Short:     "Pull finance data from sheet to db",
	Long:      `Pull finance data from sheet to db`,
	ValidArgs: append(gsheetFinancePullArgs, "all"),
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
			transaction: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullTransaction(ctx, force) },
				nameFn: transaction,
			},
			transfer: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullTransfer(ctx, force) },
				nameFn: transfer,
			},
			credit: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullCredit(ctx, force) },
				nameFn: credit,
			},
			debt: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullDebt(ctx, force) },
				nameFn: debt,
			},
			account: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullAccount(ctx, force) },
				nameFn: account,
			},
			category: {
				fn:     func(ctx context.Context, force bool) error { return gsheetFinanceApp.PullCategory(ctx, force) },
				nameFn: category,
			},
		}

		// Helper to run one target with nice error context
		runOne := func(ctx context.Context, name string) error {
			pullFn, ok := registry[name]
			if !ok {
				return fmt.Errorf("invalid target %q; available: %s", name, cmd.ValidArgs)
			}
			fmt.Printf("Start pulling %s...\n", pullFn.nameFn)
			if err := pullFn.fn(ctx, forcegsheetFinancePull); err != nil {
				return fmt.Errorf("error in %s: %w", name, err)
			}
			fmt.Printf("Success pulling %s.\n", pullFn.nameFn)
			return nil
		}

		// Handle "all" or a single target
		if target == "all" {
			// Create an errgroup with the context
			g, ctx := errgroup.WithContext(ctx)
			for _, name := range gsheetFinancePullArgs {
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
	gsheetFinanceCmd.AddCommand(gsheetFinancePullCmd)

	gsheetFinancePullCmd.Use = fmt.Sprintf(fmt.Sprint("pull ", gsheetFinancePullCmd.ValidArgs))
	gsheetFinancePullCmd.ValidArgs = []string{"transaction", "transfer", "credit", "debt", "account", "category", "all"}
	gsheetFinancePullCmd.Flags().BoolVarP(&forcegsheetFinancePull, "force", "f", false, "truncate data first")
}
