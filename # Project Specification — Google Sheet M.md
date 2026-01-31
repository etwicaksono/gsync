# Project Specification — Google Sheet Manager (gsync)

This document specifies the architecture, behavior, configuration, data contracts, and operational characteristics of the Google Sheet ↔ Database synchronization CLI application.

- Module root: [go.mod](go.mod)
- CLI entrypoint: [cmd/cli/main.go](cmd/cli/main.go)
- Primary command wiring: [internal/commands/root.go](internal/commands/root.go)
- Application service (finance app): [internal/adapters/app/finance_app/adapter.go](internal/adapters/app/finance_app/adapter.go)
- Infrastructure (database): [internal/infrastructure/gorm.go](internal/infrastructure/gorm.go)
- Configuration: [internal/config/config.go](internal/config/config.go), [internal/config/load_config.go](internal/config/load_config.go)
- Google Sheets utilities: [internal/adapters/util/gsheet.util.go](internal/adapters/util/gsheet.util.go)
- SQL schema migrations: 
  - [sql/migration/create_accounts_table.sql](sql/migration/create_accounts_table.sql)
  - [sql/migration/create_categories_table.sql](sql/migration/create_categories_table.sql)
  - [sql/migration/create_credits_table.sql](sql/migration/create_credits_table.sql)
  - [sql/migration/create_debts_table.sql](sql/migration/create_debts_table.sql)
  - [sql/migration/create_transactions_table.sql](sql/migration/create_transactions_table.sql)

## 1) Overview

The application is a command-line tool that:
- Pulls finance-related data from Google Sheets into a relational database (MySQL via GORM).
- Pushes summarized “resume” data from the database back into a Google Sheet.

It follows a ports-and-adapters (hexagonal/clean) approach:
- CLI commands orchestrate use cases.
- An application service (Finance App) implements operations to pull/push data.
- Secondary adapters implement database persistence.
- Google Sheets integration is accessed via a utility layer.

Primary executable flow:
- [main.main()](cmd/cli/main.go:5) invokes [commands.Execute()](internal/commands/root.go:41)
- [commands.Execute()](internal/commands/root.go:41) loads config, initializes logger, connects DB, composes the Finance App via DI, then executes Cobra commands.

## 2) High-Level Architecture

Layers and boundaries:
- CLI (Cobra): defines commands and flags, routes to application layer.
  - Root wiring and lifecycle: [internal/commands/root.go](internal/commands/root.go)
  - Finance command group: [internal/commands/finance.go](internal/commands/finance.go)
  - Pull subcommand: [internal/commands/finance_pull.go](internal/commands/finance_pull.go)
  - Push subcommand: [internal/commands/finance_push.go](internal/commands/finance_push.go)
- Application (Finance App): orchestrates Google Sheet reading/writing and DB persistence.
  - Constructor: [finance_app.NewFinanceApp()](internal/adapters/app/finance_app/adapter.go:23)
  - Pull workflows:
    - [finance_app.PullTransaction()](internal/adapters/app/finance_app/pull_transaction.go:17)
    - [finance_app.PullTransfer()](internal/adapters/app/finance_app/pull_transfer.go:17)
    - [finance_app.PullCredit()](internal/adapters/app/finance_app/pull_credit.go:17)
    - [finance_app.PullDebt()](internal/adapters/app/finance_app/pull_debt.go:17)
    - [finance_app.PullAccount()](internal/adapters/app/finance_app/pull_account.go:17)
    - [finance_app.PullCategory()](internal/adapters/app/finance_app/pull_category.go:17)
  - Push workflows:
    - [finance_app.PushResumeDetail()](internal/adapters/app/finance_app/push_resume_detail.go:17)
    - Push resume recap (implemented similarly; adapter wiring present)
