// Package zerobridge pipes zerolog logger entries to log/slog logger.
package zerobridge

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

// ErrNilSlogger occurs when nil is assigned to something where *log/slog.Logger is expected.
var ErrNilSlogger = errors.New("nil passed to Slogger field in ZerologToSlogBridge struct")

// ZerologToSlogBridge contains a log/slog logger.
type ZerologToSlogBridge struct {
	Slogger *slog.Logger
}

// Write overloads zerlolog's original Write to conform to log/slog format.
func (b *ZerologToSlogBridge) Write(p []byte) (int, error) {
	// nil guard
	if b.Slogger == nil {
		return len(p), ErrNilSlogger
	}
	// Parse the zerolog JSON output
	var logEntry map[string]interface{}
	err := json.Unmarshal(p, &logEntry)
	if err != nil { // If not JSON, just log as info with raw content
		return len(p), fmt.Errorf("couldn't unmarshal zerolog message (%v): %w", strings.TrimSpace(string(p)), err)
	}
	// Extract standard fields
	level, _ := logEntry["level"].(string)
	message, _ := logEntry["message"].(string)
	// Convert all other fields to slog attributes
	var attrs []slog.Attr
	for k, v := range logEntry {
		if k != "level" && k != "message" && k != "time" && k != "errorTime" {
			attrs = append(attrs, slog.Any(k, v))
		}
	}
	// Convert zerolog level to slog levels
	switch level {
	case "debug":
		b.Slogger.LogAttrs(context.TODO(), slog.LevelDebug, message, attrs...)
	case "info":
		b.Slogger.LogAttrs(context.TODO(), slog.LevelInfo, message, attrs...)
	case "warn":
		b.Slogger.LogAttrs(context.TODO(), slog.LevelWarn, message, attrs...)
	case "error":
		b.Slogger.LogAttrs(context.TODO(), slog.LevelError, message, attrs...)
	default:
		b.Slogger.LogAttrs(context.TODO(), slog.LevelInfo, message, attrs...)
	}
	return len(p), nil
}
