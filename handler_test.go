package ctxslog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/exp/slog"

	"go.yhsif.com/ctxslog"
)

func TestContextHandler(t *testing.T) {
	backup := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(backup)
	})

	var sb strings.Builder
	base := slog.New(ctxslog.ContextHandler(slog.NewJSONHandler(&sb, nil)))
	slog.SetDefault(base)
	ctx := ctxslog.Attach(context.Background(), slog.String("foo", "bar"))
	slog.InfoCtx(ctx, "test")
	line := sb.String()
	for _, s := range []string{
		`"msg":"test"`,
		`"foo":"bar"`,
	} {
		if !strings.Contains(line, s) {
			t.Errorf("%s does not have %s", line, s)
		}
	}
}

func TestJSONCallstackHandler(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(ctxslog.JSONCallstackHandler(slog.NewJSONHandler(
		&buf,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}), slog.LevelInfo,
	)).With("foo", "bar")
	l := func(l slog.Level) {
		logger.Log(context.Background(), l, "test")
	}
	type lineJSON struct {
		Source    slog.Source   `json:"source"`
		Callstack []slog.Source `json:"callstack"`
	}

	t.Run(slog.LevelDebug.String(), func(t *testing.T) {
		buf.Reset()
		l(slog.LevelDebug)
		var line lineJSON
		if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
			t.Fatal(err)
		}
		if len(line.Callstack) > 0 {
			t.Errorf("Don't expect callstack, got %#v", line.Callstack)
		}
	})

	for _, level := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		t.Run(level.String(), func(t *testing.T) {
			buf.Reset()
			l(level)
			var line lineJSON
			if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
				t.Fatal(err)
			}
			if len(line.Callstack) == 0 {
				t.Fatalf("No callstack in log, fullline=%s", buf.String())
			}
			if line.Callstack[0] != line.Source {
				t.Errorf("line.Callstack[0]=%#v != line.Source=%#v", line.Callstack[0], line.Source)
			}
		})
	}
}

func TestTextCallstackHandler(t *testing.T) {
	// Example:
	// source=/path/to/ctxslog/handler_test.go:74
	// or
	// source="/path/to/some directory/ctxslog/handler_test.go:74"
	re, err := regexp.Compile(`source="?(.*?handler_test.go:(?:\d*))`)
	if err != nil {
		t.Fatalf("Failed to compile regexp: %v", err)
	}
	var sb strings.Builder
	logger := slog.New(ctxslog.TextCallstackHandler(slog.NewTextHandler(
		&sb,
		&slog.HandlerOptions{
			AddSource: true,
			Level:     slog.LevelDebug,
		}), slog.LevelInfo,
	)).With("foo", "bar")
	l := func(l slog.Level) {
		logger.Log(context.Background(), l, "test")
	}

	t.Run(slog.LevelDebug.String(), func(t *testing.T) {
		sb.Reset()
		l(slog.LevelDebug)
		line := sb.String()
		if strings.Contains(line, "callstack=") {
			t.Errorf("Should not have callstack on this level: %s", line)
		}
	})

	for _, level := range []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError} {
		t.Run(level.String(), func(t *testing.T) {
			sb.Reset()
			l(level)
			line := sb.String()
			groups := re.FindStringSubmatch(line)
			if len(groups) == 0 {
				t.Fatalf("Didn't find source in log: %s", line)
			}
			callstack0 := groups[1]
			if !strings.Contains(line, `callstack="[`+callstack0) {
				t.Errorf("Did not find first callstack matching source (%s): %s", callstack0, line)
			}
		})
	}
}
