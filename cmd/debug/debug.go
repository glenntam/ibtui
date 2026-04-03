// Package debug is used for throwaway code to test ibsync behavior
package main

import (
	"fmt"
	//"os"
	// "time"

	"github.com/scmhub/ibsync"
)

const (
	logLinesDisplayed = 10
	logFilePermission = 0o600 // RW for owner only
)

func main() {

	log := ibsync.Logger()
	ibsync.SetLogLevel(2)
	ibsync.SetConsoleWriter()

	// New IB Client
	ib := ibsync.NewIB()

	// Connect
	err := ib.Connect(
		ibsync.NewConfig(
			ibsync.WithHost("127.0.0.1"),
			ibsync.WithPort(4001),
			ibsync.WithClientID(0),
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("Connect")
		return
	}
	defer ib.Disconnect()

	//
	fut := ibsync.NewFuture("MNQ", "202606", "CME", "2", "USD")
	err = ib.QualifyContract(fut)
	if err != nil {
		panic(err)
	}

	// endDateTime := ""      // "yyyymmdd HH:mm:ss ttt", where "ttt" is an optional time zone
	duration := "1 D"      // "60 S", "30 D", "13 W", "6 M", "10 Y". The unit must be specified (S for seconds, D for days, W for weeks, etc.).
	barSize := "1 hour"    // "1 secs", "5 secs", "10 secs", "15 secs", "30 secs", "1 min", "2 mins", "5 mins", etc.
	whatToShow := "TRADES" // "TRADES", "MIDPOINT", "BID", "ASK", "BID_ASK", "HISTORICAL_VOLATILITY", etc.
	useRTH := false        // `true` limits data to regular trading hours (RTH), `false` includes all data.
	formatDate := 1        // `1` for the "yyyymmdd HH:mm:ss ttt" format, or `2` for Unix timestamps.
	// barChan, _ := ib.ReqHistoricalData(eurusd, endDateTime, duration, barSize, whatToShow, useRTH, formatDate)

	var bars []ibsync.Bar
	// for bar := range barChan {
	//     fmt.Println(bar)
	//     bars = append(bars, bar)
	// }

	// fmt.Println("Number of bars:", len(bars))
	// fmt.Println("First Bar", bars[0])
	// fmt.Println("Last Bar", bars[len(bars)-1])

	// Historical Data with realtime Updates
	duration = "1 D"
	barSize = "1 hour"
	barChan, _ := ib.ReqHistoricalData(fut, "", duration, barSize, whatToShow, useRTH, formatDate)

	for bar := range barChan {
		fmt.Println(bar)
		bars = append(bars, bar)
	}

	fmt.Println("Number of bars:", len(bars))
	fmt.Println("First Bar", bars[0])
	fmt.Println("Last Bar", bars[len(bars)-1])


	// go func() {
	//     for bar := range barChan {
	//         fmt.Println(bar)
	//         bars = append(bars, bar)
	//     }
	// }()

	// time.Sleep(10 * time.Second)
	// cancel()

	// Historical schedule
	// historicalSchedule, err := ib.ReqHistoricalSchedule(eurusd, endDateTime, duration, useRTH)
	// if err != nil {
	//     log.Error().Err(err).Msg("ReqHistoricalSchedule")
	//     return
	// }

	// fmt.Printf("Historical schedule start date: %v, end date: %v, time zone: %v\n", historicalSchedule.StartDateTime, historicalSchedule.EndDateTime, historicalSchedule.TimeZone)
	// for i, session := range historicalSchedule.Sessions {
	//     fmt.Printf("session %v: %v\n", i, session)
	// }

	// // Real time bars
	// whatToShow = "MIDPOINT" // "TRADES", "MIDPOINT", "BID" or "ASK"
	// rtBarChan, cancel := ib.ReqRealTimeBars(eurusd, 5, whatToShow, useRTH)

	// var rtBars []ibsync.RealTimeBar
	// go func() {
	//     for rtBar := range rtBarChan {
	//         fmt.Println(rtBar)
	//         rtBars = append(rtBars, rtBar)
	//     }

	// }()

	// time.Sleep(10 * time.Second)
	// cancel()

	// fmt.Println("Number of RT bars:", len(rtBars))
	// fmt.Println("First RT Bar", rtBars[0])
	// fmt.Println("Last RT Bar", rtBars[len(rtBars)-1])

	// time.Sleep(1 * time.Second)
	// log.Info().Msg("Good Bye!!!")
}
