// The main entry point of ibtui and starting the TUI.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glenntam/ibtui/internal/env"
	"github.com/glenntam/ibtui/internal/logger"
	"github.com/glenntam/ibtui/internal/smtp"
	"github.com/glenntam/ibtui/internal/state"
	"github.com/glenntam/ibtui/internal/zerobridge"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scmhub/ibsync"
)

const (
	logLinesDisplayed = 10
	logFilePermission = 0o600 // RW for owner only
)

// Assemble ibtui top-level components, including config, logger and tui.
func main() {
	cfg := env.ParseDotEnv()

	// Set up Logger:
	timezone, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		time.Local = time.UTC
	} else {
		time.Local = timezone
	}

	smtp := smtp.NewClient(
		cfg.SMTPPort,
		cfg.SMTPHost,
		cfg.SMTPUsername,
		string(cfg.SMTPPassword),
		cfg.SMTPSender,
		cfg.SMTPRecipient)

	logFile, err := os.OpenFile(cfg.LogFile,
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		logFilePermission,
	)
	if err != nil {
		fmt.Printf("OS Error: Unable to open nor create %v\n", cfg.LogFile)
		os.Exit(1)
	}
	defer closeLogFile(logFile)

	slogger := logger.New(logFile, smtp)
	slog.SetDefault(slogger)
	// (pipe ibsync's internal zerologger to stdlib slog)
	bridge := &zerobridge.ZerologToSlogBridge{Slogger: slogger}
	log.Logger = zerolog.New(bridge).With().Timestamp().Logger()
	ib := ibsync.NewIB()
	ib.SetLogger(log.Logger)
	ib.SetClientLogLevel(1)

	// Set up TUI model:
	ibs := state.NewIBState()
	tui := &model{
		ib:        ib,
		ibs:       ibs,
		timezone:  cfg.Timezone,
		logFile:   logFile,
		logHeight: logLinesDisplayed,
		logFollow: true,
	}

	// Connect to IB API and start TUI:
	ibCfg := ibsync.NewConfig(
		ibsync.WithHost(cfg.Host),
		ibsync.WithPort(cfg.Port),
		ibsync.WithClientID(cfg.ClientID),
	)

	if err = ib.Connect(ibCfg); err != nil {
		slog.Error("Couldn't connect to IB", "error", err)
	}
	defer disconnect(ib)

	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("Couldn't run bubbletea", "error", err)
	}
}

// A deferred cleanup function to close previously opened log file.
func closeLogFile(f *os.File) {
	if f == nil {
		return
	}
	err := f.Close()
	if err != nil {
		if slog.Default() != nil {
			slog.Error("Failed to close log file", "error", err)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to close log file: %v\n", err)
		}
	}
}

// A deferred cleanup function to gracefully disconnect from IB API.
func disconnect(ib *ibsync.IB) {
	if ib == nil {
		return
	}
	err := ib.Disconnect()
	if err != nil {
		if slog.Default() != nil {
			slog.Error("Couldn't disconnect IB", "error", err)
		} else {
			fmt.Fprintf(os.Stderr, "Couldn't disconnect IB: %v\n", err)
		}
	} else {
		slog.Info("Gracefully disconnected")
	}
}
