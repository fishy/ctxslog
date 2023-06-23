package ctxslog_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"go.yhsif.com/ctxslog"
)

func TestContextHandler(t *testing.T) {
	backup := slog.Default()
	t.Cleanup(func() {
		slog.SetDefault(backup)
	})

	var sb strings.Builder
	ctxslog.New(ctxslog.WithWriter(&sb))
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
	const min = slog.LevelInfo + 1
	var buf bytes.Buffer
	logger := ctxslog.New(
		ctxslog.WithWriter(&buf),
		ctxslog.WithJSON,
		ctxslog.WithAddSource(true),
		ctxslog.WithLevel(slog.LevelDebug),
		ctxslog.WithCallstack(min),
		ctxslog.WithGlobalKVs("foo", "bar"),
	)
	l := func(l slog.Level) {
		logger.Log(context.Background(), l, "test")
	}
	type lineJSON struct {
		Source    slog.Source   `json:"source"`
		Callstack []slog.Source `json:"callstack"`
	}

	for level := slog.LevelDebug; level <= slog.LevelError; level++ {
		t.Run(level.String(), func(t *testing.T) {
			buf.Reset()
			l(level)
			t.Log(buf.String())
			var line lineJSON
			if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
				t.Fatal(err)
			}
			if level < min {
				if len(line.Callstack) > 0 {
					t.Errorf("Don't expect callstack, got %#v", line.Callstack)
				}
			} else {
				if len(line.Callstack) == 0 {
					t.Fatal("No callstack in log")
				}
				if line.Callstack[0] != line.Source {
					t.Errorf("line.Callstack[0]=%#v != line.Source=%#v", line.Callstack[0], line.Source)
				}
			}
		})
	}
}

func TestTextCallstackHandler(t *testing.T) {
	const min = slog.LevelInfo + 1

	// Example:
	// source=/path/to/ctxslog/handler_test.go:74
	// or
	// source="/path/to/some directory/ctxslog/handler_test.go:74"
	re, err := regexp.Compile(`source="?(.*?handler_test.go:(?:\d*))`)
	if err != nil {
		t.Fatalf("Failed to compile regexp: %v", err)
	}

	var sb strings.Builder
	logger := ctxslog.New(
		ctxslog.WithWriter(&sb),
		ctxslog.WithText,
		ctxslog.WithAddSource(true),
		ctxslog.WithLevel(slog.LevelDebug),
		ctxslog.WithCallstack(min),
		ctxslog.WithGlobalKVs("foo", "bar"),
	)
	l := func(l slog.Level) {
		logger.Log(context.Background(), l, "test")
	}

	for level := slog.LevelDebug; level <= slog.LevelError; level++ {
		t.Run(level.String(), func(t *testing.T) {
			sb.Reset()
			l(level)
			line := sb.String()
			t.Log(line)
			if level < min {
				if strings.Contains(line, "callstack=") {
					t.Error("Should not have callstack on this level")
				}
			} else {
				groups := re.FindStringSubmatch(line)
				if len(groups) == 0 {
					t.Fatal("Didn't find source in log")
				}
				callstack0 := groups[1]
				if !strings.Contains(line, `callstack="[`+callstack0) {
					t.Errorf("Did not find first callstack matching source: %s", callstack0)
				}
			}
		})
	}
}
