// The main entry point of ibtui and starting the TUI.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/glenntam/ibtui/internal/env"
	"github.com/glenntam/ibtui/internal/logger"
	"github.com/glenntam/ibtui/internal/smtp"
	"github.com/glenntam/ibtui/internal/state"
	"github.com/glenntam/ibtui/internal/zerobridge"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scmhub/ibsync"
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

	smtp := smtp.NewSMTPClient(
		cfg.SMTPPort,
		cfg.SMTPHost,
		cfg.SMTPUsername,
		string(cfg.SMTPPassword),
		cfg.SMTPSender,
		cfg.SMTPRecipient)

	logFile, err := os.OpenFile(cfg.LogFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o666)
	if err != nil {
		fmt.Println("OS Error: Unable to open nor create %v\n", cfg.LogFile)
		os.Exit(1)
	}
	defer logFile.Close()

	slogger, err := logger.New(logFile, smtp)
	if err != nil {
		slog.Error("Couldn't start new slogger", "error", err)
	}
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
		logHeight: 10,
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
	defer cleanup(ib)

	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("Couldn't run bubbletea", "error", err)
	}
	return
}

// A deferred cleanup function to gracefully disconnect from IB API.
func cleanup(ib *ibsync.IB) {
	if err := ib.Disconnect(); err != nil {
		slog.Error("Couldn't disconnect IB", "error", err)
	} else {
		slog.Info("Gracefully disconnected")
	}
	return
}
