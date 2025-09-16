package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"
	"golang.org/x/term"

	"github.com/glenntam/ibtui/internal/panels"
	"github.com/glenntam/ibtui/internal/state"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scmhub/ibsync"
)

type refreshMsg     time.Time

type model struct {
	ib          *ibsync.IB
	ibs         *state.IBState
	timezone    string
	currentTime string

	logFile   *os.File
	logHeight int
	logLines  []string
	logCursor int64
	logFollow bool

	lastUpdate   time.Time
	panelGroups  []*panels.PanelGroup
	selectedTab  int
	screenWidth  int
	screenHeight int
}

func (m *model) renderPorfolioTab() string {
	return fmt.Sprintf("%s (%v)", m.ibs.CurrentTime.Format(time.StampMilli), m.ibs.CurrentTime.Location())
}

func (m *model) renderWatchlistTab() string {
	return "renderWatchlistTab"
}

func (m *model) renderOrderEntryTab() string {
	return "renderOrderEntryTab"
}

func (m *model) renderOpenOrdersTab() string {
	return "renderOpenOrdersTab"
}

func (m *model) renderAITab() string {
	return "renderAITab"
}

func (m *model) renderLogTab() string {
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
	m.logLines = panels.RenderLog(m.logFile, offset, 10, m.screenWidth - 4)

	str := strings.Join(m.logLines, "\n")
	strings.TrimRight(str, "\r\n")

	return str
}

func (m *model) renderTradeLogTab() string {
	return "renderTradeLogTab"
}

// Get the state of IB account and update model fields
func (m *model) refreshIBState() tea.Cmd {
	// portfolio tab
	m.ibs.ReqCurrentTimeMilli(m.ib, m.timezone)
	// log tab
	if m.logFollow == true {
		m.logCursor = panels.GetFileSize(m.logFile)
		m.panelGroups[2].Tabs[0] = "6. Log"
	} else {
	m.panelGroups[2].Tabs[0] = "6. Log*"
	}
	// render all tabs
	m.panelGroups[0].Content[0] = m.renderPorfolioTab()
	m.panelGroups[0].Content[1] = m.renderWatchlistTab()
	m.panelGroups[1].Content[0] = m.renderOrderEntryTab()
	m.panelGroups[1].Content[1] = m.renderOpenOrdersTab()
	m.panelGroups[1].Content[2] = m.renderAITab()
	m.panelGroups[2].Content[0] = m.renderLogTab()
	m.panelGroups[2].Content[1] = m.renderTradeLogTab()

	return tea.Batch(
		tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return refreshMsg(t)
		}),
	)
}

func (m *model) Init() tea.Cmd {
	fileSize := panels.GetFileSize(m.logFile)
	m.logCursor = fileSize
	m.logFollow = true
	m.logHeight = 10
	// Use x/term to temprarily get init screen width
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = 48
		termHeight = 22
	}
	m.screenWidth = termWidth
	m.screenHeight = termHeight

	topGroup := &panels.PanelGroup{
		Tabs:    []string{"1. Portfolio", "2. Watchlist",},
		Content: []string{m.renderPorfolioTab(), m.renderWatchlistTab()},
	}
	midGroup := &panels.PanelGroup{
		Tabs:    []string{"3. Quote / Order Entry", "4. Open Orders", "5. AI"},
		Content: []string{m.renderOrderEntryTab(), m.renderOpenOrdersTab(), m.renderAITab()},
	}
	botGroup := &panels.PanelGroup{
		Tabs:    []string{"6. Log", "7. Trade Log"},
		Content: []string{m.renderLogTab(), m.renderTradeLogTab()},
	}
	m.panelGroups = append(m.panelGroups, topGroup, midGroup, botGroup)
	slog.Info("TUI started")
	return m.refreshIBState()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		switch v.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.selectedTab = 1
			m.panelGroups[0].ActiveTab = 0
		case "2":
			m.selectedTab = 2
			m.panelGroups[0].ActiveTab = 1
		case "3":
			m.selectedTab = 3
			m.panelGroups[1].ActiveTab = 0
		case "4":
			m.selectedTab = 4
			m.panelGroups[1].ActiveTab = 1
		case "5":
			m.selectedTab = 5
			m.panelGroups[1].ActiveTab = 2
		case "6":
			m.selectedTab = 6
			m.panelGroups[2].ActiveTab = 0
		case "7":
			m.selectedTab = 7
			m.panelGroups[2].ActiveTab = 1
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
		m.screenHeight = v.Height
		return m, nil
	case refreshMsg:
		return m, m.refreshIBState()
	}
	return m, nil
}

func (m *model) View() string {
	if m.screenWidth == 0 || m.screenHeight == 0 {
		return "loading…"
	}
	top := panels.RenderPanelGroup(m.panelGroups[0], m.screenWidth)
	mid := panels.RenderPanelGroup(m.panelGroups[1], m.screenWidth)
	bot := panels.RenderPanelGroup(m.panelGroups[2], m.screenWidth)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		top,
		mid,
		bot,
		"STATUS LINE",
	)
}
