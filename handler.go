package ctxslog

import (
	"context"
	"fmt"
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
	h slog.Handler
}

func (ch ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	if l, ok := ctx.Value(logKey).(*slog.Logger); ok && l != nil {
		// override the logger in context to avoid infinite recursion
		ctx := context.WithValue(ctx, logKey, (*slog.Logger)(nil))
		return l.Handler().Handle(ctx, r)
	}
	return ch.h.Handle(ctx, r)
}

func (ch ctxHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return ch.h.Enabled(ctx, l)
}

func (ch ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{h: ch.h.WithAttrs(attrs)}
}

func (ch ctxHandler) WithGroup(name string) slog.Handler {
	return &ctxHandler{h: ch.h.WithGroup(name)}
}

// ContextHandler wraps handler to handle contexts from Attach.
func ContextHandler(h slog.Handler) slog.Handler {
	if _, ok := h.(*ctxHandler); ok {
		// avoid double wrapping
		return h
	}
	return &ctxHandler{h: h}
}

type callstackHandler struct {
	h slog.Handler

	level slog.Leveler
	json  bool
}

func (ch *callstackHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return ch.h.Enabled(ctx, l)
}

func (ch *callstackHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &callstackHandler{
		h: ch.h.WithAttrs(attrs),

		level: ch.level,
		json:  ch.json,
	}
}

func (ch *callstackHandler) WithGroup(name string) slog.Handler {
	return &callstackHandler{
		h: ch.h.WithGroup(name),

		level: ch.level,
		json:  ch.json,
	}
}

func (ch *callstackHandler) Handle(ctx context.Context, r slog.Record) error {
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
		// Skip everything before r.PC if possible.
		// Those are mostly just internal slog related wrappers.
		for i, pc := range pcs {
			if pc == r.PC {
				pcs = pcs[i:]
				break
			}
		}

		if len(pcs) > 0 {
			r = r.Clone()
			r.AddAttrs(slog.Any("callstack", callstack(pcs, ch.json)))
		}
	}
	return ch.h.Handle(ctx, r)
}

func callstack(pcs []uintptr, json bool) []any {
	fs := runtime.CallersFrames(pcs)
	stack := make([]any, 0, len(pcs))
	for {
		f, next := fs.Next()
		if json {
			stack = append(stack, &slog.Source{
				Function: f.Function,
				File:     f.File,
				Line:     f.Line,
			})
		} else {
			stack = append(stack, fmt.Sprintf("%s:%d", f.File, f.Line))
		}
		if !next {
			break
		}
	}
	return stack
}

// JSONCallstackHandler wraps handler to print out full callstack at minimal
// level in full json format (function+filename+line).
func JSONCallstackHandler(h slog.Handler, min slog.Leveler) slog.Handler {
	if ch, ok := h.(*callstackHandler); ok {
		// avoid double wrapping
		ch.level = min
		ch.json = true
		return ch
	}
	return &callstackHandler{
		h: h,

		level: min,
		json:  true,
	}
}

// TextCallstackHandler wraps handler to print out full callstack at minimal
// level in simple text format (filename:line).
func TextCallstackHandler(h slog.Handler, min slog.Leveler) slog.Handler {
	if ch, ok := h.(*callstackHandler); ok {
		// avoid double wrapping
		ch.level = min
		ch.json = false
		return ch
	}
	return &callstackHandler{
		h: h,

		level: min,
		json:  false,
	}
}
