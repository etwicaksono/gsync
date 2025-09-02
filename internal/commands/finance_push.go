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

var resumeStartColRange, startDateStr, endDateStr string

const (
	resumeDetail = "resume-detail"
	resumeRecap  = "resume-recap"
)

var financePushCmd = &cobra.Command{
	Use:   "push {resume-detail}",
	Short: "Push finance data from db to sheet",
	Long:  `Push finance data from db to sheet`,
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
					return financeApp.PushResumeDetail(ctx, resumeStartColRange, entity.VResumeDetailListFilter{
						StartDate: startDate,
						EndDate:   endDate,
					})
				},
				nameFn: resumeDetail,
			},
			resumeRecap: {
				fn: func(ctx context.Context) error {
					return financeApp.PushResumeRecap(ctx, resumeStartColRange, entity.VResumeDetailListFilter{
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
				return fmt.Errorf("invalid target %q; available: %s", name, resumeDetail, resumeRecap)
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
			order := []string{resumeDetail, resumeRecap}
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
	financeCmd.AddCommand(financePushCmd)
	financePushCmd.ValidArgs = []string{resumeDetail, resumeRecap}

	financePushCmd.Flags().StringVarP(&resumeStartColRange, "start-col", "c", "", "Start column range, e.g. ExampleSheet!A3:C3")
	financePushCmd.Flags().StringVarP(&startDateStr, "start-date", "s", "", "Start date, e.g. 2021-01-01")
	financePushCmd.Flags().StringVarP(&endDateStr, "end-date", "e", "", "End date, e.g. 2021-01-31")
	_ = financePushCmd.MarkFlagRequired("start-col")
	_ = financePushCmd.MarkFlagRequired("start-date")
	_ = financePushCmd.MarkFlagRequired("end-date")
}