- Ports and Adapters:
  - App port definition: [internal/ports/app/finance.app.go](internal/ports/app/finance.app.go)
  - DB ports (examples): [internal/ports/secondary/db/transaction.db.go](internal/ports/secondary/db/transaction.db.go)
  - DB adapters (implementations): internal/adapters/secondary/db/* (wired in DI)
- Infrastructure:
  - DB connectivity, logging integration, connection pool tuning:
    - [infrastructure.NewGormDb()](internal/infrastructure/gorm.go:26)
    - [infrastructure.AdapterGorm.Connect()](internal/infrastructure/gorm.go:52)
    - [infrastructure.AdapterGorm.Disconnect()](internal/infrastructure/gorm.go:98)
- Dependency Injection (google/wire):
  - Set and provider: [injector.FinanceAppProvider()](injector/injector.go:32)

## 3) Execution and Control Flow

Startup sequence in [commands.Execute()](internal/commands/root.go:41):
- Load configuration via [config.LoadConfig()](internal/config/load_config.go:15)
- Initialize structured logger (see injector/logger.injector.go)
- Connect database via [infrastructure.AdapterGorm.Connect()](internal/infrastructure/gorm.go:52) and acquire GORM client
- Construct finance app using [injector.FinanceAppProvider()](injector/injector.go:32) which wires:
  - DB adapters: transaction, transfer, debt, credit, account, category, resume
  - App adapter: [finance_app.NewFinanceApp()](internal/adapters/app/finance_app/adapter.go:23)
- Run Cobra root command (flags, subcommands).

Command routing:
- Root “version” flag prints version and exits.
- Finance command group is added in [commands.init()](internal/commands/finance.go:11)
- Subcommands registered via init in each command file:
  - Pull: [commands.init()](internal/commands/finance_pull.go:105)
  - Push: [commands.init()](internal/commands/finance_push.go:101)

## 4) CLI Specification

Binary name default: “gsync” (overridable by APP_CLI_NAME)

- Version
  - Syntax: gsync --version | -v
  - Behavior: prints version injected into [internal/commands/root.go](internal/commands/root.go) and exits.

- Finance group
  - Command: gsync finance
  - Pull
    - Syntax: gsync finance pull {transaction|transfer|credit|debt|account|category|all} [--force|-f]
    - Behavior:
      - For a single target: executes the corresponding pull use case.
      - For “all”: runs all pull targets concurrently using errgroup, failing fast on any error.
    - Flag: --force | -f
      - Behavior: Truncates relevant data before upserting in the target table(s).
    - Implementation flow (example: transactions):
      - Reading sheets: [util.MustSheetsService()](internal/adapters/util/gsheet.util.go:23), [util.ReadSheetWithPos()](internal/adapters/util/gsheet.util.go:58)
      - Transform rows → entity list, with positional metadata and timestamps.
      - DB upsert within transaction (truncate optional), commit on success.
  - Push
    - Syntax: gsync finance push {resume-detail|resume-recap|all} --start-col|-c=<A1Start> --start-date|-s=YYYY-MM-DD --end-date|-e=YYYY-MM-DD
    - Behavior:
      - For a single target: computes and writes a “resume” range back to the target sheet.
      - For “all”: iterates both “resume-detail” and “resume-recap” sequentially.
    - Flags (required):
      - --start-col, -c: A1 start cell in the target sheet (e.g., Resume!B3)
      - --start-date, -s: Inclusive start date (YYYY-MM-DD)
      - --end-date, -e: Inclusive end date (YYYY-MM-DD)
    - Implementation (resume-detail):
      - Query DB for aggregated rows
      - Build headers and rows
      - Ensure sheet/tab exists [util.EnsureTargetSheet()](internal/adapters/util/gsheet.util.go:192)
      - Batch write via Sheets API using USER_ENTERED input

## 5) Configuration

Loaded by [config.LoadConfig()](internal/config/load_config.go:15). Viper loads environment variables and optionally a .env file from:
- Next to the executable
- One level up from executable directory
- Current working directory

Primary config fields (defaults shown):
- CLI
  - APP_CLI_NAME: gsync
  - APP_CLI_VERSION: dev
  - APP_CLI_LOG_LEVEL: info (valid: debug, info, warn, error)
  - APP_CLI_TIMEOUT: 1m
- Database
  - DB_ADDRESS: localhost:27017
  - DB_NAME: example
  - DB_USERNAME: admin
  - DB_PASSWORD: admin_password
  - DB_OPTION: ?authSource=admin&amp;readPreference=secondaryPreferred
  - DB_MAX_OPEN_CONNECTIONS: 10
  - DB_MAX_IDLE_CONNECTIONS: 2
  - DB_MAX_CONNECTION_LIFETIME: 60s
  - DB_MAX_CONNECTION_IDLE_TIME: 20s
- Google Sheets — Source (Data Transaksi)
  - GSHEET_DATA_TRANSAKSI_ID: (required for IO)
  - GSHEET_CREDENTIAL_PATH: (required for IO)
  - GSHEET_DATA_TRANSAKSI_TRANSACTION_RANGE: Transaction!B5:H
  - GSHEET_DATA_TRANSAKSI_TRANSFER_RANGE: Transaction!B5:G
  - GSHEET_DATA_TRANSAKSI_CREDIT_RANGE: Liquid Assets!O5:R27
  - GSHEET_DATA_TRANSAKSI_DEBT_RANGE: Liquid Assets!U5:X27
  - GSHEET_DATA_TRANSAKSI_ACCOUNT_RANGE: Liquid Assets!B5:M28
  - GSHEET_DATA_TRANSAKSI_CATEGORY_RANGE: Categories &amp; Budget Mapping!B5:G98
- Google Sheets — Target (Rekap Keuangan)
  - GSHEET_REKAP_KEUANGAN_ID: (required for IO)
  - GSHEET_CREDENTIAL_PATH: (reused)
  - GSHEET_REKAP_KEUANGAN_RESUME_START_COL_RANGE: Resume!B3

## 6) Data Contracts

Entities (in-memory) include values and their original sheet positions (for traceability):
- Transaction: id, date, account, category, amount, note, type + position cells
  - Parser: [finance_app.PullTransaction()](internal/adapters/app/finance_app/pull_transaction.go:17)
- Transfer: id, date, fromAccount, toAccount, amount, note + position cells
  - Parser: [finance_app.PullTransfer()](internal/adapters/app/finance_app/pull_transfer.go:17)
- Credit: id, account, name, amount + position cells
  - Parser: [finance_app.PullCredit()](internal/adapters/app/finance_app/pull_credit.go:17)
- Debt: id, account, name, amount + position cells
  - Parser: [finance_app.PullDebt()](internal/adapters/app/finance_app/pull_debt.go:17)
- Account: id, code, name, active, usability, ownership + position cells
  - Parser: [finance_app.PullAccount()](internal/adapters/app/finance_app/pull_account.go:17)
- Category: id, parent, name, code, nature, visibility + position cells
  - Parser: [finance_app.PullCategory()](internal/adapters/app/finance_app/pull_category.go:17)

Google Sheets helpers:
- Credentialed service: [util.MustSheetsService()](internal/adapters/util/gsheet.util.go:23)
- Range parsing and addressing:
  - [util.ReadSheetWithPos()](internal/adapters/util/gsheet.util.go:58)
  - [util.ParseA1RangeStart()](internal/adapters/util/gsheet.util.go:99)
  - [util.ColToLetter()](internal/adapters/util/gsheet.util.go:130)
  - [util.ParseA1Cell()](internal/adapters/util/gsheet.util.go:142)
  - [util.QuoteSheet()](internal/adapters/util/gsheet.util.go:168)
- Date conversion and validation:
  - [util.FromSheetSerialDate()](internal/adapters/util/gsheet.util.go:176)
  - [util.NonEmpty()](internal/adapters/util/gsheet.util.go:182)
- Target sheet provisioning:
  - [util.EnsureTargetSheet()](internal/adapters/util/gsheet.util.go:192)

## 7) Database Schema

Tables created via SQL migrations (MySQL/InnoDB/utf8mb4):

- Accounts: [create_accounts_table.sql](sql/migration/create_accounts_table.sql)
  - Columns: id (PK), code, name, active, usability, ownership, position (JSON), created_at/by, updated_at/by
- Categories: [create_categories_table.sql](sql/migration/create_categories_table.sql)
  - Columns: id (PK), parent, name, code, nature, visibility, position (JSON), created_at/by, updated_at/by
- Credits: [create_credits_table.sql](sql/migration/create_credits_table.sql)
  - Columns: id (PK), account, name, amount, position (JSON), created_at/by, updated_at/by
- Debts: [create_debts_table.sql](sql/migration/create_debts_table.sql)
  - Columns: id (PK), account, name, amount, position (JSON), created_at/by, updated_at/by
- Transactions: [create_transactions_table.sql](sql/migration/create_transactions_table.sql)
  - Columns: id (PK), date, account, category, amount, note, type, position (JSON), created_at/by, updated_at/by

DB access layer:
- GORM 2.x, MySQL driver
- Connection management: [infrastructure.AdapterGorm.Connect()](internal/infrastructure/gorm.go:52)
- Custom GORM logger bridging slog: [infrastructure.slogGormLogger.Trace()](internal/infrastructure/gorm.go:141)

## 8) Use Cases (Behavioral Details)

Pull (Sheets → DB):
- Reads configured A1 ranges with positions using [util.ReadSheetWithPos()](internal/adapters/util/gsheet.util.go:58)
- Validates non-empty required fields per entity
- Parses numeric/date values, maps into entities with position metadata
- Begins DB transaction, optionally truncates target table (via repo), bulk upserts, commits

Push (DB → Sheets):
- Queries aggregated view/repository for the requested window
- Builds 2D row array (headers + data)
- Ensures target sheet present [util.EnsureTargetSheet()](internal/adapters/util/gsheet.util.go:192)
- Writes a top-of-sheet “Last Update” timestamp
- Appends row slices starting at provided A1 start cell using BatchUpdate with USER_ENTERED

## 9) Error Handling, Logging, and Timeouts

- Each use case wraps a context with CLI timeout from config (default 1m).
- Structured logging via slog; GORM logs routed through a custom adapter with source capture.
- Fatal or setup errors are printed to stderr and cause process exit during bootstrap.
- DB operations use gorm.WithContext, explicit Begin/Commit/Rollback.
- Sheets failures log contextual messages and return errors up the stack.

Key functions:
- [infrastructure.AdapterGorm.Connect()](internal/infrastructure/gorm.go:52)
- [infrastructure.AdapterGorm.Disconnect()](internal/infrastructure/gorm.go:98)
- [util.MustSheetsService()](internal/adapters/util/gsheet.util.go:23)

## 10) Concurrency and Performance

- finance pull “all” runs items concurrently using errgroup:
  - See logic in [internal/commands/finance_pull.go](internal/commands/finance_pull.go)
- Each pull/push is bounded by context timeouts to avoid runaway calls.
- DB pool settings configurable (max open/idle, lifetimes).
- Sheets API calls:
  - Pull: a single Values.Get per configured range.
  - Push: a BatchUpdate with accumulated row updates; header and data rows batched per request.

Note: In Go versions before 1.22, closures over loop variables require care; the “all” pull loop uses errgroup with a closure. If running on pre-1.22 Go, ensure the loop variable is captured correctly by redeclaring within the loop or target Go ≥ 1.22 where range var capture semantics are fixed.

## 11) Security and Credentials

- Google Service Account JSON path is provided by GSHEET_CREDENTIAL_PATH.
- Scope used is SpreadsheetsScope (read+write) only.
- Credentials are read from disk; ensure file permissions are restricted.
- Database credentials supplied via environment variables or .env (not committed).

## 12) Build, Install, and Run

Prerequisites:
- Go toolchain
- MySQL instance reachable from the app
- Google Service Account with Sheets access; share sheets with the service account email

Migrations:
- Apply SQL files in [sql/migration](sql/migration) to your target database using your preferred migration tool or mysql client.

Build:
- go build -o bin/gsync ./cmd/cli

Run examples:
- Pull all with truncate:
  - gsync finance pull all -f
- Pull only transactions:
  - gsync finance pull transaction
- Push resume detail for January 2024 starting at Resume!B3:
  - gsync finance push resume-detail -c "Resume!B3" -s 2024-01-01 -e 2024-01-31
- Show version:
  - gsync -v

## 13) Assumptions and Constraints

- Sheet ranges configured include headers or start directly at data rows depending on actual sheet; parsing uses indices per entity expectations.
- IDs come from sheet data and are treated as primary keys in DB.
- Position JSON stores original A1 cell locations for traceability.
- No foreign key constraints defined in migrations; relationships (e.g., category/account) are string-keyed.

## 14) Known Limitations

- No built-in migration runner; SQL must be applied externally.
- Limited validation of data types and domain values from sheets.
- Resume “recap” push is wired in CLI but its adapter implementation is not shown in this spec; expected to mirror detail push behavior.
- Error messages from Sheets are surfaced but not categorized for retry strategies.
- Assumes Google API quotas are not exceeded; no backoff/retry implemented beyond context timeouts.

## 15) Extensibility and Future Enhancements

- Add retries with exponential backoff for Sheets/DB transient errors.
- Introduce structured validation and per-row error reporting with optional “write-back” of errors into sheet notes.
- Add configurable concurrency limits on “all” pull to respect API quotas.
- Implement FK constraints and normalization in DB as needed.
- Provide a dedicated migration tool integration (golang-migrate).
- Add metrics and tracing (OpenTelemetry) for end-to-end observability.

## 16) Key Code References

- Bootstrap
  - [main.main()](cmd/cli/main.go:5)
  - [commands.Execute()](internal/commands/root.go:41)
- Commands
  - [commands.init()](internal/commands/finance.go:11)
  - [internal/commands/finance_pull.go](internal/commands/finance_pull.go)
  - [internal/commands/finance_push.go](internal/commands/finance_push.go)
- Application
  - [finance_app.NewFinanceApp()](internal/adapters/app/finance_app/adapter.go:23)
  - [finance_app.PullTransaction()](internal/adapters/app/finance_app/pull_transaction.go:17)
  - [finance_app.PullTransfer()](internal/adapters/app/finance_app/pull_transfer.go:17)
  - [finance_app.PullCredit()](internal/adapters/app/finance_app/pull_credit.go:17)
  - [finance_app.PullDebt()](internal/adapters/app/finance_app/pull_debt.go:17)
  - [finance_app.PullAccount()](internal/adapters/app/finance_app/pull_account.go:17)
  - [finance_app.PullCategory()](internal/adapters/app/finance_app/pull_category.go:17)
  - [finance_app.PushResumeDetail()](internal/adapters/app/finance_app/push_resume_detail.go:17)
- Sheets Utilities
  - [util.MustSheetsService()](internal/adapters/util/gsheet.util.go:23)
  - [util.ReadSheetWithPos()](internal/adapters/util/gsheet.util.go:58)
  - [util.ParseA1RangeStart()](internal/adapters/util/gsheet.util.go:99)
  - [util.EnsureTargetSheet()](internal/adapters/util/gsheet.util.go:192)
- Infrastructure / Config
  - [infrastructure.NewGormDb()](internal/infrastructure/gorm.go:26)
  - [infrastructure.AdapterGorm.Connect()](internal/infrastructure/gorm.go:52)
  - [infrastructure.AdapterGorm.Disconnect()](internal/infrastructure/gorm.go:98)
  - [config.LoadConfig()](internal/config/load_config.go:15)

This specification reflects the current repository state and code paths, with links pointing to the exact definitions for quick navigation and verification.