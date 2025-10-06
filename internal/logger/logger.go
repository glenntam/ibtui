// Package logger is a custom log/slog file and email logger.
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

// SMTPHandler passes minLevel and above slog messages to the smtp client.
type SMTPHandler struct {
	minLevel slog.Level
	smtp     *smtp.Client
}

// Enabled determines if a slog message will be passed to the smtp client.
func (h *SMTPHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel
}

// Handle the emailing of the slog message.
func (h *SMTPHandler) Handle(_ context.Context, r slog.Record) error {
	if r.Level < h.minLevel {
		return nil
	}

	var buf bytes.Buffer
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&buf, "%s=%v ", a.Key, a.Value)
		return true
	})

	msg := fmt.Sprintf(
		"Level: %s\nTime: %s\nMessage: %s\nAttributes: %s",
		r.Level.String(),
		r.Time.Format(time.RFC3339),
		r.Message,
		buf.String(),
	)
	err := h.smtp.Send("Log Alert", msg, h.smtp.Recipient)
	if err != nil {
		return fmt.Errorf("logger couldn't send email: %w", err)
	}
	return nil
}

// WithAttrs satisfies handler interface.
func (h *SMTPHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h
}

// WithGroup satisfies handler interface.
func (h *SMTPHandler) WithGroup(_ string) slog.Handler {
	return h
}

// MultiHandler can contain multiple different types of logging handlers.
// In the case of ibtui, it is a JSON file logger and (optional) smtp emailer.
type MultiHandler struct {
	handlers []slog.Handler
}

// Enabled determines if a slog message will be processed.
func (m *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle determines how a slog message will be processed.
func (m *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

// WithAttrs satisfies handler interface.
func (m *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &MultiHandler{handlers: hs}
}

// WithGroup satisfies handler interface.
func (m *MultiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &MultiHandler{handlers: hs}
}

// New creates a custom multi-logger that implements log/slog.Logger
//
// DEBUG and above:   Save to log file.
// WARNING and above: Email (but only if smtp.AdminEmail is provided).
func New(file *os.File, smtpClient *smtp.Client) *slog.Logger {
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
		multi.handlers = []slog.Handler{fileHandler, smtpHandler}
	} else {
		multi.handlers = []slog.Handler{fileHandler}
	}
	slog.SetDefault(slog.New(multi))
	return slog.New(multi)
}
