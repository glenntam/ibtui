package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/glenntam/ibtui/internal/env"
	"github.com/glenntam/ibtui/internal/logger"
	"github.com/glenntam/ibtui/internal/smtp"
	"github.com/glenntam/ibtui/internal/state"
	// "github.com/glenntam/ibtui/internal/wrapper"
	"github.com/glenntam/ibtui/internal/zerobridge"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/scmhub/ibsync"
)

func main() {
	cfg := env.ParseDotEnv()

	smtp := smtp.NewSMTPClient(cfg.SMTPPort,
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
	slog.Info("DEFAULT SLOG SET")
	// Pipe ibsync's internal zerologger to stdlib slog
	bridge := &zerobridge.ZerologToSlogBridge{Slogger: slogger}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = zerolog.New(bridge).With().Timestamp().Logger()
	zerolog.DefaultContextLogger = &log.Logger
	slog.Info("BRIDGED")

	ib := ibsync.NewIB()
	ibCfg := ibsync.NewConfig(ibsync.WithHost(cfg.Host), ibsync.WithPort(cfg.Port), ibsync.WithClientID(cfg.ClientID))
	if err = ib.Connect(ibCfg); err != nil {
		slog.Error("Couldn't connect to IB", "error", err)
	}
	defer cleanup(ib)
	ibs := state.NewIBState()
	m := &model{
		ib:        ib,
		ibs:       ibs,
		timezone:  cfg.Timezone,
		logFile:   logFile,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		slog.Error("Couldn't run bubbletea", "error", err)
	}
	return
}

func cleanup(ib *ibsync.IB) {
	if err := ib.Disconnect(); err != nil {
		slog.Error("Couldn't disconnect IB", "error", err)
	} else {
		slog.Info("Gracefully disconnected")
	}
	return
}
