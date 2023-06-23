package ctxslog

import (
	"io"
	"os"

	"golang.org/x/exp/slog"
)

type options struct {
	w           io.Writer
	json        bool
	addSource   bool
	level       slog.Leveler
	replaceAttr ReplaceAttrFunc
	callstack   slog.Leveler
	kvs         []any
}

// Option define logger options for New.
type Option func(*options)

// WithWriter sets the writer of the logger.
//
// Default: os.Stderr.
func WithWriter(w io.Writer) Option {
	return func(o *options) {
		o.w = w
	}
}

// WithJSON sets the logger to be json logger.
//
// This is the default behavior.
func WithJSON(o *options) {
	o.json = true
}

// WithText sets the logger to be text logger.
func WithText(o *options) {
	o.json = false
}

// WithLevel sets the minimal log level.
//
// Default: slog.InfoLevel.
func WithLevel(l slog.Leveler) Option {
	return func(o *options) {
		o.level = l
	}
}

// WithAddSource sets the AddSouce option.
//
// Default: false.
func WithAddSource(v bool) Option {
	return func(o *options) {
		o.addSource = v
	}
}

// WithReplaceAttr sets the ReplaceAttr option.
//
// Note that this option is overwriting not cumulative.
// To chain several ReplaceAttr functions, use ChainReplaceAttr.
func WithReplaceAttr(f ReplaceAttrFunc) Option {
	return func(o *options) {
		o.replaceAttr = f
	}
}

// WithCallstack adds callstack at min level.
//
// Set it to nil to disable callstack at all levels.
// To add callstack at all levels, use slog.Level(math.MinInt) instead.
//
// Default: nil.
func WithCallstack(min slog.Leveler) Option {
	return func(o *options) {
		o.callstack = min
	}
}

// WithGlobalKVs sets global key-value pairs.
func WithGlobalKVs(kv ...any) Option {
	return func(o *options) {
		o.kvs = kv
	}
}

// New creates a *slog.Logger that can handle contexts.
//
// It also calls slog.SetDefault before returning.
//
// Note that importing this package also has side-effect of calling New with all
// default options (setting global, default logger for slog to be context aware
// logger).
func New(opts ...Option) *slog.Logger {
	opt := options{
		w:    os.Stderr,
		json: true,
	}
	for _, o := range opts {
		o(&opt)
	}

	var handler slog.Handler
	if opt.json {
		handler = slog.NewJSONHandler(opt.w, &slog.HandlerOptions{
			AddSource:   opt.addSource,
			Level:       opt.level,
			ReplaceAttr: opt.replaceAttr,
		})
	} else {
		handler = slog.NewTextHandler(opt.w, &slog.HandlerOptions{
			AddSource:   opt.addSource,
			Level:       opt.level,
			ReplaceAttr: opt.replaceAttr,
		})
	}
	if opt.callstack != nil {
		handler = CallstackHandler(handler, opt.callstack)
	}
	handler = ContextHandler(handler)

	logger := slog.New(handler).With(opt.kvs...)
	slog.SetDefault(logger)
	return logger
}

func init() {
	New()
}
