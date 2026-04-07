package panel

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	colorSelected = lipgloss.Color("5") // Purple
	colorDimmed   = lipgloss.Color("8") // Dark gray
)

// Styles contains styling to render tabs and content.
type Styles struct {
	RedFG           lipgloss.Style
	RedBG           lipgloss.Style
	RedBrightFG     lipgloss.Style
	RedBrightBG     lipgloss.Style
	GreenFG         lipgloss.Style
	GreenBG         lipgloss.Style
	BlueFG          lipgloss.Style
	BlueBG          lipgloss.Style
	YellowFG        lipgloss.Style
	YellowBG        lipgloss.Style
	StatusBar       lipgloss.Style
	BaseContent     lipgloss.Style
	ActiveContent   lipgloss.Style
	InactiveContent lipgloss.Style
	ActiveTab       lipgloss.Style
	InactiveTab     lipgloss.Style
	TabOpen         lipgloss.Style
	TabHidden       lipgloss.Style
	FirstTabOpen    lipgloss.Style
	FirstTabHidden  lipgloss.Style
	TrailingTab     lipgloss.Style
	HelpBar         lipgloss.Style
}

// Default creates a struct of styles used by the TUI
func DefaultTheme() *Styles {
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
		RedFG: lipgloss.NewStyle().Foreground(lipgloss.Color("52")),

		RedBG: lipgloss.NewStyle().Background(lipgloss.Color("52")),

		RedBrightFG: lipgloss.NewStyle().Foreground(lipgloss.Color("1")),

		RedBrightBG: lipgloss.NewStyle().Background(lipgloss.Color("1")),

		GreenFG: lipgloss.NewStyle().Foreground(lipgloss.Color("22")),

		GreenBG: lipgloss.NewStyle().Background(lipgloss.Color("22")),

		BlueFG: lipgloss.NewStyle().Foreground(lipgloss.Color("12")),

		BlueBG: lipgloss.NewStyle().Background(lipgloss.Color("12")),

		YellowFG: lipgloss.NewStyle().Foreground(lipgloss.Color("3")),

		YellowBG: lipgloss.NewStyle().Background(lipgloss.Color("3")),

		StatusBar: lipgloss.NewStyle().Bold(true),

		BaseContent: base,

		ActiveContent: lipgloss.NewStyle().
			Inherit(base).
			Padding(0, 1).
			BorderForeground(colorSelected),

		InactiveContent: lipgloss.NewStyle().
			Inherit(base).
			Padding(0, 1).
			BorderForeground(colorDimmed),

		ActiveTab: lipgloss.NewStyle().
			BorderForeground(colorSelected),

		InactiveTab: lipgloss.NewStyle().
			BorderForeground(colorDimmed),

		TabOpen: lipgloss.NewStyle().
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

		TabHidden: lipgloss.NewStyle().
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

		FirstTabOpen: lipgloss.NewStyle().
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

		FirstTabHidden: lipgloss.NewStyle().
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

		TrailingTab: lipgloss.NewStyle().
			Border(lipgloss.Border{
				Bottom:      "─",
				BottomLeft:  "─",
				BottomRight: "╮",
			}, false, true, true),

		HelpBar: lipgloss.NewStyle().Bold(true),
	}
}
