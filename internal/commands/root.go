package commands

import (
	"fmt"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	cfg     config.Config
)

var rootCmd = &cobra.Command{
	Use:   "gsync",
	Short: "CLI that syncs google sheets and database",
	Long:  `CLI that syncs google sheets and database`,
	// Run omitted so help is shown if no subcommand is provided
	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		// Load config into the cfg struct for all subcommands
		cfg = config.LoadConfig()
		return nil
	},
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
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "show version")
}
