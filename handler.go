package ctxslog

import (
	"context"

	"golang.org/x/exp/slog"
)

type logKeyType struct{}

var logKey logKeyType

// Attaches logger args into context.
func Attach(ctx context.Context, args ...any) context.Context {
	logger := slog.Default()
	if l, ok := ctx.Value(logKey).(*slog.Logger); ok {
		logger = l
	}
	return context.WithValue(ctx, logKey, logger.With(args...))
}

type ctxHandler struct {
	slog.Handler
}

func (ch ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if l, ok := ctx.Value(logKey).(*slog.Logger); ok && l != nil {
		// override the logger in context to avoid infinite recursion
		ctx := context.WithValue(ctx, logKey, (*slog.Logger)(nil))
		return l.Handler().Handle(ctx, r)
	}
	return ch.Handler.Handle(ctx, r)
}

func (ch ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{ch.Handler.WithAttrs(attrs)}
}

// ContextHandler wraps handler to handle contexts from Attach.
func ContextHandler(h slog.Handler) slog.Handler {
	if _, ok := h.(*ctxHandler); ok {
		// avoid double wrapping
		return h
	}
	return &ctxHandler{h}
}
