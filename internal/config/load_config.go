package config

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

var configInstance *Config

func LoadConfig() Config {
	vpr := viper.New()
	vpr.AutomaticEnv() // real env vars override values from file

	// Candidate locations for .env (first found wins)
	var candidates []string

	// 1) next to the binary
	if execPath, err := os.Executable(); err == nil {
		binDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(binDir, ".env"))

		// 2) parent of binary dir (e.g., project root if binary in ./temp)
		candidates = append(candidates, filepath.Join(binDir, "..", ".env"))
	}

	// 3) CWD
	candidates = append(candidates, ".env")

	// Try to read the first existing .env with Viper
	var readOK bool
	for _, path := range candidates {
		vpr.SetConfigFile(filepath.Clean(path))
		vpr.SetConfigType("env") // tell Viper this is dotenv-style
		if err := vpr.ReadInConfig(); err == nil {
			readOK = true
			break
		}
	}
	if !readOK {
		// Not fatal: you might rely entirely on OS env vars
		slog.InfoContext(context.Background(), "No .env file found in candidate locations; relying on environment variables")
	}

	configInstance = &Config{
		CliApp: CliApp{
			App: App{
				Name:     coalesce(vpr.GetString("APP_CLI_NAME"), "gsync"),
				Version:  coalesce(vpr.GetString("APP_CLI_VERSION"), "dev"),
				LogLevel: coalesce(vpr.GetString("APP_CLI_LOG_LEVEL"), "info"),
			},
			CliTimeout: coalesce(vpr.GetDuration("APP_CLI_TIMEOUT"), 1*time.Minute),
		},
		Db: DbConfig{
			Address:               coalesce(vpr.GetString("DB_ADDRESS"), "localhost:27017"),
			Name:                  coalesce(vpr.GetString("DB_NAME"), "example"),
			Username:              coalesce(vpr.GetString("DB_USERNAME"), "admin"),
			Password:              coalesce(vpr.GetString("DB_PASSWORD"), "admin_password"),
			Option:                coalesce(vpr.GetString("DB_OPTION"), "?authSource=admin&readPreference=secondaryPreferred"),
			MaxOpenConnections:    coalesce(vpr.GetInt("DB_MAX_OPEN_CONNECTIONS"), 10),
			MaxIdleConnections:    coalesce(vpr.GetInt("DB_MAX_IDLE_CONNECTIONS"), 2),
			MaxConnectionLifetime: coalesce(vpr.GetDuration("DB_MAX_CONNECTION_LIFETIME"), 60*time.Second),
			MaxConnectionIdleTime: coalesce(vpr.GetDuration("DB_MAX_CONNECTION_IDLE_TIME"), 20*time.Second),
		},
		GsheetDataTransaksi: GsheetDataTransaksi{
			Id:               vpr.GetString("GSHEET_DATA_TRANSAKSI_ID"),
			CredentialPath:   vpr.GetString("GSHEET_CREDENTIAL_PATH"),
			TransactionRange: coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_TRANSACTION_RANGE"), "Transaction!B5:H"),
			TransferRange:    coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_TRANSFER_RANGE"), "Transaction!B5:G"),
			CreditRange:      coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_CREDIT_RANGE"), "Liquid Assets!O5:R27"),
			DebtRange:        coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_DEBT_RANGE"), "Liquid Assets!U5:X27"),
			AccountRange:     coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_ACCOUNT_RANGE"), "Liquid Assets!B5:M28"),
			CategoryRange:    coalesce(vpr.GetString("GSHEET_DATA_TRANSAKSI_CATEGORY_RANGE"), "Categories & Budget Mapping!B5:G98"),
		},
		GsheetRekapKeuangan: GsheetRekapKeuangan{
			Id:                  vpr.GetString("GSHEET_REKAP_KEUANGAN_ID"),
			CredentialPath:      vpr.GetString("GSHEET_CREDENTIAL_PATH"),
			ResumeStartColRange: coalesce(vpr.GetString("GSHEET_REKAP_KEUANGAN_RESUME_START_COL_RANGE"), "Resume!B3"),
		},
	}

	return *configInstance
}

func coalesce[T comparable](first, second T) T {
	var zero T
	if first != zero {
		return first
	}
	return second
}
