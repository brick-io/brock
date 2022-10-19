package main

import (
	"sync/atomic"

	"go.onebrick.io/brock"
)

func main() {
	cb := new(circuit_breaker)
	cb.FSM, _ = brock.FiniteStateMachine.Create("close", cb.OnTransition, brock.FSMTransition{
		"close": {"half": "half", "open": "open"},
		"half":  {"close": "close", "open": "open"},
		"open":  {"half": "half", "close": "close"},
	})
}

var (
	// ErrTooManyRequests is returned when the CB state is half open and the requests count is over the cb maxRequests.
	ErrTooManyRequests = brock.Errorf("too many requests")
	// ErrOpenState is returned when the CB state is open.
	ErrOpenState = brock.Errorf("circuit breaker is open")
)

type circuit_breaker struct {
	brock.FSM
	c circuit_breaker_counter
}

func (x *circuit_breaker) OnTransition(state, action, nextState string) { brock.Nop() }

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
