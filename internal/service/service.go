// Package service is an abstraction layer in between ibsync/ibapi and the TUI.
// It handles context, so TUI keypresses can be fire-and-forget without any lag.
<<<<<<< Updated upstream
=======
//
// The TUI can read IBService struct fields directly without worrying about data
// corruption since ibsync internally already has thread-safe mutex.
>>>>>>> Stashed changes
package service

import (
	"context"
	"errors"
	// "fmt"
	"log/slog"
	"sync"
<<<<<<< Updated upstream
	"sync/atomic"
=======
	"strconv"
	// "strings"
	// "sync/atomic"
>>>>>>> Stashed changes
	"time"

	"github.com/glenntam/ibtui/internal/zerobridge"

	"github.com/rs/zerolog"
<<<<<<< Updated upstream
=======
	//"github.com/scmhub/ibapi"
	//"github.com/robaho/fixed"
>>>>>>> Stashed changes
	"github.com/scmhub/ibsync"
)

const (
	connTimeout    = 10 * time.Second
<<<<<<< Updated upstream
	sysTimeRefresh = 250 * time.Millisecond
=======
	sysTimeRefresh = 100 * time.Millisecond
>>>>>>> Stashed changes
	oneThousand    = 1_000
	oneMillion     = 1_000_000
)

<<<<<<< Updated upstream
var ErrContractRequired = errors.New("include at least one contract to start ibtui")
=======
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
>>>>>>> Stashed changes

// IBService continuously gets the IB account state and makes it suitable for TUI consumption.
type IBService struct {
	ib     *ibsync.IB
	logger *slog.Logger
	wg     sync.WaitGroup

	Contracts []*ibsync.Contract

<<<<<<< Updated upstream
	systemTime       atomic.Pointer[time.Time]
	cancelSystemTime context.CancelFunc

	tickers       atomic.Pointer[[]*ibsync.Ticker]
	cancelTickers context.CancelFunc

	bars          atomic.Pointer[[]ibsync.Bar]
=======
	SystemTime       *time.Time
	cancelSystemTime context.CancelFunc

	Tickers       []*ibsync.Ticker
	tickerIdx     map[int64]*ibsync.Ticker
	cancelTickers context.CancelFunc

	LastTime      string
	secondLastBar ibsync.Bar
	Bars          []ibsync.Bar
>>>>>>> Stashed changes
	barChan       chan ibsync.Bar
	cancelReqHist ibsync.CancelFunc
	cancelBars    context.CancelFunc

	ServiceStarted bool
}

// NewIBService instantiates a new IBService and connects to the IB client.
func NewIBService(host string, port, clientID int, contracts []*ibsync.Contract, logger *slog.Logger) (*IBService, error) {
	if len(contracts) < 1 {
		return nil, ErrContractRequired
	}
	// Init values
<<<<<<< Updated upstream
	ibs := &IBService{
		Contracts:        contracts,
		logger:           logger,
=======
	now := time.Now()
	ibs := &IBService{
		logger:           logger,
		Contracts:        contracts,
		SystemTime:       &now,
		LastTime:         string(now.Unix()),
>>>>>>> Stashed changes
		cancelSystemTime: func() {},
		cancelTickers:    func() {},
		cancelBars:       func() {},
		cancelReqHist:    func() {},
	}
<<<<<<< Updated upstream
	placeholder := time.Now()
	ibs.systemTime.Store(&placeholder)
=======
>>>>>>> Stashed changes
	ibs.stopTickers()
	ibs.stopBars()

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
<<<<<<< Updated upstream
=======
		//ibs.StartBars(ibs.Contracts[0])
>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
func (ibs *IBService) ReadTime() time.Time {
	return *ibs.systemTime.Load()
}

=======
// startSystemTime starts syncing the Service time to IB
>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
				systemTime := time.Unix(seconds, nanoseconds)
				ibs.systemTime.Store(&systemTime)
=======
				ibTime := time.Unix(seconds, nanoseconds)
				ibs.SystemTime = &ibTime
>>>>>>> Stashed changes
			}
		}
	}()
}

// Tickers
<<<<<<< Updated upstream
func (ibs *IBService) ReadTickers() []*ibsync.Ticker {
	return *ibs.tickers.Load()
}

=======
// startTickers clears ibs.Tickers slice and inserts ReqMktData for each Contract
>>>>>>> Stashed changes
func (ibs *IBService) startTickers() {
	ibs.stopTickers()
	var ctx context.Context
	ctx, ibs.cancelTickers = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
<<<<<<< Updated upstream
		tickers := make([]*ibsync.Ticker, 0, len(ibs.Contracts))
		for _, c := range ibs.Contracts {
			tickers = append(tickers, ibs.ib.ReqMktData(c, ""))
		}
		ibs.tickers.Store(&tickers)
=======
		for _, c := range ibs.Contracts {
			ibs.Tickers = append(ibs.Tickers, ibs.ib.ReqMktData(c, ""))
		}
>>>>>>> Stashed changes
		// wait for Cancel and then cancel tickers
		<-ctx.Done()
		for _, c := range ibs.Contracts {
			ibs.ib.CancelMktData(c)
		}
	}()
