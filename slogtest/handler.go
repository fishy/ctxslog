package slogtest

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"go.yhsif.com/ctxslog"
)

type bufferWriter struct {
	tb  testing.TB
	buf bytes.Buffer
}

func (bw *bufferWriter) Write(p []byte) (n int, err error) {
	defer func() {
		for bw.buf.Len() > 0 {
			str, err := bw.buf.ReadString('\n')
			if err != nil {
				// Not a full line yet, return the string back and end this for loop
				bw.buf.WriteString(str)
				return
			}
			str = str[:len(str)-1] // remove the final '\n'
			bw.tb.Log(str)
		}
	}()
	return bw.buf.Write(p)
}

type handler struct {
	slog.Handler

	tb     testing.TB
	failAt slog.Leveler
}

func (h handler) Handle(ctx context.Context, r slog.Record) error {
	defer func() {
		if l := r.Level; l >= h.failAt.Level() {
			h.tb.Errorf("slog called at %v level with %q", l, r.Message)
		}
	}()
	return h.Handler.Handle(ctx, r)
}

// Handler returns a *slog.Handler that fails the test when logged at failAt
// level, and logs everything at min level (both inclusive).
func Handler(tb testing.TB, min, failAt slog.Leveler) slog.Handler {
	h := ctxslog.CallstackHandler(
		slog.NewTextHandler(
			&bufferWriter{
				tb: tb,
			},
			&slog.HandlerOptions{
				AddSource: true,
				Level:     min,
			},
		),
		min,
	)
	return handler{
		Handler: h,

		tb:     tb,
		failAt: failAt,
	}
}
