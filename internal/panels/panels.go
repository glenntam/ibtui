// Package panels contains the components for various different panels displayed.
package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	colorSelected = lipgloss.Color("#7D56f4") // Purple
	colorDimmed   = lipgloss.Color("#666666") // Dark gray
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
		//Height(10).
		//MaxHeight(12)

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

	statusStyle = lipgloss.NewStyle().
		//Foreground(lipgloss.Color("#666666")).
		Italic(true)
)

// An individual tabbed panel.
type Panel struct {
	Index    int
	Tab      string
	Content  string
	Revealed bool
}

// Helper struct to style tabs.
type tab struct {
	tab   string
	style lipgloss.Style
}

// A general renderer to make all panel groups to look the same.
func RenderHorizontalGroup(panels []*Panel, selectedTab, width int) string {
	var focused bool
	var style lipgloss.Style
	var content string
	var tabRow []string
	var tabs []tab

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
			style = style.Copy().Inherit(activeTabStyle)
			// Restyle previous tabs accordingly, if necessary
			if i > 0 {
				for j := i; j > 0; j-- {
					tabs[j-1].style = tabs[j-1].style.Copy().BorderBottomForeground(colorSelected)
				}
			}
		} else {
			style = style.Copy().Inherit(inactiveTabStyle)
			if focused {
				style = style.Copy().BorderBottomForeground(colorSelected)
			}
		}
		tabs = append(tabs, tab{tab: p.Tab, style: style.Copy()})

		// Get the total string length of the tab element, including any border
		//tabsLength += lipgloss.Width(tabs[len(tabs)-1].tab)

		// Content styling
		if p.Revealed {
			if selectedTab == p.Index {
				content = activeContentStyle.Width(width - 2).Render(p.Content)
			} else {
				content = inactiveContentStyle.Width(width - 2).Render(p.Content)
			}
		}
	}

	for i, t := range tabs {
		tabRow = append(tabRow, t.style.Render(t.tab))
		tabsLength += lipgloss.Width(tabRow[i])
	}

	// Final trailing tab
	if focused {
		style = trailingTabStyle.Copy().Inherit(activeTabStyle)
	} else {
		style = trailingTabStyle.Copy().Inherit(inactiveTabStyle)
	}
	tabRow = append(tabRow, style.Render(strings.Repeat(" ", width-tabsLength-2)))

	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}
