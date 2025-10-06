// Package panels contains styling components to render strings for a TUI.
package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	colorSelected = lipgloss.Color("#7D56f4") // Purple
	colorDimmed   = lipgloss.Color("#666666") // Dark gray
	bordersWidth  = 2
)

var (
	baseContentStyle = lipgloss.NewStyle().
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
		// Height(10).
		// MaxHeight(12)

	activeContentStyle = lipgloss.NewStyle().
				Inherit(baseContentStyle).
				Padding(0, 1).
				BorderForeground(colorSelected)

	inactiveContentStyle = lipgloss.NewStyle().
				Inherit(baseContentStyle).
				Padding(0, 1).
				BorderForeground(colorDimmed)

	activeTabStyle = lipgloss.NewStyle().
			BorderForeground(colorSelected)

	inactiveTabStyle = lipgloss.NewStyle().
				BorderForeground(colorDimmed)

	tabOpenStyle = lipgloss.NewStyle().
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
		}, true)

	tabHiddenStyle = lipgloss.NewStyle().
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
		})

	tabFirstOpenStyle = lipgloss.NewStyle().
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
		})

	tabFirstHiddenStyle = lipgloss.NewStyle().
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
		})

	trailingTabStyle = lipgloss.NewStyle().
				Border(lipgloss.Border{
			Bottom:      "─",
			BottomLeft:  "─",
			BottomRight: "╮",
		}, false, true, true)

	statusStyle = lipgloss.NewStyle().Bold(true)
)

// Panel represents a a horizontal grouping of tabs.
type Panel struct {
	Index    int
	Tab      string
	Content  string
	Revealed bool
}

// RenderHorizontalGroup styles a panel grouping into horizontal tabs.
func RenderHorizontalGroup(panels []*Panel, selectedTab, width int) string {
	var content string
	var focused bool
	var style lipgloss.Style
	styles := make([]lipgloss.Style, 0)
	tabs := make([]string, 0)
	tabRow := make([]string, 0)
	tabsLength := 0

	for i, p := range panels {
		// Check if the tab is the first tab and whether it' open
		if i == 0 {
			if p.Revealed {
				style = tabFirstOpenStyle
			} else {
				style = tabFirstHiddenStyle
			}
		} else {
			if p.Revealed {
				style = tabOpenStyle
			} else {
				style = tabHiddenStyle
			}
		}
		// Add coloring to tab based on focus
		if selectedTab == p.Index {
			focused = true
			style = style.Inherit(activeTabStyle)
			// Restyle previous tabs accordingly, if necessary
			if i > 0 {
				for j := i; j > 0; j-- {
					styles[j-1] = styles[j-1].BorderBottomForeground(colorSelected)
				}
			}
		} else {
			style = style.Inherit(inactiveTabStyle)
			if focused {
				style = style.BorderBottomForeground(colorSelected)
			}
		}
		tabs = append(tabs, p.Tab)
		styles = append(styles, style)

		// Content styling
		if p.Revealed {
			if selectedTab == p.Index {
				content = activeContentStyle.Width(width - bordersWidth).Render(p.Content)
			} else {
				content = inactiveContentStyle.Width(width - bordersWidth).Render(p.Content)
			}
		}
	}

	for i, t := range tabs {
		tabRow = append(tabRow, styles[i].Render(t))
		tabsLength += lipgloss.Width(tabRow[i])
	}

	// Final trailing tab
	if focused {
		style = trailingTabStyle.Inherit(activeTabStyle)
	} else {
		style = trailingTabStyle.Inherit(inactiveTabStyle)
	}
	tabRow = append(tabRow,
		style.Render(strings.Repeat(" ", width-tabsLength-bordersWidth)),
	)

	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}

// RenderStatusLine styles the bottom status line of the TUI.
// :TODO Dynamically show allowed keypresses depending on context.
func RenderStatusLine(status string) string {
	return statusStyle.Render(status)
}
