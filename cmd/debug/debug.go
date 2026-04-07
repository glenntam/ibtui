// ntcharts - Copyright (c) 2024 Neomantra Corp.

package main

import (
	"fmt"
	"os"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var candleStyle1 = lipgloss.NewStyle().
	Foreground(lipgloss.Color("4")) // blue

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow


type model struct {}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// draw candlesticks with bodies and wicks

	return m, nil
}
func draw() string {
	c1 := canvas.New(20, 11)
	cursor := canvas.Point{0, 11 - 1}
	c1.Clear()
	graph.DrawXYAxis(&c1, cursor, axisStyle)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 1, Y: -1}), .3, 1.3, 2., 6.3, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 2, Y: -1}), .3, 1.6, 4, 6.3, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 3, Y: -1}), .6, 1.3, 4, 6.3, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 4, Y: -1}), .6, 1.6, 4, 6.3, candleStyle1)

	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 6, Y: -1}), 1.6, 1.6, 4, 4, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 7, Y: -1}), 1.6, 2.6, 4, 4, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 8, Y: -1}), 1.6, 2.3, 4, 5, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 9, Y: -1}), 1.3, 4, 4.3, 5, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: 10, Y: -1}), 1.6, 4, 4.6, 5, candleStyle1)
	return c1.View()
}

func (m model) View() string {
	s := "q to quit\n"
	s += lipgloss.JoinHorizontal(lipgloss.Top, draw()) + "\n"
	return s
}

func main() {
	m := model{}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
