package panels

import "github.com/charmbracelet/lipgloss"

const (
	colorSelected = lipgloss.Color("5") // Purple
	colorDimmed   = lipgloss.Color("8") // Dark gray
	bordersWidth  = 2
)

// Styles contains styling to render tabs and content.
type Styles struct {
	baseContent     lipgloss.Style
	activeContent   lipgloss.Style
	inactiveContent lipgloss.Style
	activeTab       lipgloss.Style
	inactiveTab     lipgloss.Style
	tabOpen         lipgloss.Style
	tabHidden       lipgloss.Style
	firstTabOpen    lipgloss.Style
	firstTabHidden  lipgloss.Style
	trailingTab     lipgloss.Style
	statusLine      lipgloss.Style
}

// NewStyles is a constructor for styles needed to render tabs and content.
func NewStyles() *Styles {
	base := lipgloss.NewStyle().
		// Height(10).
		// MaxHeight(12)
		Border(lipgloss.Border{
			Top:         " ",
			Bottom:      "─",
			Left:        "│",
			Right:       "│",
			TopLeft:     "│",
			TopRight:    "│",
			BottomLeft:  "└",
			BottomRight: "┘",
		}, true)

	return &Styles{
		baseContent: base,

		activeContent: lipgloss.NewStyle().
			Inherit(base).
			Padding(0, 1).
			BorderForeground(colorSelected),

		inactiveContent: lipgloss.NewStyle().
			Inherit(base).
			Padding(0, 1).
			BorderForeground(colorDimmed),

		activeTab: lipgloss.NewStyle().
			BorderForeground(colorSelected),

		inactiveTab: lipgloss.NewStyle().
			BorderForeground(colorDimmed),

		tabOpen: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.Border{
				Top:         "─",
				Bottom:      " ",
				Left:        "│",
				Right:       "│",
				TopLeft:     "╭",
				TopRight:    "╮",
				BottomLeft:  "┘",
				BottomRight: "└",
			}, true),

		tabHidden: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.Border{
				Top:         "─",
				Bottom:      "─",
				Left:        "│",
				Right:       "│",
				TopLeft:     "╭",
				TopRight:    "╮",
				BottomLeft:  "─",
				BottomRight: "─",
			}),

		firstTabOpen: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.Border{
				Top:         "─",
				Bottom:      " ",
				Left:        "│",
				Right:       "│",
				TopLeft:     "╭",
				TopRight:    "╮",
				BottomLeft:  "│",
				BottomRight: "└",
			}),

		firstTabHidden: lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.Border{
				Top:         "─",
				Bottom:      "─",
				Left:        "│",
				Right:       "│",
				TopLeft:     "╭",
				TopRight:    "╮",
				BottomLeft:  "╭",
				BottomRight: "─",
			}),

		trailingTab: lipgloss.NewStyle().
			Border(lipgloss.Border{
				Bottom:      "─",
				BottomLeft:  "─",
				BottomRight: "╮",
			}, false, true, true),

		statusLine: lipgloss.NewStyle().Bold(true),
	}
}
