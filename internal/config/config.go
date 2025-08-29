package config

import "time"

type Config struct {
	CliApp CliApp
	Db     DbConfig
	Gsheet GsheetConfig
}

type App struct {
	Name     string
	Version  string
	LogLevel string
}

type CliApp struct {
	App
	CliTimeout time.Duration
}

type GsheetConfig struct {
	Id               string
	CredentialPath   string
	TransactionRange string
	TransferRange    string
	CreditRange      string
	DebtRange        string
	AccountRange     string
	CategoryRange    string
}

type DbConfig struct {
	Address               string
	Name                  string
	Username              string
	Password              string
	Option                string
	MaxOpenConnections    int
	MaxIdleConnections    int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdletime time.Duration
}
