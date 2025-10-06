package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/glenntam/ibtui/internal/panels"
	"github.com/glenntam/ibtui/internal/state"
	"github.com/scmhub/ibsync"
	"golang.org/x/term"
)

const (
	millisecondRefreshRate = 30

	minTermWidth  = 48
	minTermHeight = 22
)

const (
	nofocus = iota
	portfolio
	watchlist
	quote
	orders
	algos
	logs
	trades
)

// Use this type to catch repeated refreshIBState messages in Update().
type refreshMsg time.Time

// model reflects the current state of the TUI app.
// model.ibs reflects the state of the IB account via continued polling.
type model struct {
	ib       *ibsync.IB
	ibs      *state.IBState
	timezone string

	logFile   *os.File
	logHeight int
	logLines  []string
	logCursor int64
	logFollow bool

	panels          []*panels.Panel
	prevSelectedTab int
	selectedTab     int
	screenWidth     int
	screenHeight    int
}

// Render the Portfolio panel into a string for further Bubbletea rendering.
func (m *model) renderPorfolioContent() string {
	return fmt.Sprintf(
		"%s (%v)",
		m.ibs.CurrentTime.Format(time.StampMilli),
		m.ibs.CurrentTime.Location(),
	)
}

// Render the Watchlist panel into a string for further Bubbletea rendering.
func (m *model) renderWatchlistContent() string {
	return "renderWatchlistTab"
}

// Render the Order Entry panel into a string for further Bubbletea rendering.
func (m *model) renderOrderEntryContent() string {
	return "renderOrderEntryTab"
}

// Render the Open Orders panel into a string for further Bubbletea rendering.
func (m *model) renderOpenOrdersContent() string {
	return "renderOpenOrdersTab"
}

// Render the Algo panel into a string for further Bubbletea rendering.
func (m *model) renderAlgoContent() string {
	return "renderAlgoTab"
}

// Render the Log panel into a string for further Bubbletea rendering.
func (m *model) renderLogContent() string {
	var err error
	m.logFollow, err = panels.CursorAtEOF(m.logFile, m.logCursor)
	if err != nil {
		slog.Error("Couldn't determine if cursor at end of file", "error", err)
	}
	offset := m.logCursor
	if m.logFollow {
		fileInfo, err := m.logFile.Stat()
		if err != nil {
			slog.Warn("Couldn't stat log file during log tab render")
		} else {
			offset = fileInfo.Size()
		}
	}
	m.logLines, err = panels.RenderLog(m.logFile, offset, m.logHeight, m.screenWidth)
	if err != nil {
		slog.Error("Couldn't render log display", "error", err)
	}
	str := strings.Join(m.logLines, "\n")
	str = strings.TrimRight(str, "\r\n")

	return str
}

// Render the completed Trades panel into a string.
func (m *model) renderTradeLogContent() string {
	return "renderTradeLogTab"
}

// Catchall function to gather IB account state, update
// TUI model fields, and then set itself to repeat.
func (m *model) refreshIBState() tea.Cmd {
	// Portfolio tab:
	err := m.ibs.ReqCurrentTimeMilli(m.ib)
	if err != nil {
		slog.Error("Couldn't get time from IB API", "error", err)
	}

	// Log tab:
	if m.logFollow {
		m.logCursor, err = panels.GetFileSize(m.logFile)
		if err != nil {
			slog.Error("m.logCursor couldn't retrieve file size", "error", err)
		}
		m.panels[logs].Tab = "6. Log "
	} else {
		m.panels[logs].Tab = "6. Log*"
	}

	// Render All tabs:
	m.panels[portfolio].Content = m.renderPorfolioContent()
	m.panels[watchlist].Content = m.renderWatchlistContent()
	m.panels[quote].Content = m.renderOrderEntryContent()
	m.panels[orders].Content = m.renderOpenOrdersContent()
	m.panels[algos].Content = m.renderAlgoContent()
	m.panels[logs].Content = m.renderLogContent()
	m.panels[trades].Content = m.renderTradeLogContent()

	// Re-run timer:
	return tea.Batch(
		tea.Tick(millisecondRefreshRate*time.Millisecond, func(t time.Time) tea.Msg {
			return refreshMsg(t)
		}),
	)
}

