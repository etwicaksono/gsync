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
			MaxConnectionIdletime: coalesce(vpr.GetDuration("DB_MAX_CONNECTION_IDLE_TIME"), 20*time.Second),
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
