// Package panels contains the components for various different panels displayed.
package panels

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	tabContentStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1) // No vertical padding, 1 space horizontal
	activeTabBorderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).                     // White
				Background(lipgloss.Color("#7D56F4")).                     // Purple
				Border(lipgloss.RoundedBorder(), true, true, false, true). // top, right, bottom, left
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	inactiveTabBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")). // Grey
				Background(lipgloss.Color("#1A1A1A")). // Dark background
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(lipgloss.Color("#333333")). // Dark gray border
				Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
)

// A panel group holds the tab and corresponding content string.
type PanelGroup struct {
	Tabs      []string
	Content   []string
	ActiveTab int
}

// A general renderer to make all panels look the same.
func RenderPanelGroup(g *PanelGroup, width int) string {
	var tabRow []string
	for i, tab := range g.Tabs {
		if i == g.ActiveTab {
			tabRow = append(tabRow, activeTabBorderStyle.Render(tab))
		} else {
			tabRow = append(tabRow, inactiveTabBorderStyle.Render(tab))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	body := tabContentStyle.
		MaxHeight(12).
		Width(width - 2).              // Account for border
		Render(g.Content[g.ActiveTab]) // Show content for current tab
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, body)
}
