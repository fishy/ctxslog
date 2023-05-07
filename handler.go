package ctxslog

import (
	"context"

	"golang.org/x/exp/slog"
)

type logKeyType struct{}

var logKey logKeyType

// Attaches l into context.
func Attach(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

type ctxHandler struct {
	slog.Handler
}

func (ch ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if l, ok := ctx.Value(logKey).(*slog.Logger); ok {
		return l.Handler().Handle(context.Background(), r)
	}
	return ch.Handler.Handle(ctx, r)
}

// ContextHandler wraps handler to handle contexts from Attach.
func ContextHandler(h slog.Handler) slog.Handler {
	return &ctxHandler{h}
}
