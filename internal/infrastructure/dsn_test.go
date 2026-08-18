package infrastructure

import (
	"context"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	"github.com/etwicaksono/data-transaksi-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSqliteDSN(t *testing.T) {
	t.Run("applies concurrency pragmas", func(t *testing.T) {
		dsn := buildSqliteDSN("gsync.db", "")

		path, query, found := strings.Cut(dsn, "?")
		require.True(t, found, "dsn must carry pragmas")
		assert.Equal(t, "gsync.db", path)
		assert.Contains(t, query, "journal_mode(WAL)")
		assert.Contains(t, query, "busy_timeout(10000)")
	})

	t.Run("merges extra options without duplicating separators", func(t *testing.T) {
		for _, option := range []string{"_pragma=cache_size(-2000)", "?_pragma=cache_size(-2000)", "&_pragma=cache_size(-2000)"} {
			dsn := buildSqliteDSN("gsync.db", option)
			assert.Equal(t, 1, strings.Count(dsn, "?"), "option %q produced a malformed dsn: %s", option, dsn)
			assert.Contains(t, dsn, "cache_size(-2000)")
		}
	})
}

// The configured path may point into a directory that does not exist yet.
func TestConnectCreatesMissingDirectory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "nested", "deeper", "gsync.db")

	cfg := config.Config{
		CliApp: config.CliApp{App: config.App{Name: "gsync-test", LogLevel: "error"}},
		Db: config.DbConfig{
			Path:               dbPath,
			MaxOpenConnections: 4,
			MaxIdleConnections: 2,
		},
	}

	a := NewGormDb(context.Background(), cfg, slog.New(slog.NewTextHandler(io_Discard{}, nil)))
	require.NoError(t, a.Connect(), "Connect must create the database directory")

	t.Cleanup(func() {
		if sqlDB, err := a.client.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})

	assert.True(t, a.client.Migrator().HasTable("transactions"))
}
