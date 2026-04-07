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
	// "strings"
	"time"

	"github.com/glenntam/ibtui/internal/zerobridge"

	"github.com/rs/zerolog"
	//"github.com/scmhub/ibapi"
	//"github.com/robaho/fixed"
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

	Bars          []ibsync.Bar
	BarsMin       float64
	BarsMax       float64
	BarsMutex     sync.RWMutex
	barChan       chan ibsync.Bar
	cancelReqHist ibsync.CancelFunc
	cancelBars    context.CancelFunc
	LastBarTime   string
	secondLastBar ibsync.Bar

	ServiceStarted bool
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
		LastBarTime:         string(now.Unix()),
		cancelSystemTime: func() {},
		cancelTickers:    func() {},
		cancelBars:       func() {},
		cancelReqHist:    func() {},
	}
	//ibs.Bars([]ibsync.Bar{})
	//ibs.stopTickers()
	//ibs.stopBars()

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
// It should be run after the TUI is drawn, since StartIBService takes time to populate.
func (ibs *IBService) StartIBService() {
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		err := ibs.ib.QualifyContract(ibs.Contracts...)
		if err != nil {
			ibs.logger.Error("couldn't qualify contracts", "err", err)
			return
		}
		ibs.startSystemTime(sysTimeRefresh)
		ibs.startTickers()
		ibs.StartBars(ibs.Contracts[0])
		ibs.ServiceStarted = true
	}()
}

// StopIBService gracefully disconnects the IB Client and cancels existing contexts.
func (ibs *IBService) StopIBService() {
	ibs.cancelSystemTime()
	ibs.stopTickers()
	ibs.stopBars()
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
// func (ibs *IBService) StartBars2(contract *ibsync.Contract) {
//     ibs.stopBars2()
//     var ctx context.Context
//     ctx, ibs.cancelBars = context.WithCancel(context.Background())
//     ibs.wg.Add(1)
//     go func() {
//         defer ibs.wg.Done()
//         for {
//             select {
//             case <-ctx.Done():
//                 return
//             default:
//             }
//             ibs.barChan, _ = ibs.ib.ReqHistoricalData(contract, "", "1 D", "1 hour", "TRADES", false, 1)
//             select {
//             case <-ctx.Done():
//                 return
//             default:
//             }
//             var bars []ibsync.Bar
//             for b := range ibs.barChan {
//                 bars = append(bars, b)
//             }
//             ibs.Bars = bars

//         }
//     }()
// }

// func (ibs *IBService) stopBars2() {
//     ibs.cancelBars()
//     ibs.Bars = make([]ibsync.Bar, 0)
// }

func (ibs *IBService) StartBars(contract *ibsync.Contract) {
	ibs.stopBars()
	var ctx context.Context
	ctx, ibs.cancelBars = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	ibs.logger.Debug("startbars")
	go func() {
		defer ibs.wg.Done()
		reqHistDone := make(chan struct{})
		go func() {
			ibs.barChan, ibs.cancelReqHist = ibs.ib.ReqHistoricalDataUpToDate(contract, "60 S", "10 secs", "TRADES", false, 1)
			ibs.logger.Debug("req up to date started")
			close(reqHistDone)
		}()

		select {
		case <-ctx.Done():
			ibs.stopBars()
			return
		case <-reqHistDone:
			for {
				select {
				case <-ctx.Done():
					ibs.stopBars()
					return
				case bar, ok := <-ibs.barChan: // Receive new bars; but IB only sends every 5 seconds (IB pacing limit).
					if !ok {
						continue
					}
					if ibs.BarsMax < bar.High {
						ibs.BarsMax = bar.High
					}
					if ibs.BarsMin == 0.0 || ibs.BarsMin > bar.Low {
						ibs.BarsMin = bar.Low
					}

					ibs.BarsMutex.Lock()
					if len(ibs.Bars) == 0 {
						ibs.Bars = append(ibs.Bars, bar)
					} else if ibs.Bars[len(ibs.Bars)-1].Date == bar.Date {
						ibs.Bars[len(ibs.Bars)-1] = bar
					} else {
						ibs.Bars = append(ibs.Bars, bar)
						ibs.Bars[len(ibs.Bars)-2] = ibs.secondLastBar
					}
					ibs.BarsMutex.Unlock()
					ibs.secondLastBar = bar

				default: // If no new bars in the meantime, continually update last bar with current ticker data
					for _, t := range ibs.Tickers {
						if len(ibs.Bars) > 0 && t.Contract().ConID == contract.ConID {
							// NOTE: Try to continually increment the latest bar volume with each individual tick.
							// It's not perfect, but quite close.
							prevVol := ibs.Bars[len(ibs.Bars)-1].Volume.Int()
							incrementVol := int64(0)

							if ibs.LastBarTime != t.LastTimestamp() &&
							t.Last() != t.PrevLast() &&
							t.LastSize() != t.PrevLastSize() &&
							t.Bid() != t.PrevBid() &&
							t.BidSize() != t.PrevBidSize() &&
							t.Ask() != t.PrevAsk() &&
							t.AskSize() != t.PrevAskSize() {
								incrementVol = t.LastSize().Int()
								ibs.LastBarTime = t.LastTimestamp()
							}
							vol64 := prevVol + incrementVol
							volDec := ibsync.StringToDecimal(strconv.FormatInt(vol64, 10))

							ibs.BarsMutex.Lock()
							ibs.Bars[len(ibs.Bars)-1].Volume = volDec
							ibs.Bars[len(ibs.Bars)-1].Close = t.Last()
							if t.Last() > ibs.Bars[len(ibs.Bars)-1].High {
								ibs.Bars[len(ibs.Bars)-1].High = t.Last()
							}
							if t.Last() < ibs.Bars[len(ibs.Bars)-1].Low {
								ibs.Bars[len(ibs.Bars)-1].Low = t.Last()
							}
							ibs.BarsMutex.Unlock()
							if ibs.BarsMax < t.Last() {
								ibs.BarsMax = t.Last()
							}
							if ibs.BarsMin == 0.0 || ibs.BarsMin > t.Last() {
								ibs.BarsMin = t.Last()
							}
						}
					}
				}
			}
		}
	}()
}

func (ibs *IBService) stopBars() {
	ibs.cancelReqHist()
	ibs.cancelBars()
	ibs.BarsMutex.Lock()
	ibs.Bars = make([]ibsync.Bar, 0)
	ibs.BarsMutex.Unlock()
}
