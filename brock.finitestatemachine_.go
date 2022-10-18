package brock

import "sync"

//nolint:gochecknoglobals
var FiniteStateMachine fsm

type fsm struct {
	*sync.Mutex
	curr string
	FSMOnTransition
	FSMTransition
}

type FSM interface {
	CurrentState() string
	Next(action string) (next string, ok bool)
}

type FSMTransition map[string]map[string]string

type FSMOnTransition func(state, action, nextState string)

func (fsm) Create(init string, fn FSMOnTransition, tx FSMTransition) (FSM, error) {
	if tx == nil || len(tx) < 1 {
		return nil, ErrFSMEmptyStates
	} else if _, ok := tx[init]; !ok {
		return nil, ErrFSMNoInitialStates
	}

	for _, m := range tx {
		for _, next := range m {
			if _, ok := tx[next]; !ok {
				return nil, Errorf("brock: fsm: dangling state: \"%s\"", next)
			}
		}
	}

	if fn == nil {
		fn = func(state, action, next string) { Nop(state, action, next) }
	}

	return &fsm{new(sync.Mutex), init, fn, tx}, nil
}

func (x *fsm) CurrentState() string { return x.curr }

func (x *fsm) Next(action string) (nextState string, ok bool) {
	var m map[string]string
	if m, ok = x.FSMTransition[x.curr]; ok {
		if nextState, ok = m[action]; ok {
			x.FSMOnTransition(x.curr, action, nextState)
			func() { x.Lock(); x.curr = nextState; x.Unlock() }()
		}
	}

	return nextState, ok
}
