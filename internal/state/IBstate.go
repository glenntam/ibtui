// Poll IB Gateway/TWS to receive current state of account.
package state

import (
	"log/slog"
	"time"

	"github.com/scmhub/ibsync"
)

// The primary data structure used to save IB account state.
type IBState struct {
	CurrentTime time.Time
}

// Make a new IBSState struct (used to save IB account state).
func NewIBState() *IBState {
	return &IBState{
		CurrentTime: time.Now(),
	}
}

// Unused. Retrieve IB account system time in seconds.
func (s *IBState) reqCurrentTime(ib *ibsync.IB) {
	t, err := ib.ReqCurrentTime()
	if err != nil {
		slog.Error("Couldn't request IB time, using system time instead", "error", err)
		t = time.Now()
	}
	s.CurrentTime = t
}

// Retrieve IB account system time in time.Time format.
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
	s.CurrentTime = t
}
