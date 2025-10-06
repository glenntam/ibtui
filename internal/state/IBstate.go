// Package state polls IB Gateway/TWS to receive current state of account.
package state

import (
	"fmt"
	"time"

	"github.com/scmhub/ibsync"
)

const (
	oneThousand = 1_000
	oneMillion  = 1_000_000
)

// IBState constains the results of polling the IB account state.
type IBState struct {
	CurrentTime time.Time
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
		s.CurrentTime = time.Now()
		return fmt.Errorf("couldn't request IB time, using system time instead: %w", err)
	}
	seconds := m / oneThousand
	nanoseconds := (m % oneThousand) * oneMillion
	s.CurrentTime = time.Unix(seconds, nanoseconds)
	return nil
}
