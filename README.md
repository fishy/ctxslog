[![PkgGoDev](https://pkg.go.dev/badge/go.yhsif.com/ctxslog)](https://pkg.go.dev/go.yhsif.com/ctxslog)
[![Go Report Card](https://goreportcard.com/badge/go.yhsif.com/ctxslog)](https://goreportcard.com/report/go.yhsif.com/ctxslog)

# ctxslog

Go library provides helpers around slog

## Example

(See [package example on pkg.go.dev][package-example] for up-to-date example)

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"go.yhsif.com/ctxslog"
)

func EndpointHandler(ctx context.Context, traceID string) {
	// Inside an endpoint handler, attach trace id for all logs in this context
	ctx = ctxslog.Attach(ctx, "trace", traceID)
	// Now slog's global log functions with ctx will also have trace info
	slog.ErrorCtx(ctx, "Not implemented")

	thirdPartyLibCall := func(ctx context.Context) {
		// Some third party library that uses slog and spams logs a lot.
		slog.ErrorCtx(ctx, "not really an error")
	}
	thirdPartyLibCall(
		ctxslog.AttachLogLevel(ctx, ctxslog.MaxLevel), // now even if it logs at error level it won't shown
	)
}

func HttpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := ctxslog.Attach(
		r.Context(),
		// Add httpRequest group to every log within this context
		slog.Group("httpRequest", ctxslog.HTTPRequest(r, ctxslog.RemoteAddrIP)),
	)
	// Enable callstack even at debug level for this context
	ctx = ctxslog.AttachCallstackLevel(ctx, slog.LevelDebug)
	slog.DebugCtx(ctx, "foo") // this log will contain httpRequest group and callstack.
}

func main() {
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
```

[package-example]: https://pkg.go.dev/go.yhsif.com/ctxslog#example-package