// Called once at init before the TUI loops. Use it to kick off a cmd.
func (m *model) Init() tea.Cmd {
	var err error
	m.logCursor, err = panels.GetFileSize(m.logFile)
	if err != nil {
		slog.Error("m.logCursor couldn't retrieve log file size", "error", err)
	}

	// Use x/term to temporarily get init screen width/height before passing to TUI:
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = minTermWidth
		termHeight = minTermHeight
	}
	m.screenWidth = termWidth
	m.screenHeight = termHeight

	// Initialize panels:
	m.panels = append(m.panels, &panels.Panel{
		Index: nofocus,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    portfolio,
		Tab:      "1. Porfolio",
		Content:  m.renderPorfolioContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    watchlist,
		Tab:      "2. Watchlist",
		Content:  m.renderWatchlistContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    quote,
		Tab:      "3. Quote / Order Entry",
		Content:  m.renderOrderEntryContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    orders,
		Tab:      "4. Open Orders",
		Content:  m.renderOpenOrdersContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    algos,
		Tab:      "5. Algos",
		Content:  m.renderAlgoContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    logs,
		Tab:      "6. Log ",
		Content:  m.renderLogContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index:    trades,
		Tab:      "7. Trade Log",
		Content:  m.renderTradeLogContent(),
		Revealed: false,
	})
	m.prevSelectedTab = nofocus
	m.selectedTab = nofocus
	slog.Info("TUI initializing")
	return m.refreshIBState()
}

// Catch keypresses and screen updates here, then pass to View().
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn
	var err error
	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case strconv.Itoa(portfolio):
			if m.selectedTab == portfolio {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = portfolio
				m.panels[portfolio].Revealed = true
				m.panels[watchlist].Revealed = false
			}
		case strconv.Itoa(watchlist):
			if m.selectedTab == watchlist {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = watchlist
				m.panels[portfolio].Revealed = false
				m.panels[watchlist].Revealed = true
			}
		case strconv.Itoa(quote):
			if m.selectedTab == quote {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = quote
				m.panels[quote].Revealed = true
				m.panels[orders].Revealed = false
				m.panels[algos].Revealed = false
			}
		case strconv.Itoa(orders):
			if m.selectedTab == orders {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = orders
				m.panels[quote].Revealed = false
				m.panels[orders].Revealed = true
				m.panels[algos].Revealed = false
			}
		case strconv.Itoa(algos):
			if m.selectedTab == algos {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = algos
				m.panels[quote].Revealed = false
				m.panels[orders].Revealed = false
				m.panels[algos].Revealed = true
			}
		case strconv.Itoa(logs):
			if m.selectedTab == logs {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = logs
				m.panels[logs].Revealed = true
				m.panels[trades].Revealed = false
			}
		case strconv.Itoa(trades):
			if m.selectedTab == trades {
				m.selectedTab = nofocus
			} else {
				m.selectedTab = trades
				m.panels[logs].Revealed = false
				m.panels[trades].Revealed = true
			}
		case "up", "k":
			m.logFollow = false
			m.logCursor, err = panels.PrevNewline(m.logFile, m.logCursor)
			if err != nil {
				slog.Error("Error getting previous newline", "error", err)
			}
		case "down", "j":
			m.logCursor, err = panels.NextNewline(m.logFile, m.logCursor)
			if err != nil {
				slog.Error("Error getting next newline", "error", err)
			}
			m.logFollow, err = panels.CursorAtEOF(m.logFile, m.logCursor)
			if err != nil {
				slog.Error("Couldn't determine if cursor at end of file", "error", err)
			}
		case "G":
			m.logCursor, err = panels.GetFileSize(m.logFile)
			if err != nil {
				slog.Error("m.logCursor couldn't retrieve log file size", "error", err)
			}
			m.logFollow = true
		case "d":
			slog.Debug("emit Debug")
		case "i":
			slog.Info("emit Info")
		case "w":
			slog.Warn("emit Warn")
		case "e":
			slog.Error("emit Error")
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.screenWidth = v.Width
		m.screenHeight = v.Height
		return m, nil
	case refreshMsg:
		return m, m.refreshIBState()
	}
	return m, nil
}

// Based on the TUI model state, render the data to screen.
func (m *model) View() string {
	top := panels.RenderHorizontalGroup(
		m.panels[portfolio:quote],
		m.selectedTab,
		m.screenWidth,
	)
	mid := panels.RenderHorizontalGroup(
		m.panels[quote:logs],
		m.selectedTab,
		m.screenWidth,
	)
	bot := panels.RenderHorizontalGroup(
		m.panels[logs:],
		m.selectedTab,
		m.screenWidth,
	)
	status := panels.RenderStatusLine("STATUS LINE")

	return lipgloss.JoinVertical(lipgloss.Left, top, mid, bot, status)
}
