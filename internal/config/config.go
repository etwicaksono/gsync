package config

import "time"

type Config struct {
	CliApp              CliApp
	Db                  DbConfig
	GsheetDataTransaksi GsheetDataTransaksi
	GsheetRekapKeuangan GsheetRekapKeuangan
	AppsScript          AppsScriptConfig
	BudgetMateApi       BudgetMateApiConfig
	BudgetBakers        BudgetBakersConfig
}

type BudgetBakersConfig struct {
	APIToken string
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

type GsheetDataTransaksi struct {
	Id               string
	CredentialPath   string
	TransactionRange string
	TransferRange    string
	CreditRange      string
	DebtRange        string
	AccountRange     string
	CategoryRange    string
}

type GsheetRekapKeuangan struct {
	Id                  string
	CredentialPath      string
	ResumeStartColRange string
}

type AppsScriptConfig struct {
	URL string
}

type BudgetMateApiConfig struct {
	URL      string
	Email    string
	Password string
}

type DbConfig struct {
	Address               string
	Protocol              string
	Name                  string
	Username              string
	Password              string
	Option                string
	MaxOpenConnections    int
	MaxIdleConnections    int
	MaxConnectionLifetime time.Duration
	MaxConnectionIdleTime time.Duration
}
