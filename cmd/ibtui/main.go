package main

import (
	"log/slog"

	"github.com/glenntam/ibtui/internal/service"
	"github.com/glenntam/ibtui/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/glenntam/multislog"
	"github.com/scmhub/ibsync"
)

const (
	ibPort      = 4001
	ibHost      = "localhost"
	ibClientID  = 0
	logFilename = "logfile.json"
	preferredTZ = "America/New_York"
)

// Assemble ibtui top-level components, including logger, service manager and tui.
func main() {
	var err error
	// Structured Logger
	msl := multislog.New(
		multislog.EnableTimezone(preferredTZ),
		multislog.EnableLogFile(slog.LevelDebug, logFilename, true, true),
	)
	defer msl.Close()

	// IB service
	contracts := []*ibsync.Contract{
		ibsync.NewFuture("MNQ", "202606", "CME", "2", "USD"),
		ibsync.NewFuture("MES", "202606", "CME", "5", "USD"),
	}
	ibs, err := service.NewIBService(ibHost, ibPort, ibClientID, contracts, msl.Logger)
	if err != nil {
		panic(nil)
	}

	// TUI
	tui := tui.NewTUIApp(ibs, msl.Logger, msl.LogFile, msl.Timezone) // TUI needs references to the actual logfile and
	                                                                 // preferred timezone. These may or may not be
	                                                                 // provided by your logger if you don't use multislog.
	p := tea.NewProgram(tui, tea.WithAltScreen())
	_, err = p.Run()
	if err != nil {
		panic(err)
	}
}
