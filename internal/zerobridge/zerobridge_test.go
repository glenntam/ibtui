package zerobridge

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
)

func TestZerologToSlogBridge_Write(t *testing.T) {
	t.Run("nil Slogger returns ErrNilSlogger", func(t *testing.T) {
		b := &ZerologToSlogBridge{Slogger: nil}
		data := []byte(`{"level":"info","message":"hello"}`)
		n, err := b.Write(data)
		if !errors.Is(err, ErrNilSlogger) {
			t.Fatalf("expected ErrNilSlogger, got %v", err)
		}
		if n != len(data) {
			t.Fatalf("expected n=%d got %d", len(data), n)
		}
	})

	t.Run("valid Slogger logs without error", func(t *testing.T) {
		var buf bytes.Buffer
		handler := slog.NewTextHandler(&buf, nil)
		logger := slog.New(handler)
		b := &ZerologToSlogBridge{Slogger: logger}

		jsonLog := []byte(`{"level":"info","message":"hello","foo":"bar"}`)
		n, err := b.Write(jsonLog)
		if err != nil {
			t.Fatalf("Write returned unexpected error: %v", err)
		}
		if n != len(jsonLog) {
			t.Fatalf("expected n=%d got %d", len(jsonLog), n)
		}
		if got := buf.String(); got == "" {
			t.Fatalf("expected non-empty slog output")
		}
	})

	t.Run("malformed JSON still returns error but consumes all bytes", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, nil))
		b := &ZerologToSlogBridge{Slogger: logger}

		data := []byte(`{invalid json`)
		n, err := b.Write(data)
		if err == nil {
			t.Fatalf("expected error for malformed JSON")
		}
		if n != len(data) {
			t.Fatalf("expected n=%d got %d", len(data), n)
		}
	})
}
