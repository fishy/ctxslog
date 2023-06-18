package ctxslog

import (
	"context"
	"runtime"

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

type callstackHandler struct {
	slog.Handler

	level slog.Leveler
}

func (ch *callstackHandler) WithAttr(attrs []slog.Attr) slog.Handler {
	return &callstackHandler{
		Handler: ch.Handler.WithAttrs(attrs),

		level: ch.level,
	}
}

func (ch *callstackHandler) Handle(ctx context.Context, r slog.Record) error {
	if !ch.Enabled(ctx, r.Level) {
		return nil
	}
	if r.Level >= ch.level.Level() && r.PC != 0 {
		var pcs []uintptr
		max := 20
		for {
			pcs = make([]uintptr, max)
			n := runtime.Callers(0, pcs)
			if n < max {
				pcs = pcs[:n]
				break
			}
			max = max * 2
		}
		for i, pc := range pcs {
			if pc == r.PC {
				pcs = pcs[i:]
				break
			}
		}

		fs := runtime.CallersFrames(pcs)
		var stack []*slog.Source
		for {
			f, next := fs.Next()
			stack = append(stack, &slog.Source{
				Function: f.Function,
				File:     f.File,
				Line:     f.Line,
			})
			if !next {
				break
			}
		}

		r = r.Clone()
		r.AddAttrs(slog.Any("callstack", stack))
	}
	return ch.Handler.Handle(ctx, r)
}

// CallstackHandler wraps handler to print out full callstack at minimal level.
func CallstackHandler(h slog.Handler, min slog.Leveler) slog.Handler {
	if ch, ok := h.(*callstackHandler); ok {
		// avoid double wrapping
		ch.level = min
		return ch
	}
	return &callstackHandler{
		Handler: h,

		level: min,
	}
}
