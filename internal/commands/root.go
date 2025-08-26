package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/transaction_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/app/transfer_app"
	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/db/transaction_db"
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/infrastructure"
	"github.com/etwicaksono/data-transaksi-service/internal/injector"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/app"
	"github.com/etwicaksono/data-transaksi-service/internal/ports/secondary/db"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var (
	version        = "dev"
	cfg            config.Config
	transactionApp app.TransactionAppInterface
	transferApp    app.TransferAppInterface
	transactionDb  db.TransactionDbInterface
	dbClient       *gorm.DB
)

var rootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "CLI that syncs google sheets and database",
	Long:  `CLI that syncs google sheets and database`,
	// Run omitted so help is shown if no subcommand is provided
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().NFlag() == 0 {
			_ = cmd.Help()
		}

		showVersion, _ := cmd.Flags().GetBool("version")
		if showVersion {
			fmt.Printf("version: %s\n", version)
			os.Exit(0)
		}
	},
}

// Execute runs the root command
func Execute() {
	var err error
	ctx := context.Background()
	cfg = config.LoadConfig()

	logger, err := injector.LoggerInit(cfg.CliApp.App)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	gormAdapter := infrastructure.NewGormDb(ctx, cfg, logger)
	err = gormAdapter.Connect()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gormAdapter.Disconnect()
	dbClient = gormAdapter.GetClient()

	transactionDb = transaction_db.NewTransactionDb(cfg)
	transactionApp = transaction_app.NewTransactionApp(cfg, transactionDb, dbClient)
	transferApp = transfer_app.NewTransferApp(cfg, transactionDb, dbClient)

	if err = rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}
