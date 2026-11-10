// ntcharts - Copyright (c) 2024 Neomantra Corp.

// Package main exercises the ntcharts candlestick renderer.
package main

import (
	"fmt"
	"os"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	chartWidth  = 20
	chartHeight = 11

	x1  = 1
	x2  = 2
	x3  = 3
	x4  = 4
	x6  = 6
	x7  = 7
	x8  = 8
	x9  = 9
	x10 = 10

	level03 = 0.3
	level06 = 0.6
	level13 = 1.3
	level16 = 1.6
	level20 = 2.0
	level23 = 2.3
	level26 = 2.6
	level40 = 4.0
	level43 = 4.3
	level46 = 4.6
	level50 = 5.0
	level63 = 6.3
)

type model struct{}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { //nolint:ireturn // bubbletea requires this signature
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	// draw candlesticks with bodies and wicks

	return m, nil
}

func draw() string {
	candleStyle1 := lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")) // blue
	axisStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("3")) // yellow

	c1 := canvas.New(chartWidth, chartHeight)
	cursor := canvas.Point{X: 0, Y: chartHeight - 1}
	c1.Clear()
	graph.DrawXYAxis(&c1, cursor, axisStyle)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x1, Y: -1}),
		level03, level13, level20, level63, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x2, Y: -1}),
		level03, level16, level40, level63, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x3, Y: -1}),
		level06, level13, level40, level63, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x4, Y: -1}),
		level06, level16, level40, level63, candleStyle1)

	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x6, Y: -1}),
		level16, level16, level40, level40, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x7, Y: -1}),
		level16, level26, level40, level40, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x8, Y: -1}),
		level16, level23, level40, level50, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x9, Y: -1}),
		level13, level40, level43, level50, candleStyle1)
	graph.DrawCandlestickBottomToTop(&c1, cursor.Add(canvas.Point{X: x10, Y: -1}),
		level16, level40, level46, level50, candleStyle1)
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
