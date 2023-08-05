package ctxslog

import (
	"log/slog"
	"strconv"
)

// Type alias for slog.HandlerOptions.ReplaceAttr.
type ReplaceAttrFunc = func(groups []string, a slog.Attr) slog.Attr

// ChainReplaceAttr chains multiple ReplaceAttrFunc together.
func ChainReplaceAttr(fs ...ReplaceAttrFunc) ReplaceAttrFunc {
	return func(groups []string, a slog.Attr) slog.Attr {
		for _, f := range fs {
			a = f(groups, a)
		}
		return a
	}
}

// GCPKeys is a ReplaceAttrFunc that replaces certain keys from Attr to meet
// Google Cloud Structured logging's expectations.
func GCPKeys(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 {
		// ref: https://cloud.google.com/logging/docs/structured-logging
		switch a.Key {
		case slog.MessageKey:
			a.Key = "message"
		case slog.LevelKey:
			a.Key = "severity"
		case slog.SourceKey:
			a.Key = "logging.googleapis.com/sourceLocation"
		}
	}
	return a
}

// StringDuration is a ReplaceAttrFunc that renders duration values as strings.
func StringDuration(groups []string, a slog.Attr) slog.Attr {
	if a.Value.Kind() == slog.KindDuration {
		a.Value = slog.StringValue(a.Value.Duration().String())
	}
	return a
}

// StringInt is a ReplaceAttrFunc that renders int64 and uint64 values as
// strings.
//
// It's useful in cases that your log ingester parses all number values as
// float64 and cause loss of precision.
func StringInt(groups []string, a slog.Attr) slog.Attr {
	switch a.Value.Kind() {
	case slog.KindInt64:
		a.Value = slog.StringValue(strconv.FormatInt(a.Value.Int64(), 10))
	case slog.KindUint64:
		a.Value = slog.StringValue(strconv.FormatUint(a.Value.Uint64(), 10))
	}
	return a
}
