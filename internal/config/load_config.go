package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	"github.com/spf13/viper"
)

var configInstance *Config

func LoadConfig() Config {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to get current working directory", slog.String(constants.Error, err.Error()))
		panic(err.Error())
	}

	// Root folder of this project
	projectRoot := wd
	vpr := viper.New()

	vpr.AddConfigPath(projectRoot)
	vpr.SetConfigFile(fmt.Sprintf("%s/.env", projectRoot))
	vpr.AutomaticEnv()

	err = vpr.ReadInConfig()
	if err != nil {
		slog.ErrorContext(context.Background(), "Failed to read config file", slog.String(constants.Error, err.Error()))
		panic(err.Error())
	}

	configInstance = &Config{
		CliApp: CliApp{
			CliTimeout: coalesce(vpr.GetDuration("APP_CLI_TIMEOUT"), 1*time.Minute),
		},
		Gsheet: GsheetConfig{
			Id:               vpr.GetString("GSHEET_ID"),
			CredentialPath:   vpr.GetString("GSHEET_CREDENTIAL_PATH"),
			TransactionRange: coalesce(vpr.GetString("GSHEET_TRANSACTION_RANGE"), "Transaction!B5:H"),
			TransferRange:    coalesce(vpr.GetString("GSHEET_TRANSFER_RANGE"), "Transaction!B5:G"),
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
