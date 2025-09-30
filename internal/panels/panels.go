// Package panels contains the components for various different panels displayed.
package panels

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	activeContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1).     // No vertical padding, 1 space horizontal
				MaxHeight(12)
	inactiveContentStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#333333")).
				Padding(0, 1) // No vertical padding, 1 space horizontal
	activeTabStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).                     // White
				Background(lipgloss.Color("#7D56F4")).                     // Purple
				Border(lipgloss.RoundedBorder(), true, true, false, true). // top, right, bottom, left
				BorderForeground(lipgloss.Color("#7D56F4")).
				Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")). // Grey
				Background(lipgloss.Color("#1A1A1A")). // Dark background
				Border(lipgloss.RoundedBorder(), true, true, false, true).
				BorderForeground(lipgloss.Color("#333333")). // Dark gray border
				Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Italic(true)
)

// An individual tabbed panel.
type Panel struct {
	Index    int
	Tab      string
	Content  string
	Revealed bool
}

// A general renderer to make all panels look the same.
func RenderHorizontalGroup(panels []*Panel, selectedTab, width int) string {
	var tabRow []string
	content := ""
	for _, p := range panels {
		if selectedTab == p.Index {
			tabRow = append(tabRow, activeTabStyle.Render(p.Tab))
		} else {
			tabRow = append(tabRow, inactiveTabStyle.Render(p.Tab))
		}
		if p.Revealed == true {
			if selectedTab == p.Index {
				content = activeContentStyle.Width(width - 2).Render(p.Content)
			} else {
				content = inactiveContentStyle.Width(width - 2).Render(p.Content)
			}
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, tabRow...)
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)
}
