package ctxslog_test

import (
	"context"
	"os"

	"go.yhsif.com/ctxslog"
	"golang.org/x/exp/slog"
)

func EndpointHandler(ctx context.Context, traceID string) {
	// Inside an endpoint handler, attach trace id for all logs in this context
	ctx = ctxslog.Attach(ctx, "trace", traceID)
	// Now slog's global log functions with ctx will also have trace info
	slog.ErrorCtx(ctx, "Not implemented")
}

func Example() {
	// Sets the global slog logger
	ctxslog.New(
		ctxslog.WithWriter(os.Stderr),                              // This is the default and can be omitted
		ctxslog.WithJSON,                                           // This is the default, use ctxslog.WithText instead if you want non-json logs
		ctxslog.WithAddSource(true),                                // Add source info
		ctxslog.WithLevel(slog.LevelDebug),                         // Keep debug level logs
		ctxslog.WithCallstack(slog.LevelError),                     // For error and above levels, also add callstack info
		ctxslog.WithGlobalKVs("version", os.Getenv("VERSION_TAG")), // Add version info to every log
		ctxslog.WithReplaceAttr(ctxslog.ChainReplaceAttr(
			ctxslog.GCPKeys,        // Use Google Cloud Structured logging friendly log keys
			ctxslog.StringDuration, // Log time durations as strings
		)),
	)
	// Now you can just use slog's global log functions
	slog.Info("Hello, world!", "key", "value")
}
