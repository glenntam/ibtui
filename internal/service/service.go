// Package service is an abstraction layer in between ibsync/ibapi and the TUI.
// It handles context, so TUI keypresses can be fire-and-forget without any lag.
package service

import (
	"context"
	"errors"
	// "fmt"
	"log/slog"
	"sync"
	"strconv"
	"time"

	"github.com/glenntam/ibtui/internal/zerobridge"

	"github.com/rs/zerolog"
	"github.com/scmhub/ibsync"
)

const (
	connTimeout    = 10 * time.Second
	sysTimeRefresh = 100 * time.Millisecond
	oneThousand    = 1_000
	oneMillion     = 1_000_000
)

var (
	ErrContractRequired = errors.New("include at least one contract to start ibtui")
	ErrInvalidBarDate = errors.New("couldn't convert ibsync.Bar.Date() to time.Time")
)

var barSize = map[string]time.Duration{
	"1 secs":  time.Second,
	"5 secs":  5 * time.Second,
	"10 secs": 10 * time.Second,
	"15 secs": 15 * time.Second,
	"30 secs": 30 * time.Second,
	"1 mins":  time.Minute,
	"2 mins":  2 * time.Minute,
	"3 mins":  3 * time.Minute,
	"5 mins":  5 * time.Minute,
	"10 mins": 10 * time.Minute,
	"15 mins": 15 * time.Minute,
	"20 mins": 20 * time.Minute,
	"30 mins": 30 * time.Minute,
	"1 hours": time.Hour,
	"2 hours": 2 * time.Hour,
	"3 hours": 3 * time.Hour,
	"4 hours": 4 * time.Hour,
	"8 hours": 8 * time.Hour,
	"1 day":   24 * time.Hour,
	"1 week":  7 * 24 * time.Hour,
}

type Chart struct {
	Bars           []ibsync.Bar
	Mutex          sync.RWMutex
	PriceMin       float64
	PriceMax       float64
	barChan        chan ibsync.Bar
	cancelReqHist  ibsync.CancelFunc
	cancelBars     context.CancelFunc
	lastTickerTime string
	secondLastBar  ibsync.Bar
}

// IBService continuously gets the IB account state and makes it suitable for TUI consumption.
type IBService struct {
	ib     *ibsync.IB
	logger *slog.Logger
	wg     sync.WaitGroup

	Contracts []*ibsync.Contract

	SystemTime       *time.Time
	cancelSystemTime context.CancelFunc

	Tickers       []*ibsync.Ticker
	tickerIdx     map[int64]*ibsync.Ticker
	cancelTickers context.CancelFunc

	ChartL *Chart
	ChartR *Chart

	ServiceStarted bool
}

func NewChart(contract *ibsync.Contract, ticker *ibsync.Ticker, wg *sync.WaitGroup, ib *ibsync.IB) *Chart {
	c := &Chart{
		cancelReqHist:  func() {},
		cancelBars:     func() {},
		lastTickerTime: string(time.Now().Unix()),
	}
	c.Bars = make([]ibsync.Bar, 0)
	c.StartBars(contract, ticker, wg, ib)
	return c
}

// NewIBService instantiates a new IBService and connects to the IB client.
func NewIBService(host string, port, clientID int, contracts []*ibsync.Contract, logger *slog.Logger) (*IBService, error) {
	if len(contracts) < 1 {
		return nil, ErrContractRequired
	}
	// Init values
	now := time.Now()
	ibs := &IBService{
		logger:           logger,
		Contracts:        contracts,
		SystemTime:       &now,
		cancelSystemTime: func() {},
		cancelTickers:    func() {},
	}

	// Connect to IB
	ibs.ib = ibsync.NewIB()
	ibCfg := ibsync.NewConfig(
		ibsync.WithHost(host),
		ibsync.WithPort(port),
		ibsync.WithClientID(int64(clientID)),
		ibsync.WithTimeout(connTimeout),
	)
	err := ibs.ib.Connect(ibCfg)
	if err != nil {
		return nil, err
	}
	// ibsync internal logger
	bridge := &zerobridge.ZerologToSlogBridge{Slogger: ibs.logger}
	zeroLogger := zerolog.New(bridge).With().Timestamp().Logger()
	ibs.ib.SetLogger(zeroLogger)
	ibs.ib.SetClientLogLevel(1)

	return ibs, nil
}

// StartIBService begins grabbing data from the IB client.
// It should be run after the TUI is rendered, since StartIBService takes time to populate.
func (ibs *IBService) StartIBService() {
	ibs.wg.Add(1)
	//go func() {
	//	defer ibs.wg.Done()
	err := ibs.ib.QualifyContract(ibs.Contracts...)
	if err != nil {
		ibs.logger.Error("couldn't qualify contracts", "err", err)
		return
	}
	ibs.startSystemTime(sysTimeRefresh)
	ibs.startTickers()

		// ibs.StartBars(ibs.Contracts[0])
	//}()
	ibs.ChartL = NewChart(ibs.Contracts[0], ibs.Tickers[0], &ibs.wg, ibs.ib)

	ibs.ServiceStarted = true
}

// StopIBService gracefully disconnects the IB Client and cancels existing contexts.
func (ibs *IBService) StopIBService() {
	ibs.cancelSystemTime()
	ibs.stopTickers()
	ibs.ChartL.stopBars()
	ibs.wg.Wait()
	if ibs.ib == nil {
		return
	}
	err := ibs.ib.Disconnect()
	if err != nil {
		ibs.logger.Error("couldn't disconnect from TWS/GW", "error", err)
	}
	ibs.logger.Info("gracefully disconnected from TWS/GW")
}

