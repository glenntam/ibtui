// package panels contains the components for various different panels displayed.
package panels

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	tabContentStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).              // Full border around content
		BorderForeground(lipgloss.Color("#7D56F4")).   // Purple border (matches active tab)
		Padding(0, 1)                                 // 1 line vertical, 2 spaces horizontal padding
		//MinHeight(15)                                  // Minimum height for content area
	activeTabBorderStyle = lipgloss.NewStyle().
		Bold(true).                                        // Make text bold
		Foreground(lipgloss.Color("#FFFFFF")).            // White text
		Background(lipgloss.Color("#7D56F4")).            // Purple background
		Border(lipgloss.RoundedBorder(), true, true, false, true). // Border: top=yes, right=yes, bottom=NO, left=yes
		BorderForeground(lipgloss.Color("#7D56F4")).       // Purple border color
		Padding(0, 1)                                     // No vertical padding, 1 space horizontal

	inactiveTabBorderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).             // Gray text
		Background(lipgloss.Color("#1A1A1A")).             // Dark background
		Border(lipgloss.RoundedBorder(), true, true, false, true). // Same border pattern
		BorderForeground(lipgloss.Color("#333333")).       // Dark gray border
		Padding(0, 1)                                     // Same padding

	statusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).         // Gray color for status text
		Italic(true)                                   // Make it italic
)

type PanelGroup struct {
	Tabs      []string
	Content   []string
	ActiveTab int
}

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
		Width(width - 2).      // Account for border
		Render(g.Content[g.ActiveTab]) // Show content for current tab
	return lipgloss.JoinVertical(lipgloss.Left, tabBar, body)
}

