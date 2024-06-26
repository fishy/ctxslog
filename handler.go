package ctxslog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"runtime"
)

// Minimal and maximal possible log levels.
//
// You can use MinLevel as the log level or callstack level in context/logger to
// include all logs, or use MaxLevel to exclude all logs (except logs logged
// explicitly at MaxLevel).
const (
	MinLevel = slog.Level(math.MinInt)
	MaxLevel = slog.Level(math.MaxInt)
)

type logKeyType struct{}

var logKey logKeyType

type logLevelType struct{}

var logLevelKey logLevelType

type callstackLevelType struct{}

var callstackLevelKey callstackLevelType

// Attaches logger args into context.
//
// NOTE: This does in most cases require that you already called slog.SetDefault
// on a logger retruend by New.
func Attach(ctx context.Context, args ...any) context.Context {
	logger := slog.Default()
	if l, ok := ctx.Value(logKey).(*slog.Logger); ok {
		logger = l
	}
	return context.WithValue(ctx, logKey, logger.With(args...))
}

// AttachLogLevel attaches min log level (inclusive) to the context,
// overriding the global one set on the logger.
func AttachLogLevel(ctx context.Context, level slog.Leveler) context.Context {
	return context.WithValue(ctx, logLevelKey, level)
}

// AttachCallstackLevel attaches min callstack level (inclusive) to the context,
// overriding the global one set on the logger.
func AttachCallstackLevel(ctx context.Context, level slog.Leveler) context.Context {
	return context.WithValue(ctx, callstackLevelKey, level)
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
	if level, _ := ctx.Value(logLevelKey).(slog.Leveler); level != nil {
		return l >= level.Level()
	}
	return ch.h.Enabled(ctx, l)
}

func (ch ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ctxHandler{h: ch.h.WithAttrs(attrs)}
}

func (ch ctxHandler) WithGroup(name string) slog.Handler {
	return &ctxHandler{h: ch.h.WithGroup(name)}
}

// ContextHandler wraps handler to handle contexts from Attach and
// AttachLogLevel.
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
}

func (ch *callstackHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return ch.h.Enabled(ctx, l)
}

func (ch *callstackHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &callstackHandler{
		h: ch.h.WithAttrs(attrs),

		level: ch.level,
	}
}

func (ch *callstackHandler) WithGroup(name string) slog.Handler {
	return &callstackHandler{
		h: ch.h.WithGroup(name),

		level: ch.level,
	}
}

func (ch *callstackHandler) Handle(ctx context.Context, r slog.Record) error {
	level, _ := ctx.Value(callstackLevelKey).(slog.Leveler)
	if level == nil {
		level = ch.level
	}
	if r.Level >= level.Level() && r.PC != 0 {
		var pcs []uintptr
		max := 20
		for {
			pcs = make([]uintptr, max)
			n := runtime.Callers(0, pcs)
			if n < max {
				pcs = pcs[:n]
				break
			}
			max += max
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
			r.AddAttrs(slog.Any("callstack", callstack(pcs)))
		}
	}
	return ch.h.Handle(ctx, r)
}

type wrapSource slog.Source

func (ws *wrapSource) MarshalJSON() ([]byte, error) {
	return json.Marshal((*slog.Source)(ws))
}

func (ws *wrapSource) String() string {
	return fmt.Sprintf("%s:%d", ws.File, ws.Line)
}

func callstack(pcs []uintptr) []*wrapSource {
	fs := runtime.CallersFrames(pcs)
	stack := make([]*wrapSource, 0, len(pcs))
	for {
		f, next := fs.Next()
		stack = append(stack, &wrapSource{
			Function: f.Function,
			File:     f.File,
			Line:     f.Line,
		})
		if !next {
			break
		}
	}
	return stack
}

// CallstackHandler wraps handler to print out full callstack at minimal level
// (inclusive).
//
// If h is already a CallstackHandler,
// its configured min level will be modified instead.
func CallstackHandler(h slog.Handler, min slog.Leveler) slog.Handler {
	if ch, ok := h.(*callstackHandler); ok {
		// avoid double wrapping
		ch.level = min
		return ch
	}
	return &callstackHandler{
		h: h,

		level: min,
	}
}
