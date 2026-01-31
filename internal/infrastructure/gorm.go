package infrastructure

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/etwicaksono/data-transaksi-service/internal/constants"
	errorsConst "github.com/etwicaksono/data-transaksi-service/internal/errors"
	"gorm.io/driver/postgres"
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
	var dialector gorm.Dialector

	// Parse host and port from Address (format: "host:port" or just "host")
	host := config.Db.Address
	port := "5432" // default PostgreSQL port
	if idx := strings.LastIndex(config.Db.Address, ":"); idx != -1 {
		host = config.Db.Address[:idx]
		port = config.Db.Address[idx+1:]
	}

	dialector = postgres.Open(
		fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s %s",
			host,
			port,
			config.Db.Username,
			config.Db.Password,
			config.Db.Name,
			config.Db.Option,
		),
	)

	return &AdapterGorm{
		ctx:       ctx,
		config:    config,
		dialector: dialector,
		logger:    logger,
	}
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

	gormLogger := &slogGormLogger{logger: a.logger, level: logLevel}

	db, err := gorm.Open(a.dialector, &gorm.Config{
		Logger:         gormLogger,
		TranslateError: true,
	})
	if err != nil {
		slog.ErrorContext(a.ctx, "Failed to connect to GormDB", slog.String(constants.Error, err.Error()))
		return err
	}

	slog.InfoContext(a.ctx, "GormDB connected")
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

	return nil
}

func (a *AdapterGorm) Disconnect() {
	sqlDB, err := a.client.DB()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to get database object: %v", err))
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
