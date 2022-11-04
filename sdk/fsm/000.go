package sdkfsm

import (
	"sync"

	"github.com/brick-io/brock/sdk"
)

var (
	ErrEmptyStates     = sdk.Errorf("brock/fsm: empty states")
	ErrNoInitialStates = sdk.Errorf("brock/fsm: no initial states")
)

type fsm struct {
	*sync.Mutex
	curr string
	OnTransition
	TransitionTable
}

type FSM interface {
	CurrentState() string
	Next(action string) (next string, ok bool)
}

type TransitionTable map[string]map[string]string

type OnTransition func(state, action, nextState string)

func New(init string, fn OnTransition, tx TransitionTable) (FSM, error) {
	if tx == nil || len(tx) < 1 {
		return nil, ErrEmptyStates
	} else if _, ok := tx[init]; !ok {
		return nil, ErrNoInitialStates
	}

	for _, m := range tx {
		for _, next := range m {
			if _, ok := tx[next]; !ok {
				return nil, sdk.Errorf("brock: fsm: dangling state: \"%s\"", next)
			}
		}
	}

	if fn == nil {
		fn = func(state, action, next string) { sdk.Nop(state, action, next) }
	}

	return &fsm{new(sync.Mutex), init, fn, tx}, nil
}

func (x *fsm) CurrentState() string { return x.curr }

func (x *fsm) Next(action string) (nextState string, ok bool) {
	var m map[string]string
	if m, ok = x.TransitionTable[x.curr]; ok {
		if nextState, ok = m[action]; ok {
			x.OnTransition(x.curr, action, nextState)
			func() { x.Lock(); x.curr = nextState; x.Unlock() }()
		}
	}

	return nextState, ok
}
