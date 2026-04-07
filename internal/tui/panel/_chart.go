package panel

import (
	"log/slog"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/canvas/graph"
	"github.com/scmhub/ibsync"
)

type Chart struct {
	can canvas.Model
	cursor canvas.Point
}

func RenderBars(logger *slog.Logger, bars []ibsync.Bar, style *Styles) string {
	can := canvas.New(20, 11)
	c := Chart{can: can}

	c.can.Clear()
	graph.DrawXYAxis(&c.can, c.cursor, style.BlueFG)
	for i, b := range bars {
		x := i + 1
		l := b.Low
		bl := b.Open
		bh := b.Close
		h := b.High
		s := style.GreenFG
		if b.Open < b.Close {
			s = style.RedFG
			bl = b.Close
			bh = b.Open
		}
		graph.DrawCandlestickBottomToTop(&c.can, c.cursor.Add(canvas.Point{X: x, Y: -1}), l, bl, bh, h, s)
	}
	return c.can.View()
}

