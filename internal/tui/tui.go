package tui

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/glenntam/ibtui/internal/service"
	"github.com/glenntam/ibtui/internal/tui/layout"
	"github.com/glenntam/ibtui/internal/typewriter"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/scmhub/ibsync"
	"golang.org/x/term"
)

const (
<<<<<<< Updated upstream
	ibSystemTimeRefreshRate = 500 * time.Millisecond
	displayRefreshRate      = 200 * time.Millisecond
=======
	displayRefreshRate      = 100 * time.Millisecond
>>>>>>> Stashed changes
	minTermWidth            = 48
	minTermHeight           = 22
	logLinesToDisplay       = 10
)

const (
	nofocus = iota
	portfolio
	watchlist
	chart
	algos
	logs
	orders
	trades
)

// TUI is a bubbletea model.
type TUI struct {
	service *service.IBService
	logger  *slog.Logger

	screenWidth  int
	screenHeight int

	logFile   *os.File
	logCursor int64
	logFollow bool
	logHeight int
	logLines  []string

	timezone        *time.Location
	styles          *layout.Styles
	panels          []*layout.Panel
	statusBar       string
	helpBar         string
	prevSelectedTab int
	selectedTab     int

	selectedContract *ibsync.Contract
}

type refreshDisplayMsg time.Time

func (t *TUI) refreshDisplay() tea.Cmd {
	if t.service.ServiceStarted {
		t.statusBar = t.renderStatusBarContent()
		t.panels[portfolio].Content = t.renderPortfolioContent()
		t.panels[watchlist].Content = t.renderWatchlistContent()
		t.panels[chart].Content = t.renderChartContent()
		t.panels[algos].Content = t.renderAlgosContent()
		t.panels[logs].Content = t.renderLogsContent()
		t.panels[orders].Content = t.renderOrdersContent()
		t.panels[trades].Content = t.renderTradesContent()
		t.helpBar = t.renderHelpBarContent()
	}
    return tea.Tick(displayRefreshRate, func(t time.Time) tea.Msg {
          return refreshDisplayMsg(t)
    })
}

func NewTUIApp(service *service.IBService, logger *slog.Logger, logFile *os.File, preferredTimeZone *time.Location) *TUI {
	return &TUI{
		service:  service,
		logger:   logger,
		logFile:  logFile,
		timezone: preferredTimeZone,
	}
}

// Init is called once before the TUI loops. Use it to kick off a cmd.
func (t *TUI) Init() tea.Cmd {
	t.service.StartIBService()
	// Use x/term to set init screen size before Update() runs
	termWidth, termHeight, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		termWidth = minTermWidth
		termHeight = minTermHeight
	}
	t.screenWidth = termWidth
	t.screenHeight = termHeight

	// Init log display
	t.logHeight = logLinesToDisplay
	t.logLines = make([]string, 0)
	t.logFollow = true
	t.logCursor, err = typewriter.GetFileSize(t.logFile)
	if err != nil {
		t.logger.Error("couldn't retrieve log file size", "error", err)
		t.logCursor = 0
	}

	// Get lipgloss styles
	t.styles = layout.DefaultTheme()

	// Arrange initial tab grouping:
	t.panels = append(t.panels, &layout.Panel{
		Index: nofocus,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    portfolio,
		Tab:      "1. Porfolio",
		Content:  "",
		Revealed: false,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    watchlist,
		Tab:      "2. Watchlist",
		Content:  "",
		Revealed: true,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    chart,
		Tab:      "3. Chart",
		Content:  "",
		Revealed: true,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    algos,
		Tab:      "4. Algos",
		Content:  "",
		Revealed: false,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    logs,
		Tab:      "5. Log ",
		Content:  "",
		Revealed: true,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    orders,
		Tab:      "6. Open Orders",
		Content:  "",
		Revealed: false,
	})
	t.panels = append(t.panels, &layout.Panel{
		Index:    trades,
		Tab:      "7. Trade Log",
		Content:  "",
		Revealed: false,
	})
	t.prevSelectedTab = nofocus
	t.selectedTab = nofocus

	// Set initial contract (t.service.Contracts is guaranteed to have at least one)
	t.selectedContract = t.service.Contracts[0]

	return t.refreshDisplay()
}

