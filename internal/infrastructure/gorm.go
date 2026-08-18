package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/adapters/secondary/model"
	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	errorsConst "github.com/etwicaksono/data-transaksi-service/internal/errors"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AdapterGorm struct {
	ctx       context.Context
	config    config.Config
	client    *gorm.DB
	logger    *slog.Logger
	dialector gorm.Dialector
}

func NewGormDb(
	ctx context.Context,
	config config.Config,
	logger *slog.Logger,
) *AdapterGorm {
	return &AdapterGorm{
		ctx:       ctx,
		config:    config,
		dialector: sqlite.Open(buildSqliteDSN(config.Db.Path, config.Db.Option)),
		logger:    logger,
	}
}

// buildSqliteDSN appends the pragmas required for reliable concurrent access.
// WAL lets readers run alongside a writer, and busy_timeout makes competing
// writers wait for the lock instead of failing with "database is locked".
func buildSqliteDSN(path, option string) string {
	pragmas := []string{
		"_pragma=journal_mode(WAL)",
		"_pragma=busy_timeout(10000)",
		"_pragma=foreign_keys(ON)",
	}

	if trimmed := strings.TrimLeft(strings.TrimSpace(option), "?&"); trimmed != "" {
		pragmas = append(pragmas, trimmed)
	}

	return fmt.Sprintf("%s?%s", path, strings.Join(pragmas, "&"))
}

func (a *AdapterGorm) Connect() error {
	// Set GORM logger to use SlogGormLogger with the desired log level
	var logLevel logger.LogLevel

	switch a.config.CliApp.LogLevel {
	case "debug":
		logLevel = logger.Info
	case "info":
		logLevel = logger.Info
	case "warn":
		logLevel = logger.Warn
	case "error":
		logLevel = logger.Error
	default:
		return errorsConst.ErrLogLevelInvalid
	}

	// The database is a plain file, so make sure its parent directory exists
	// before the driver tries to create it.
	if dir := filepath.Dir(a.config.Db.Path); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			slog.ErrorContext(a.ctx, "Failed to create database directory", slog.String(constants.Error, err.Error()))
			return err
		}
	}

	gormLogger := &slogGormLogger{logger: a.logger, level: logLevel}

	db, err := gorm.Open(a.dialector, &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	})
	if err != nil {
		slog.ErrorContext(a.ctx, "Failed to connect to GormDB", slog.String(constants.Error, err.Error()))
		return err
	}

	slog.InfoContext(a.ctx, "GormDB connected", slog.String("path", a.config.Db.Path))
	a.client = db

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("Failed to get *sql.DB: %v", slog.String(constants.Error, err.Error()))
		return err
	}

	// Set advanced connection settings
	sqlDB.SetMaxOpenConns(a.config.Db.MaxOpenConnections)
	sqlDB.SetMaxIdleConns(a.config.Db.MaxIdleConnections)
	sqlDB.SetConnMaxLifetime(a.config.Db.MaxConnectionLifetime * time.Second)
	sqlDB.SetConnMaxIdleTime(a.config.Db.MaxConnectionIdleTime * time.Second)

	if err = a.migrate(); err != nil {
		return err
	}

	return nil
}

// migrate creates the schema on first run so a fresh SQLite file is usable
// without any manual setup step.
func (a *AdapterGorm) migrate() error {
	err := a.client.AutoMigrate(
		&model.Account{},
		&model.Category{},
		&model.Credit{},
		&model.Debt{},
		&model.Transaction{},
		&model.Transfer{},
	)
	if err != nil {
		slog.ErrorContext(a.ctx, "Failed to migrate database schema", slog.String(constants.Error, err.Error()))
		return err
	}

	slog.InfoContext(a.ctx, "Database schema is up to date")
	return nil
}

func (a *AdapterGorm) Disconnect() {
	if a.client == nil {
		return
	}

	sqlDB, err := a.client.DB()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get database object: %v", err))
		return
	}

	if err = sqlDB.Close(); err != nil {
		slog.Error(fmt.Sprintf("Failed to close database connection: %v", err))
	}
	slog.Info("Database connection closed")
}

func (a *AdapterGorm) GetClient() *gorm.DB {
	return a.client
}

type slogGormLogger struct {
	logger *slog.Logger
	level  logger.LogLevel
}

func (l *slogGormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &slogGormLogger{logger: l.logger, level: level}
}

func (l *slogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		l.logger.InfoContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *slogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		l.logger.WarnContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *slogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		l.logger.ErrorContext(ctx, fmt.Sprintf(msg, data...))
	}
}

func (l *slogGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Capture the caller (file and line number)
	_, file, line, _ := runtime.Caller(3) // Adjust the depth as needed
	source := fmt.Sprintf("%s:%d", file, line)

	logData := []interface{}{
		"sql", sql,
		"rows", rows,
		"elapsed", elapsed,
		"source", source,
	}

	if err != nil && l.level >= logger.Error {
		l.logger.ErrorContext(ctx, "Error executing SQL", slog.Any(constants.Error, logData))
	} else if l.level >= logger.Info {
		l.logger.InfoContext(ctx, "SQL executed", logData...)
	}
}
