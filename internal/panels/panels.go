// Package panels contains styling components to render strings for the TUI.
package panels

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Panel represents a horizontal grouping of tabs.
type Panel struct {
	Index    int
	Tab      string
	Content  string
	Revealed bool
}

// RenderHorizontalGroup styles a panel grouping into horizontal tabs.
func RenderHorizontalGroup(panels []*Panel, styles *Styles, selectedTab, width int) string {
	var content string
	var focused bool
	var style lipgloss.Style
	tabStyles := make([]lipgloss.Style, 0)
	tabs := make([]string, 0)
	tabRow := make([]string, 0)
	tabsLength := 0

	for i, p := range panels {
		// Check if the tab is the first tab and whether it' open
		switch {
		case i == 0 && p.Revealed:
			style = styles.firstTabOpen
		case i == 0 && !p.Revealed:
			style = styles.firstTabHidden
		case p.Revealed:
			style = styles.tabOpen
		default:
			style = styles.tabHidden
		}

		// Add coloring to tab based on focus
		if selectedTab == p.Index {
			focused = true
			style = style.Inherit(styles.activeTab)
			// Restyle previous tabs accordingly, if necessary
			if i > 0 {
				for j := i; j > 0; j-- {
					tabStyles[j-1] = tabStyles[j-1].BorderBottomForeground(colorSelected)
				}
			}
		} else {
			style = style.Inherit(styles.inactiveTab)
			if focused {
				style = style.BorderBottomForeground(colorSelected)
			}
		}
		tabs = append(tabs, p.Tab)
		tabStyles = append(tabStyles, style)

		// Content styling
		if p.Revealed {
			if selectedTab == p.Index {
				content = styles.activeContent.Width(width - bordersWidth).Render(p.Content)
			} else {
				content = styles.inactiveContent.Width(width - bordersWidth).Render(p.Content)
			}
		}
	}

	for i, t := range tabs {
		tabRow = append(tabRow, tabStyles[i].Render(t))
		tabsLength += lipgloss.Width(tabRow[i])
	}

	// Final trailing tab
	if focused {
		style = styles.trailingTab.Inherit(styles.activeTab)
	} else {
		style = styles.trailingTab.Inherit(styles.inactiveTab)
	}
	tabRow = append(tabRow,
		style.Render(strings.Repeat(" ", width-tabsLength-bordersWidth)),
	)

	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}

// RenderStatusLine styles the bottom status line of the TUI.
// :TODO Dynamically show allowed keypresses depending on context.
func RenderStatusLine(status string, styles *Styles) string {
	return styles.statusLine.Render(status)
}
