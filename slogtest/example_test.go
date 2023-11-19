package slogtest_test

import (
	"log/slog"
	"testing"

	"go.yhsif.com/ctxslog/slogtest"
)

// In a real test this function should be called TestSlog instead,
// but doing that will break the example.
func SlogTest(t *testing.T) {
	// Back up global logger and restore it after this test
	slogtest.BackupGlobalLogger(t)

	slog.SetDefault(slog.New(slogtest.Handler(
		t,
		slog.LevelInfo, // minimal level to keep
		slog.LevelWarn, // minimal level to fail test
	)))

	slog.Debug("debug") // this log will not show
	slog.Info("info")

	// The next 2 logs will fail the test
	slog.Warn("warn")
	slog.Error("error")
}

// This example demonstrates how to use slogtest to write unit tests with slog.
func Example() {
	// The real example is in SlogTest function.
}
