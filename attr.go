package ctxslog

import (
	"golang.org/x/exp/slog"
)

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
