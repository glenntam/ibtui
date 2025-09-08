package state

import (
	// "fmt"
	"log/slog"
	"time"

	"github.com/scmhub/ibsync"
)

type IBState struct {
	CurrentTime    time.Time
}

func NewIBState() *IBState {
	return &IBState{
		CurrentTime: time.Now(),
	}
}

func (s *IBState) reqCurrentTime(ib *ibsync.IB) {
	t, err := ib.ReqCurrentTime()
	if err != nil {
		slog.Error("Couldn't request IB time, using system time instead", "error", err)
		t = time.Now()
	}
	s.CurrentTime = t
}

func (s *IBState) ReqCurrentTimeMilli(ib *ibsync.IB, timezone string) {
	t := time.Now()
	m, err := ib.ReqCurrentTimeInMillis()
	if err != nil {
		slog.Error("Couldn't request IB time, using system time instead", "error", err)
	} else {
		seconds := m / 1000
		nanoseconds := (m % 1000) * 1_000_000
		t = time.Unix(seconds, nanoseconds)
	}
	tz, err := time.LoadLocation(timezone)
	if err != nil {
		slog.Error("Couldn't find "+timezone+" time zone.", "error", err)
		tz = time.UTC
	}
	s.CurrentTime = t.In(tz)
}
