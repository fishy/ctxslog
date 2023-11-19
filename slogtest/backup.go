package slogtest

import (
	"log/slog"
	"testing"
)

// BackupGlobalLogger backs up the global slog logger and restores it after
// test execution.
func BackupGlobalLogger(tb testing.TB) {
	backup := slog.Default()
	tb.Cleanup(func() {
		slog.SetDefault(backup)
	})
}