// System Time
// startSystemTime starts syncing the Service time to IB
func (ibs *IBService) startSystemTime(interval time.Duration) {
	var ctx context.Context
	ctx, ibs.cancelSystemTime = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				i64, _ := ibs.ib.ReqCurrentTimeInMillis()
				seconds := i64 / oneThousand
				nanoseconds := (i64 % oneThousand) * oneMillion
				ibTime := time.Unix(seconds, nanoseconds)
				ibs.SystemTime = &ibTime
			}
		}
	}()
}

// Tickers
// startTickers clears ibs.Tickers slice and inserts ReqMktData for each Contract
func (ibs *IBService) startTickers() {
	ibs.stopTickers()
	var ctx context.Context
	ctx, ibs.cancelTickers = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		for _, c := range ibs.Contracts {
			ibs.Tickers = append(ibs.Tickers, ibs.ib.ReqMktData(c, ""))
		}
		// wait for Cancel and then cancel tickers
		<-ctx.Done()
		for _, c := range ibs.Contracts {
			ibs.ib.CancelMktData(c)
		}
	}()
	// map the tickers' ContractID for easy reference later.
	ibs.tickerIdx = make(map[int64]*ibsync.Ticker, len(ibs.Tickers))
	for _, t := range ibs.Tickers {
		conID := t.Contract().ConID
		ibs.tickerIdx[conID] = t
	}
}

// stopTickers cancels any existing Tickers and makes a blank slick of Tickers
func (ibs *IBService) stopTickers() {
	ibs.cancelTickers()
	ibs.Tickers = make([]*ibsync.Ticker, 0, len(ibs.Contracts))
}

// Bars
func (c *Chart) StartBars(con *ibsync.Contract, t *ibsync.Ticker, wg *sync.WaitGroup, ib *ibsync.IB) {
	c.stopBars()
	var ctx context.Context
	ctx, c.cancelBars = context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		reqHistDone := make(chan struct{})
		go func() {
			c.barChan, c.cancelReqHist = ib.ReqHistoricalDataUpToDate(con, "60 S", "5 secs", "TRADES", false, 1)
			close(reqHistDone) // Keep this thread running until ibs.cancelReqHist is called elsewhere.
		}()

		select {
		case <-ctx.Done():
			c.stopBars()
			return
		case <-reqHistDone:  // While reqHistDone is unclosed and the above ReqHistoricalDataUpToDate is still running.
			for {
				select {
				case <-ctx.Done():
					c.stopBars()
					return
				case bar, ok := <-c.barChan: // Receive new bars; but IB only sends every 5 seconds (IB pacing limit).
					if !ok {
						continue  // No new bar updates. Start the for-select loop again and try current ticker data.
					}
					if c.PriceMax < bar.High {
						c.PriceMax = bar.High
					}
					if c.PriceMin == 0.0 || c.PriceMin > bar.Low {
						c.PriceMin = bar.Low
					}

					c.Mutex.Lock()
					if len(c.Bars) == 0 {
						c.Bars = append(c.Bars, bar)
					} else if c.Bars[len(c.Bars)-1].Date == bar.Date {
						c.Bars[len(c.Bars)-1] = bar
					} else {
						c.Bars = append(c.Bars, bar)
						c.Bars[len(c.Bars)-2] = c.secondLastBar
					}
					c.Mutex.Unlock()
					c.secondLastBar = bar

				default: // If no new bars in the meantime, continually update last bar with current ticker data
					if len(c.Bars) > 0 && t.Contract().ConID == con.ConID {
						// NOTE: Try to continually increment the latest bar volume with each individual tick.
						// It's not perfect, but quite close.
						lastBar := c.Bars[len(c.Bars)-1]
						prevVol := lastBar.Volume.Int()
						incrementVol := int64(0)

						if c.lastTickerTime != t.LastTimestamp() &&
						t.Last() != t.PrevLast() &&
						t.LastSize() != t.PrevLastSize() &&
						t.Bid() != t.PrevBid() &&
						t.BidSize() != t.PrevBidSize() &&
						t.Ask() != t.PrevAsk() &&
						t.AskSize() != t.PrevAskSize() {
							incrementVol = t.LastSize().Int()
							c.lastTickerTime = t.LastTimestamp()
						}
						vol64 := prevVol + incrementVol
						volDec := ibsync.StringToDecimal(strconv.FormatInt(vol64, 10))

						c.Mutex.Lock()
						lastBar.Volume = volDec
						lastBar.Close = t.Last()
						if t.Last() > lastBar.High {
							lastBar.High = t.Last()
						}
						if t.Last() < lastBar.Low {
							lastBar.Low = t.Last()
						}
						c.Mutex.Unlock()
						if c.PriceMax < t.Last() {
							c.PriceMax = t.Last()
						}
						if c.PriceMin == 0.0 || c.PriceMin > t.Last() {
							c.PriceMin = t.Last()
						}
					}
				}
			}
		}
	}()
}

func (c *Chart) stopBars() {
	c.cancelReqHist()
	c.cancelBars()
	c.Mutex.Lock()
	c.Bars = make([]ibsync.Bar, 0)
	c.Mutex.Unlock()
}
