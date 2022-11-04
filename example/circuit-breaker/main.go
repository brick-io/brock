package main

import (
	"sync/atomic"

	"github.com/brick-io/brock/sdk"
	sdkfsm "github.com/brick-io/brock/sdk/fsm"
)

func main() {
	cb := new(circuit_breaker)
	cb.FSM, _ = sdkfsm.New("close", cb.OnTransition, sdkfsm.TransitionTable{
		"close": {"half": "half", "open": "open"},
		"half":  {"close": "close", "open": "open"},
		"open":  {"half": "half", "close": "close"},
	})
}

var (
	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests.
	ErrTooManyRequests = sdk.Errorf("too many requests")
	// ErrOpenState is returned when the CB state is open.
	ErrOpenState = sdk.Errorf("circuit breaker is open")
)

type circuit_breaker struct {
	sdkfsm.FSM
	c circuit_breaker_counter
}

func (x *circuit_breaker) OnTransition(state, action, nextState string) { sdk.Nop() }

func (x *circuit_breaker) OnBefore() error {
	if x.FSM.CurrentState() == "open" {
		return ErrOpenState
	}

	return nil
}

func (x *circuit_breaker) OnAfter(err error) error {
	add, load := atomic.AddUint32, atomic.LoadUint32
	add(&x.c.Requests, 1)

	if err != nil {
		add(&x.c.TotalFailures, 1)
		add(&x.c.ConsecutiveFailures, 1)
		add(&x.c.ConsecutiveSuccesses, -load(&x.c.ConsecutiveSuccesses))
	} else {
		add(&x.c.TotalSuccesses, 1)
		add(&x.c.ConsecutiveSuccesses, 1)
		add(&x.c.ConsecutiveFailures, -load(&x.c.ConsecutiveFailures))
	}

	return err
}

type circuit_breaker_counter struct {
	Requests             uint32
	MaxRequests          uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}
