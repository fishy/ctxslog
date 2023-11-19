package slogtest

import (
	"bufio"
	"log/slog"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	logger := slog.New(Handler(t, slog.LevelInfo, slog.LevelError))
	logger.Debug("debug")
	logger.Info("info")
	// Next one will fail the test
	// logger.Error("error")
}

type fakeTB struct {
	testing.TB

	realTB   testing.TB
	calls    int
	lastCall string
}

func (f *fakeTB) Log(args ...any) {
	f.calls++
	if len(args) != 1 {
		f.realTB.Errorf("Log not called with exactly one arg: %#v", args)
		return
	}
	arg := args[0]
	str, ok := arg.(string)
	if ok {
		f.lastCall = str
		return
	}
	f.realTB.Errorf("Log called with %#v, expected a string", arg)
}

func TestBufferWriter(t *testing.T) {
	t.Run("one-line", func(t *testing.T) {
		const str = "foobar"
		f := &fakeTB{realTB: t}
		writer := bufio.NewWriter(&bufferWriter{tb: f})
		if _, err := writer.WriteString(str + "\n"); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
		if err := writer.Flush(); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}
		if f.calls != 1 {
			t.Errorf("tb.calls got %d want 1", f.calls)
		}
		if f.lastCall != str {
			t.Errorf("tb.lastCall got %q want %q", f.lastCall, str)
		}
	})
	t.Run("multi-line", func(t *testing.T) {
		const str = "foobar"
		f := &fakeTB{realTB: t}
		writer := bufio.NewWriter(&bufferWriter{tb: f})
		if _, err := writer.WriteString(strings.Join([]string{
			"msg",
			str,
			"",
		}, "\n")); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
		if err := writer.Flush(); err != nil {
			t.Fatalf("Failed to flush: %v", err)
		}
		if f.calls != 2 {
			t.Errorf("tb.calls got %d want 2", f.calls)
		}
		if f.lastCall != str {
			t.Errorf("tb.lastCall got %q want %q", f.lastCall, str)
		}
	})
	t.Run("one-at-a-time", func(t *testing.T) {
		const str = "foobar"
		f := &fakeTB{realTB: t}
		writer := &bufferWriter{tb: f}
		wantCalls := 0
		for i, b := range []byte(str + "\n") {
			if _, err := writer.Write([]byte{b}); err != nil {
				t.Fatalf("Failed to write the %dth byte %q: %v", i, b, err)
			}
			if b == '\n' {
				wantCalls++
			}
			if f.calls != wantCalls {
				t.Errorf("%d %q: tb.calls got %d want %d", i, b, f.calls, wantCalls)
			}
		}
		if f.lastCall != str {
			t.Errorf("tb.lastCall got %q want %q", f.lastCall, str)
		}
	})
}
