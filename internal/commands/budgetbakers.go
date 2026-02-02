package commands

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/etwicaksono/data-transaksi-service/internal/embedded"
	"github.com/spf13/cobra"
)

var (
	bbSkip  int
	bbLimit int
	bbMode  string
)

var budgetbakersCmd = &cobra.Command{
	Use:     "budgetbakers",
	Short:   "BudgetBakers browser automation commands",
	Long:    `BudgetBakers browser automation commands using Playwright. Requires Node.js to be installed.`,
	Aliases: []string{"bb"},
}

var bbTransactionCmd = &cobra.Command{
	Use:   "transaction",
	Short: "Run transaction automation",
	Long:  `Run BudgetBakers transaction automation to insert transactions from Google Sheets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBudgetBakersScript("budgetbakers-transaction.js")
	},
}

var bbTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Run transfer automation",
	Long:  `Run BudgetBakers transfer automation to insert transfers from Google Sheets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBudgetBakersScript("budgetbakers-transfer.js")
	},
}

var bbSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install Node.js dependencies and Playwright browser",
	Long:  `Extract scripts, install npm dependencies, and install Playwright Chromium browser.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return setupBudgetBakers()
	},
}

func init() {
	rootCmd.AddCommand(budgetbakersCmd)
	budgetbakersCmd.AddCommand(bbTransactionCmd)
	budgetbakersCmd.AddCommand(bbTransferCmd)
	budgetbakersCmd.AddCommand(bbSetupCmd)

	for _, cmd := range []*cobra.Command{bbTransactionCmd, bbTransferCmd} {
		cmd.Flags().IntVarP(&bbSkip, "skip", "s", 0, "Number of records to skip")
		cmd.Flags().IntVarP(&bbLimit, "limit", "l", 0, "Number of records to process (0 = all)")
		cmd.Flags().StringVarP(&bbMode, "mode", "m", "auto", "Mode: 'auto' or 'assist'")
	}
}

func getScriptsDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	scriptsDir := filepath.Join(configDir, "gsync", "budgetbakers")
	return scriptsDir, nil
}

func extractScripts() (string, error) {
	scriptsDir, err := getScriptsDir()
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create scripts directory: %w", err)
	}

	err = fs.WalkDir(embedded.BudgetBakersFS, "budgetbakers", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel("budgetbakers", path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(scriptsDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		content, err := embedded.BudgetBakersFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		if err := os.WriteFile(destPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", destPath, err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to extract scripts: %w", err)
	}

	return scriptsDir, nil
}

func checkNodeInstalled() error {
	cmd := exec.Command("node", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Node.js is not installed or not in PATH. Please install Node.js from https://nodejs.org/")
	}
	return nil
}

func checkDependenciesInstalled(scriptsDir string) bool {
	nodeModules := filepath.Join(scriptsDir, "node_modules")
	info, err := os.Stat(nodeModules)
	return err == nil && info.IsDir()
}

func setupBudgetBakers() error {
	if err := checkNodeInstalled(); err != nil {
		return err
	}

	fmt.Println("Extracting scripts...")
	scriptsDir, err := extractScripts()
	if err != nil {
		return err
	}
	fmt.Printf("Scripts extracted to: %s\n", scriptsDir)

	fmt.Println("Installing npm dependencies...")
	npmCmd := "npm"
	if runtime.GOOS == "windows" {
		npmCmd = "npm.cmd"
	}

	cmd := exec.Command(npmCmd, "install")
	cmd.Dir = scriptsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install npm dependencies: %w", err)
	}

	fmt.Println("Installing Playwright Chromium browser...")
	cmd = exec.Command(npmCmd, "run", "setup")
	cmd.Dir = scriptsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install Playwright browser: %w", err)
	}

	fmt.Println("\nSetup complete! You can now run:")
	fmt.Println("  gsync budgetbakers transaction --skip 0 --limit 10")
	fmt.Println("  gsync budgetbakers transfer --skip 0 --limit 10")
	return nil
}

func runBudgetBakersScript(scriptName string) error {
	// Safety check: confirm if limit=0 (process all records)
	if bbLimit == 0 {
		fmt.Println("WARNING: --limit is 0, which means ALL records will be processed.")
		fmt.Println("This may cause duplicate entries if records were already inserted.")
		fmt.Print("Are you sure you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	if err := checkNodeInstalled(); err != nil {
		return err
	}

	scriptsDir, err := extractScripts()
	if err != nil {
		return err
	}

	if !checkDependenciesInstalled(scriptsDir) {
		fmt.Println("Dependencies not installed. Running setup first...")
		if err := setupBudgetBakers(); err != nil {
			return err
		}
	}

	scriptPath := filepath.Join(scriptsDir, scriptName)

	args := []string{scriptPath}
	args = append(args, "--skip", fmt.Sprintf("%d", bbSkip))
	args = append(args, "--limit", fmt.Sprintf("%d", bbLimit))
	args = append(args, "--mode", bbMode)

	cmd := exec.Command("node", args...)
	cmd.Dir = scriptsDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("APPS_SCRIPT_URL=%s", cfg.AppsScript.URL))
	cmd.Env = append(cmd.Env, fmt.Sprintf("BB_SESSION_PATH=%s", filepath.Join(scriptsDir, ".budgetbakers-session")))

	return cmd.Run()
}

