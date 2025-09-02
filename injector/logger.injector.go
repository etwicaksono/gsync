package injector

import (
	"log/slog"
	"os"

	"github.com/etwicaksono/data-transaksi-service/internal/config"
	errorsConst "github.com/etwicaksono/data-transaksi-service/internal/errors"
)

func LoggerInit(cfg config.App) (logger *slog.Logger, err error) {
	var logLevel slog.Level

	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		return nil, errorsConst.ErrLogLevelInvalid
	}

	//Initiate logger
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     logLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			return a
		},
	}).WithAttrs([]slog.Attr{
		slog.String("service", cfg.Name),
		slog.String("with-release", cfg.Version),
	})
	logger = slog.New(logHandler)
	slog.SetDefault(logger)
	return logger, nil
}
