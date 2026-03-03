// Package state polls IB Gateway/TWS to receive current state of account.
package state

import (
	"fmt"
	//"log/slog"
	"time"

	"github.com/scmhub/ibsync"
)

const (
	oneThousand = 1_000
	oneMillion  = 1_000_000
)

// Chart contains the parameters and output of bars for the selected contract.
type Chart struct {
	Contract   *ibsync.Contract
	Duration   string
	BarSize    string
	whatToShow string
	useRTH     bool
	formatDate int
	Bars       []ibsync.Bar
	Cancel     ibsync.CancelFunc
}

// IBState constains the results of polling the IB account state.
type IBState struct {
	Chart            *Chart
	CurrentTime      time.Time
	Contracts        []*ibsync.Contract
	Tickers          []*ibsync.Ticker
}

// NewIBState makes a new IBSState container.
func NewIBState() *IBState {
	return &IBState{
		CurrentTime: time.Now(),
	}
}

// ReqCurrentTimeMilli retrieves IB account system time in time.Time format.
func (s *IBState) ReqCurrentTimeMilli(ib *ibsync.IB) error {
	m, err := ib.ReqCurrentTimeInMillis()
	if err != nil {
		s.CurrentTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		return fmt.Errorf("couldn't request IB time, using system time instead: %w", err)
	}
	seconds := m / oneThousand
	nanoseconds := (m % oneThousand) * oneMillion
	s.CurrentTime = time.Unix(seconds, nanoseconds)
	return nil
}

// GetBars populates s.Chart's Bars
func (s *IBState) GetBars(ib *ibsync.IB) {
	if s.Chart == nil && s.Contracts != nil {
		time.Sleep(2 * time.Second)
		s.Chart = &Chart{}
		s.Chart.Contract = s.Contracts[0]
		s.Chart.BarSize = "1 hour"
		s.Chart.Duration = "1 D"
		s.Chart.whatToShow = "TRADES"
		s.Chart.useRTH = false
		s.Chart.formatDate = 1
		//slog.Debug(fmt.Sprintf("%v", s.Contracts[0]))
	}
	// if s.Chart.Cancel != nil {
	//     s.Chart.Cancel()
	//     time.Sleep(1 * time.Second)
	// }
	//barChan, cancel := ib.ReqHistoricalDataUpToDate(
	barChan, cancel := ib.ReqHistoricalDataUpToDate(
		s.Chart.Contract,
		s.Chart.Duration,
		s.Chart.BarSize,
		s.Chart.whatToShow,
		s.Chart.useRTH,
		s.Chart.formatDate,
	)
	s.Chart.Cancel = cancel
	go func() {
		for bar := range barChan {
			if len(s.Chart.Bars) == 0 {
				s.Chart.Bars = append(s.Chart.Bars, bar)
			else

			> 0 && bar.Date == s.Chart.Bars[len(s.Chart.Bars)-1].Date {
				s.Chart.Bars[len(s.Chart.Bars)-1] = bar
			} else {
				s.Chart.Bars = append(s.Chart.Bars, bar)
			}
		}
	}()
}
// func (s *IBState) GetTickerPrice(ib *ibsync.IB) error {
//     fut := ibsync.NewFuture("MNQ", "202603", "CME", "2", "USD")
//     err := ib.QualifyContract(fut)
//     if err != nil {
//         panic(err)
//     }
//     futTicker := ib.ReqMktData(fut, "")
//     s.Last = string(futTicker.Last)

//     time.Sleep(5 * time.Second)
//     ib.CancelMktData(eurusd)
// }