// Update catches keypresses and screen updates then passes them to View().
func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn
	var err error
	switch msg := msg.(type) {

	// bubbletea always sends tea.WindowSizeMsg on Update()'s first run
	case tea.WindowSizeMsg:
		t.screenWidth, t.screenHeight = msg.Width, msg.Height
		return t, nil

	// Catch Init()'s return func and keep rerunning it. See tea.Tick
	case refreshDisplayMsg:
		return t, t.refreshDisplay()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Gracefully quit
			if t.service.ServiceStarted {
				t.service.StopIBService()
			}
			return t, tea.Quit
		case strconv.Itoa(portfolio):
			if t.selectedTab == portfolio {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = portfolio
				t.panels[portfolio].Revealed = true
				t.panels[watchlist].Revealed = false
			}
		case strconv.Itoa(watchlist):
			if t.selectedTab == watchlist {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = watchlist
				t.panels[portfolio].Revealed = false
				t.panels[watchlist].Revealed = true
			}
		case strconv.Itoa(chart):
			if t.selectedTab == chart {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = chart
				t.panels[chart].Revealed = true
				t.panels[algos].Revealed = false
			}
		case strconv.Itoa(algos):
			if t.selectedTab == algos {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = algos
				t.panels[chart].Revealed = false
				t.panels[algos].Revealed = true
			}
		case strconv.Itoa(logs):
			if t.selectedTab == logs {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = logs
				t.panels[logs].Revealed = true
				t.panels[orders].Revealed = false
				t.panels[trades].Revealed = false
			}
		case strconv.Itoa(orders):
			if t.selectedTab == orders {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = orders
				t.panels[logs].Revealed = false
				t.panels[orders].Revealed = true
				t.panels[trades].Revealed = false
			}
		case strconv.Itoa(trades):
			if t.selectedTab == trades {
				t.selectedTab = nofocus
			} else {
				t.selectedTab = trades
				t.panels[logs].Revealed = false
				t.panels[orders].Revealed = false
				t.panels[trades].Revealed = true
			}
		case "up", "k":
			t.logFollow = false
			t.panels[logs].Tab = "6. Log*"
			t.logCursor, err = typewriter.PrevNewline(t.logFile, t.logCursor)
			if err != nil {
				t.logger.Error("Error getting previous newline", "error", err)
			}
		case "down", "j":
			t.logCursor, err = typewriter.NextNewline(t.logFile, t.logCursor)
			if err != nil {
				t.logger.Error("Error getting next newline", "error", err)
			}
			t.logFollow, err = typewriter.CursorAtEOF(t.logFile, t.logCursor)
			if err != nil {
				t.logger.Error("Couldn't determine if cursor at end of file", "error", err)
			}
			if t.logFollow {
				t.panels[logs].Tab = "6. Log "
			}
		case "G":
			t.logCursor, err = typewriter.GetFileSize(t.logFile)
			if err != nil {
				t.logger.Error("m.logCursor couldn't retrieve log file size", "error", err)
			}
			t.logFollow = true
			t.panels[logs].Tab = "6. Log "
		case "d":
			t.logger.Debug("emit Debug")
		case "i":
			t.logger.Info("emit Info")
		case "w":
			t.logger.Warn("emit Warn")
		case "e":
			t.logger.Error("emit Error")
		}
	}
	return t, t.refreshDisplay()
}

