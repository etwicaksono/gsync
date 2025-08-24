package config

import "time"

type Config struct {
	CliApp CliApp
	Gsheet GsheetConfig
}

type CliApp struct {
	CliTimeout time.Duration
}

type GsheetConfig struct {
	Id               string
	CredentialPath   string
	TransactionRange string
	TransferRange    string
}
