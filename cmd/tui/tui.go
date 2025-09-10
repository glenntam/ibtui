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
)

type (
	refreshMsg     time.Time
)

var (
	// Active tab style - this is the currently selected tab
	activeTabBorderStyle = lipgloss.NewStyle().
				Bold(true).                                        // Make text bold
				Foreground(lipgloss.Color("#FFFFFF")).            // White text
				Background(lipgloss.Color("#7D56F4")).            // Purple background
				Border(lipgloss.RoundedBorder(), true, true, false, true). // Border: top=yes, right=yes, bottom=NO, left=yes
				BorderForeground(lipgloss.Color("#7D56F4")).       // Purple border color
				Padding(0, 1)                                     // No vertical padding, 1 space horizontal

	// Inactive tab style - tabs that are not selected
	inactiveTabBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).             // Gray text
				Background(lipgloss.Color("#1A1A1A")).             // Dark background
				Border(lipgloss.RoundedBorder(), true, true, false, true). // Same border pattern
				BorderForeground(lipgloss.Color("#333333")).       // Dark gray border
				Padding(0, 1)                                     // Same padding

	// Content panel style - the main content area below tabs
	tabContentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).              // Full border around content
			BorderForeground(lipgloss.Color("#7D56F4")).   // Purple border (matches active tab)
			Padding(0, 1)                                 // 1 line vertical, 2 spaces horizontal padding
			//MinHeight(15)                                  // Minimum height for content area

	// Status line style - help text at the bottom
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).         // Gray color for status text
			Italic(true)                                   // Make it italic
)

type model struct {
	ib          *ibsync.IB
	ibs         *state.IBState
	timezone    string
	currentTime string

	logFile     *os.File
	logLines    []string
	logCursor   int64
	logFollow   bool

	lastUpdate  time.Time
	panelGroups []*panelGroup
	selectedTab int
	screenWidth  int
	screenHeight int
}

type panelGroup struct {
	tabs        []string
	content     []string
	activeTab   int
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
	m.logLines = panels.RenderLog(m.logFile, offset, 10)
	str := strings.Join(m.logLines, "")
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
		m.panelGroups[2].tabs[0] = "6. Log"
	} else {
	m.panelGroups[2].tabs[0] = "6. Log*"
	}
	// render all tabs
	m.panelGroups[0].content[0] = m.renderPorfolioTab()
	m.panelGroups[0].content[1] = m.renderWatchlistTab()
	m.panelGroups[1].content[0] = m.renderOrderEntryTab()
	m.panelGroups[1].content[1] = m.renderOpenOrdersTab()
	m.panelGroups[1].content[2] = m.renderAITab()
	m.panelGroups[2].content[0] = m.renderLogTab()
	m.panelGroups[2].content[1] = m.renderTradeLogTab()

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

	topGroup := &panelGroup{
		tabs:    []string{"1. Portfolio", "2. Watchlist",},
		content: []string{m.renderPorfolioTab(), m.renderWatchlistTab()},
	}
	midGroup := &panelGroup{
		tabs:    []string{"3. Quote / Order Entry", "4. Open Orders", "5. AI"},
		content: []string{m.renderOrderEntryTab(), m.renderOpenOrdersTab(), m.renderAITab()},
	}
	botGroup := &panelGroup{
		tabs:    []string{"6. Log", "7. Trade Log"},
		content: []string{m.renderLogTab(), m.renderTradeLogTab()},
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
			m.panelGroups[0].activeTab = 0
		case "2":
			m.selectedTab = 2
			m.panelGroups[0].activeTab = 1
		case "3":
			m.selectedTab = 3
			m.panelGroups[1].activeTab = 0
		case "4":
			m.selectedTab = 4
			m.panelGroups[1].activeTab = 1
		case "5":
			m.selectedTab = 5
			m.panelGroups[1].activeTab = 2
		case "6":
			m.selectedTab = 6
			m.panelGroups[2].activeTab = 0
		case "7":
			m.selectedTab = 7
			m.panelGroups[2].activeTab = 1
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

func (m *model) renderPanelGroup(g *panelGroup) string {
	var tabRow []string
	for i, tab := range g.tabs {
		if i == g.activeTab {
			tabRow = append(tabRow, activeTabBorderStyle.Render(tab))
		} else {
			tabRow = append(tabRow, inactiveTabBorderStyle.Render(tab))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	body := tabContentStyle.
		Width(m.screenWidth - 2).      // Account for border
		Render(g.content[g.activeTab]) // Show content for current tab
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, body)
}

func (m *model) View() string {
	if m.screenWidth == 0 || m.screenHeight == 0 {
		return "loading…"
	}
	top := m.renderPanelGroup(m.panelGroups[0])
	mid := m.renderPanelGroup(m.panelGroups[1])
	bot := m.renderPanelGroup(m.panelGroups[2])

	return lipgloss.JoinVertical(
		lipgloss.Left,
		top,
		mid,
		bot,
	)
}