<<<<<<< Updated upstream
}

func (ibs *IBService) stopTickers() {
	ibs.cancelTickers()
	emptyTickers := make([]*ibsync.Ticker, 0)
	ibs.tickers.Store(&emptyTickers)
}

// Bars
func (ibs *IBService) ReadBars() []ibsync.Bar {
	// return *ibs.bars.Load()
	tmp := *ibs.bars.Load()
	tmpBar := make([]ibsync.Bar, 0)
	if len(tmp) != 0 {
		tmpBar = append(tmpBar, tmp[0])
	}
	return tmpBar
}
=======
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
>>>>>>> Stashed changes

func (ibs *IBService) StartBars(contract *ibsync.Contract) {
	ibs.stopBars()
	var ctx context.Context
	ctx, ibs.cancelBars = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		reqHistDone := make(chan struct{})
		go func() {
<<<<<<< Updated upstream
			ibs.barChan, ibs.cancelReqHist = ibs.ib.ReqHistoricalDataUpToDate(contract, "1 D", "1 hour", "TRADES", false, 1)
=======
			ibs.barChan, ibs.cancelReqHist = ibs.ib.ReqHistoricalDataUpToDate(contract, "60 S", "10 secs", "TRADES", false, 1)
>>>>>>> Stashed changes
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
				case bar, ok := <-ibs.barChan:
					if !ok {
						continue
					}
<<<<<<< Updated upstream
					currentBars := *ibs.bars.Load()
					newBars := make([]ibsync.Bar, len(currentBars))
					if len(currentBars) > 0 && currentBars[len(currentBars)-1].Date == bar.Date {
						copy(newBars, currentBars)
						newBars[len(newBars)-1] = bar
						ibs.bars.Store(&newBars)
					} else {
						newBars = append(currentBars, bar)
						ibs.bars.Store(&newBars)
=======
					if len(ibs.Bars) == 0 {
						ibs.Bars = append(ibs.Bars)
					} else if ibs.Bars[len(ibs.Bars)-1].Date != bar.Date {
						ibs.Bars = append(ibs.Bars)
						if len(ibs.Bars) >= 2 {
							ibs.Bars[len(ibs.Bars)-2] = ibs.secondLastBar
						}
					} else {
						ibs.Bars[len(ibs.Bars)-1] = bar
					}
					ibs.secondLastBar = bar

					// bars := make([]ibsync.Bar, len(ibs.Bars))
					// copy(bars, ibs.Bars)
					// if len(bars) > 0 && bars[len(bars)-1].Date == bar.Date {
					//     ibs.secondLastBar = bars[len(bars)-1]
					//     bars[len(bars)-1] = bar
					// } else {
					//     bars = append(bars, bar)
					//     if len(bars) > 2 {
					//         bars[len(bars)-2] = ibs.secondLastBar
					//     }
					// }
					// ibs.Bars = bars

				default:
					if len(ibs.Bars) > 0 {
						// barTime, err := ibsync.ParseIBTime(ibs.Bars[len(ibs.Bars)-1].Date)
						// if err != nil {
						//     ibs.logger.Error("couldn't parse bar date to time.time", "barDate", ibs.Bars[len(ibs.Bars)-1].Date)
						//     break
						// }

						// tick := ibs.tickerIdx[contract.ConID]
						// tickTime := tick.Time()

						// nextBarTime := ibs.Bars[len(ibs.Bars)-1].Date

						for _, t := range ibs.Tickers {
							if t.Contract().ConID == contract.ConID {

								ibs.Bars[len(ibs.Bars)-1].Close = t.Last()
								vol := ibs.Bars[len(ibs.Bars)-1].Volume.Int()
								incvol := int64(0)
								if ibs.LastTime != t.LastTimestamp() &&
								t.Last() != t.PrevLast() &&
								t.LastSize() != t.PrevLastSize() &&
								t.Bid() != t.PrevBid() &&
								t.BidSize() != t.PrevBidSize() &&
								t.Ask() != t.PrevAsk() &&
								t.AskSize() != t.PrevAskSize() {
									incvol = t.LastSize().Int()
									ibs.LastTime = t.LastTimestamp()
								}
								ibs.logger.Debug("v", "vol", vol, "incvol", incvol, "lasttimestamp", t.LastTimestamp())

								v := vol + incvol
								vStr := strconv.FormatInt(v, 10)
								vDec := ibsync.StringToDecimal(vStr)
								ibs.Bars[len(ibs.Bars)-1].Volume = vDec

							}
						}
>>>>>>> Stashed changes
					}
				}
			}
		}
	}()
}

func (ibs *IBService) stopBars() {
	ibs.cancelReqHist()
	ibs.cancelBars()
<<<<<<< Updated upstream
	emptyBars := make([]ibsync.Bar, 0)
	ibs.bars.Store(&emptyBars)
=======
	ibs.Bars = make([]ibsync.Bar, 0)
>>>>>>> Stashed changes
}
