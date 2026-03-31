// Package service is an abstraction layer in between ibsync/ibapi and the TUI.
// It handles context, so TUI keypresses can be fire-and-forget without any lag.
package service

import (
	"context"
	"errors"
	// "fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glenntam/ibtui/internal/zerobridge"

	"github.com/rs/zerolog"
	"github.com/scmhub/ibsync"
)

const (
	connTimeout    = 10 * time.Second
	sysTimeRefresh = 250 * time.Millisecond
	oneThousand    = 1_000
	oneMillion     = 1_000_000
)

var ErrContractRequired = errors.New("include at least one contract to start ibtui")

// IBService continuously gets the IB account state and makes it suitable for TUI consumption.
type IBService struct {
	ib     *ibsync.IB
	logger *slog.Logger
	wg     sync.WaitGroup

	Contracts []*ibsync.Contract

	systemTime       atomic.Pointer[time.Time]
	cancelSystemTime context.CancelFunc

	tickers       atomic.Pointer[[]*ibsync.Ticker]
	cancelTickers context.CancelFunc

	bars          atomic.Pointer[[]ibsync.Bar]
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
	ibs := &IBService{
		Contracts:        contracts,
		logger:           logger,
		cancelSystemTime: func() {},
		cancelTickers:    func() {},
		cancelBars:       func() {},
		cancelReqHist:    func() {},
	}
	placeholder := time.Now()
	ibs.systemTime.Store(&placeholder)
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
func (ibs *IBService) ReadTime() time.Time {
	return *ibs.systemTime.Load()
}

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
				systemTime := time.Unix(seconds, nanoseconds)
				ibs.systemTime.Store(&systemTime)
			}
		}
	}()
}

// Tickers
func (ibs *IBService) ReadTickers() []*ibsync.Ticker {
	return *ibs.tickers.Load()
}

func (ibs *IBService) startTickers() {
	ibs.stopTickers()
	var ctx context.Context
	ctx, ibs.cancelTickers = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		tickers := make([]*ibsync.Ticker, 0, len(ibs.Contracts))
		for _, c := range ibs.Contracts {
			tickers = append(tickers, ibs.ib.ReqMktData(c, ""))
		}
		ibs.tickers.Store(&tickers)
		// wait for Cancel and then cancel tickers
		<-ctx.Done()
		for _, c := range ibs.Contracts {
			ibs.ib.CancelMktData(c)
		}
	}()
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

func (ibs *IBService) StartBars(contract *ibsync.Contract) {
	ibs.stopBars()
	var ctx context.Context
	ctx, ibs.cancelBars = context.WithCancel(context.Background())
	ibs.wg.Add(1)
	go func() {
		defer ibs.wg.Done()
		reqHistDone := make(chan struct{})
		go func() {
			ibs.barChan, ibs.cancelReqHist = ibs.ib.ReqHistoricalDataUpToDate(contract, "1 D", "1 hour", "TRADES", false, 1)
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
					currentBars := *ibs.bars.Load()
					newBars := make([]ibsync.Bar, len(currentBars))
					if len(currentBars) > 0 && currentBars[len(currentBars)-1].Date == bar.Date {
						copy(newBars, currentBars)
						newBars[len(newBars)-1] = bar
						ibs.bars.Store(&newBars)
					} else {
						newBars = append(currentBars, bar)
						ibs.bars.Store(&newBars)
					}
				}
			}
		}
	}()
}

func (ibs *IBService) stopBars() {
	ibs.cancelReqHist()
	ibs.cancelBars()
	emptyBars := make([]ibsync.Bar, 0)
	ibs.bars.Store(&emptyBars)
}
