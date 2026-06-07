package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/entity"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/util"
	"github.com/spf13/cobra"
)

const (
	resumeDetail = "resume-detail"
	resumeRecap  = "resume-recap"
)

var gsheetFinancePushArgs = []string{resumeDetail, resumeRecap}
var resumeStartColRange, startDateStr, endDateStr string
var gsheetFinancePushCmd = &cobra.Command{
	Use:       "push {args}",
	Short:     "Push finance data from db to sheet",
	Long:      `Push finance data from db to sheet`,
	ValidArgs: append(gsheetFinancePushArgs, "all"),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(2)
		}

		target := args[0]

		authData := entity.AuthData{Name: "cli"}
		ctx := util.SetAuthData(cmd.Context(), authData)

		startDate, err := time.Parse(time.DateOnly, startDateStr)
		if err != nil {
			return fmt.Errorf("failed to parse start date: %s", err.Error())
		}

		endDate, err := time.Parse(time.DateOnly, endDateStr)
		if err != nil {
			return fmt.Errorf("failed to parse end date: %s", err.Error())
		}

		// Define the callable signature and registry
		type pushFn struct {
			fn     func(context.Context) error
			nameFn string
		}
		registry := map[string]pushFn{
			resumeDetail: {
				fn: func(ctx context.Context) error {
					return gsheetFinanceApp.PushResumeDetail(ctx, resumeStartColRange, entity.VResumeDetailListFilter{
						StartDate: startDate,
						EndDate:   endDate,
					})
				},
				nameFn: resumeDetail,
			},
			resumeRecap: {
				fn: func(ctx context.Context) error {
					return gsheetFinanceApp.PushResumeRecap(ctx, resumeStartColRange, entity.VResumeDetailListFilter{
						StartDate: startDate,
						EndDate:   endDate,
					})
				},
				nameFn: resumeRecap,
			},
		}

		// Helper to run one target with nice error context
		runOne := func(name string) error {
			pushFn, ok := registry[name]
			if !ok {
				return fmt.Errorf("invalid target %q; available: %s", name, cmd.ValidArgs)
			}
			fmt.Printf("Start pushing %s from %s to %s...\n", pushFn.nameFn, startDateStr, endDateStr)
			if err := pushFn.fn(ctx); err != nil {
				return fmt.Errorf("error in %s: %w", name, err)
			}
			fmt.Printf("Success pushing %s from %s to %s.\n", pushFn.nameFn, startDateStr, endDateStr)
			return nil
		}

		// Handle "all" or a single target
		if target == "all" {
			for _, name := range gsheetFinancePushArgs {
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
	gsheetFinanceCmd.AddCommand(gsheetFinancePushCmd)

	gsheetFinancePushCmd.Use = fmt.Sprintf(fmt.Sprint("push ", gsheetFinancePushCmd.ValidArgs))
	gsheetFinancePushCmd.Flags().StringVarP(&resumeStartColRange, "start-col", "c", "", "Start column range, e.g. ExampleSheet!A3:C3")
	gsheetFinancePushCmd.Flags().StringVarP(&startDateStr, "start-date", "s", "", "Start date, e.g. 2021-01-01")
	gsheetFinancePushCmd.Flags().StringVarP(&endDateStr, "end-date", "e", "", "End date, e.g. 2021-01-31")
	_ = gsheetFinancePushCmd.MarkFlagRequired("start-col")
	_ = gsheetFinancePushCmd.MarkFlagRequired("start-date")
	_ = gsheetFinancePushCmd.MarkFlagRequired("end-date")
}
