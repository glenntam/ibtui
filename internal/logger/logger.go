package logger

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/glenntam/ibtui/internal/smtp"
)

// A simple SMTP logging handler that includes an SMTP client
type SMTPHandler struct {
	minLevel slog.Level
	smtp     *smtp.SMTPClient
}

func (h *SMTPHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

func (h *SMTPHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level < h.minLevel {
		return nil
	}

	var buf bytes.Buffer
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&buf, "%s=%v ", a.Key, a.Value)
		return true
	})

	msg := fmt.Sprintf("Level: %s\nTime: %s\nMessage: %s\nAttributes: %s",
		r.Level.String(), r.Time.Format(time.RFC3339), r.Message, buf.String())

	return h.smtp.Send("Log Alert", msg, h.smtp.Recipient)
}

func (h *SMTPHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *SMTPHandler) WithGroup(name string) slog.Handler {
	return h
}

// A logging handler that can contain multiple different type of handlers
type MultiHandler struct {
	handlers []slog.Handler
}

func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: hs}
}

func (m *MultiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: hs}
}

// Create a custom multi-logger that implements log/slog.Logger
//
// DEBUG and above:   Output to standard output.
// INFO and above:    Save to log file.
// WARNING and above: Email (but only if smtp.AdminEmail is provided).
func New(file *os.File, smtpClient *smtp.SMTPClient) (*slog.Logger, error) {
	// stdoutHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	//     Level: slog.LevelDebug,
	// })
	fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	smtpHandler := &SMTPHandler{
		minLevel: slog.LevelWarn,
		smtp:     smtpClient,
	}

	multi := &MultiHandler{}
	if smtpHandler.smtp.Recipient != "" {
		// multi.handlers = []slog.Handler{stdoutHandler, fileHandler, smtpHandler}
		multi.handlers = []slog.Handler{fileHandler, smtpHandler}
	} else {
		// multi.handlers =  []slog.Handler{stdoutHandler, fileHandler}
		multi.handlers = []slog.Handler{fileHandler}
	}
	slog.SetDefault(slog.New(multi))
	return slog.New(multi), nil
}