// View gathers the TUI model state and renders the data to screen.
func (t *TUI) View() string {
	top := layout.RenderHorizontalGroup(
		t.panels[portfolio:chart],
		t.styles,
		t.selectedTab,
		t.screenWidth,
	)
	middle := layout.RenderHorizontalGroup(
		t.panels[chart:logs],
		t.styles,
		t.selectedTab,
		t.screenWidth,
	)
	bottom := layout.RenderHorizontalGroup(
		t.panels[logs:],
		t.styles,
		t.selectedTab,
		t.screenWidth,
	)
	return lipgloss.JoinVertical(lipgloss.Left, t.statusBar, top, middle, bottom, t.helpBar)
}

func (t *TUI) renderStatusBarContent() string {
<<<<<<< Updated upstream
	sysTime := t.service.ReadTime()
	preferredTZ := sysTime.In(t.timezone)
=======
	preferredTZ := t.service.SystemTime.In(t.timezone)
>>>>>>> Stashed changes
	fmtTime := fmt.Sprintf("%s (%v)", preferredTZ.Format(time.StampMilli), preferredTZ.Location())
	return t.styles.StatusBar.Render(fmtTime)
}

func (t *TUI) renderPortfolioContent() string {
	return "renderPortfolioContent"
}

func (t *TUI) renderWatchlistContent() string {
	content := ""
<<<<<<< Updated upstream
	tickers := t.service.ReadTickers()
	for _, ticker := range tickers {
=======
	for _, ticker := range t.service.Tickers {
>>>>>>> Stashed changes
		styledPrice := ""
		prev := ticker.PrevLast()
		last := ticker.Last()
		switch {
		case prev < last:
			styledPrice = t.styles.GreenBG.Render(strconv.FormatFloat(last, 'f', 2, 64))
		case prev > last:
			styledPrice = t.styles.RedBG.Render(strconv.FormatFloat(last, 'f', 2, 64))
		default:
			styledPrice = strconv.FormatFloat(last, 'f', -1, 64)
		}
		content += ticker.Contract().LocalSymbol + " " + styledPrice + "  "
		if utf8.RuneCountInString(content)+15 > t.screenWidth {
			content += "\n"
		}
	}
	return content
}

func (t *TUI) renderChartContent() string {
	content := ""
<<<<<<< Updated upstream
	bars := t.service.ReadBars()
	if len(bars) != 0 {
		for _, b := range bars {
			content += fmt.Sprintf("%v %.2f, Vol: %v, BarCount: %v\n", b.Date, b.Close, b.Volume, b.BarCount)
		}
		slog.Debug(strconv.Itoa(len(bars)))
=======
	if len(t.service.Bars) != 0 {
		for _, b := range t.service.Bars {
			content += fmt.Sprintf("%v O:%.2f H:%.2f L:%.2f C:%.2f Vol: %v, BarCount: %v\n",
			b.Date, b.Open, b.High, b.Low, b.Close, b.Volume, b.BarCount)
		}
		slog.Debug(strconv.Itoa(len(t.service.Bars)))
>>>>>>> Stashed changes
	}
	return content
}

func (t *TUI) renderAlgosContent() string {
	return "renderAlgosContent"
}

func (t *TUI) renderLogsContent() string {
	var err error
	content := ""
	if t.logFollow {
		t.logCursor, err = typewriter.GetFileSize(t.logFile)
		if err != nil {
			t.logger.Error("couldn't auto follow log", "error", err)
		}
	}
	t.logLines, _ = typewriter.RenderLog(t.logFile, t.logCursor, t.logHeight, t.screenWidth)
	// if err != nil {
	//     t.logger.Error("Couldn't render log display", "error", err)
	// }
	content = strings.Join(t.logLines, "\n")
	content = strings.TrimRight(content, "\r\n")
	return content
}

func (t *TUI) renderOrdersContent() string {
	return "renderOpenOrdersTab"
}

func (t *TUI) renderTradesContent() string {
	return "renderTradeLogTab"
}

func (t *TUI) renderHelpBarContent() string {
	return "Press ? for help"
}
