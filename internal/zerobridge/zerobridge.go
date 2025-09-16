// Custom bridge to pipe zerolog logger entries to log/slog logger.
package zerobridge

import (
	"encoding/json"
	"log/slog"
	"strings"
)

// Container for log/slog logger.
type ZerologToSlogBridge struct {
	Slogger *slog.Logger
}

// Overload zerlolog Write() to conform to log/slog format.
func (b *ZerologToSlogBridge) Write(p []byte) (n int, err error) {
	// Parse the zerolog JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal(p, &logEntry); err != nil {
		// If not JSON, just log as info with raw content
		b.Slogger.Info("zerolog", "raw", strings.TrimSpace(string(p)))
		return len(p), nil
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
		b.Slogger.LogAttrs(nil, slog.LevelDebug, message, attrs...)
	case "info":
		b.Slogger.LogAttrs(nil, slog.LevelInfo, message, attrs...)
	case "warn":
		b.Slogger.LogAttrs(nil, slog.LevelWarn, message, attrs...)
	case "error":
		b.Slogger.LogAttrs(nil, slog.LevelError, message, attrs...)
	default:
		b.Slogger.LogAttrs(nil, slog.LevelInfo, message, attrs...)
	}
	return len(p), nil
}
