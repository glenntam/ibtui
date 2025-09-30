package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/glenntam/ibtui/internal/panels"
	"github.com/glenntam/ibtui/internal/state"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scmhub/ibsync"
	"golang.org/x/term"
)

// Use this type to catch repeated refreshIBState messages in Update().
type refreshMsg time.Time

// model reflects the current state of the TUI app.
// model.ibs reflects the state of the IB account via continued polling.
type model struct {
	ib           *ibsync.IB
	ibs          *state.IBState
	timezone     string
	currentTime  string

	logFile      *os.File
	logHeight    int
	logLines     []string
	logCursor    int64
	logFollow    bool

	lastUpdate   time.Time
	panels       []*panels.Panel
	selectedTab  int
	screenWidth  int
	screenHeight int
}

// Render the Portfolio panel into a string for further Bubbletea rendering.
func (m *model) renderPorfolioContent() string {
	return fmt.Sprintf("%s (%v)", m.ibs.CurrentTime.Format(time.StampMilli), m.ibs.CurrentTime.Location())
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
	if panels.CursorAtEOF(m.logFile, m.logCursor) == true {
		m.logFollow = true
	}
	var offset int64 = m.logCursor
	if m.logFollow == true {
		fileInfo, err := m.logFile.Stat()
		if err != nil {
			slog.Warn("Couldn't stat log file during log tab render")
		} else {
			offset = fileInfo.Size()
		}
	}
	m.logLines = panels.RenderLog(m.logFile, offset, 10, m.screenWidth-4)

	str := strings.Join(m.logLines, "\n")
	strings.TrimRight(str, "\r\n")

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
	m.ibs.ReqCurrentTimeMilli(m.ib, m.timezone)

	// Log tab:
	if m.logFollow == true {
		m.logCursor = panels.GetFileSize(m.logFile)
		m.panels[5].Tab = "6. Log"
	} else {
		m.panels[5].Tab = "6. Log*"
	}

	// Render All tabs:
	m.panels[0].Content = m.renderPorfolioContent()
	m.panels[1].Content = m.renderWatchlistContent()
	m.panels[2].Content = m.renderOrderEntryContent()
	m.panels[3].Content = m.renderOpenOrdersContent()
	m.panels[4].Content = m.renderAlgoContent()
	m.panels[5].Content = m.renderLogContent()
	m.panels[6].Content = m.renderTradeLogContent()

	// Re-run timer:
	return tea.Batch(
		tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
			return refreshMsg(t)
		}),
	)
}

// Called once at init before the TUI loops. Use it to kick off a cmd.
func (m *model) Init() tea.Cmd {
	logFileSize := panels.GetFileSize(m.logFile)
	m.logCursor = logFileSize

	// Use x/term to temporarily get init screen width/height before passing to TUI:
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 48
		termHeight = 22
	}
	m.screenWidth = termWidth
	m.screenHeight = termHeight

	// Initialize panels:
	m.panels = append(m.panels, &panels.Panel{
		Index: 1,
		Tab: "1. Porfolio",
		Content: m.renderPorfolioContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 2,
		Tab: "2. Watchlist",
		Content: m.renderWatchlistContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 3,
		Tab: "3. Quote / Order Entry",
		Content: m.renderOrderEntryContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 4,
		Tab: "4. Open Orders",
		Content: m.renderOpenOrdersContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 5,
		Tab: "5. Algos",
		Content: m.renderAlgoContent(),
		Revealed: false,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 6,
		Tab: "6. Log",
		Content: m.renderLogContent(),
		Revealed: true,
	})
	m.panels = append(m.panels, &panels.Panel{
		Index: 7,
		Tab: "7. Trade Log",
		Content: m.renderTradeLogContent(),
		Revealed: false,
	})
	m.selectedTab = 0
	slog.Info("TUI initializing")
	return m.refreshIBState()
}

// Catch keypresses and screen updates here, then pass to View().
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.selectedTab = 1
			m.panels[0].Revealed = true
			m.panels[1].Revealed = false
		case "2":
			m.selectedTab = 2
			m.panels[0].Revealed = false
			m.panels[1].Revealed = true
		case "3":
			m.selectedTab = 3
			m.panels[2].Revealed = true
			m.panels[3].Revealed = false
			m.panels[4].Revealed = false
		case "4":
			m.selectedTab = 4
			m.panels[2].Revealed = false
			m.panels[3].Revealed = true
			m.panels[4].Revealed = false
		case "5":
			m.selectedTab = 5
			m.panels[2].Revealed = false
			m.panels[3].Revealed = false
			m.panels[4].Revealed = true
		case "6":
			m.selectedTab = 6
			m.panels[5].Revealed = true
			m.panels[6].Revealed = false
		case "7":
			m.selectedTab = 7
			m.panels[5].Revealed = false
			m.panels[6].Revealed = true
		case "up", "k":
			m.logFollow = false
			m.logCursor = panels.PrevNewline(m.logFile, m.logCursor)
		case "down", "j":
			m.logCursor = panels.NextNewline(m.logFile, m.logCursor)
			if panels.CursorAtEOF(m.logFile, m.logCursor) == true {
				m.logFollow = true
			}
		case "G":
			m.logCursor = panels.GetFileSize(m.logFile)
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
		m.screenHeight =  v.Height
		return m, nil
	case refreshMsg:
		return m, m.refreshIBState()
	}
	return m, nil
}

// Based on the TUI model state, render the data to screen.
func (m *model) View() string {
	if m.screenWidth == 0 || m.screenHeight == 0 {
		return "loading…"
	}
	top := panels.RenderHorizontalGroup(m.panels[:2], m.selectedTab, m.screenWidth)
	mid := panels.RenderHorizontalGroup(m.panels[2:5], m.selectedTab, m.screenWidth)
	bot := panels.RenderHorizontalGroup(m.panels[5:], m.selectedTab, m.screenWidth)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		top,
		mid,
		bot,
		"STATUS LINE",
	)
}
