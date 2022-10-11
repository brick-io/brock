package brock_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"go.onebrick.io/brock"
)

func testFSM(t *testing.T) {
	t.Parallel()

	Expect := NewWithT(t).Expect
	{
		fsm, err := brock.FiniteStateMachine.Create("", nil, brock.FSMTransition{})
		Expect(fsm).To(BeNil())
		Expect(err).To(Equal(brock.ErrFSMEmptyStates))
	}

	{
		fsm, err := brock.FiniteStateMachine.Create("", nil, brock.FSMTransition{"init": {}})
		Expect(fsm).To(BeNil())
		Expect(err).To(Equal(brock.ErrFSMNoInitialStates))
	}

	{
		fsm, err := brock.FiniteStateMachine.Create("init", nil, brock.FSMTransition{"init": {"1": "dangle"}})
		Expect(fsm).To(BeNil())
		Expect(err.Error()).To(Equal("brock: fsm: dangling state: \"dangle\""))
	}

	fsm, err := brock.FiniteStateMachine.Create("init", nil, brock.FSMTransition{
		"init": {
			"start": "started",
		},
		"started": {
			"hold":      "init",
			"in_review": "in_review",
		},
		"in_review": {
			"reject": "rejected",
			"done":   "done",
		},
		"rejected": {},
		"done":     {},
	})

	Expect(err).To(Succeed())
	Expect(fsm.CurrentState()).To(Equal("init"))

	next, ok := fsm.Next("start")
	Expect(ok).To(BeTrue())
	Expect(next).To(Equal("started"))
	Expect(fsm.CurrentState()).To(Equal("started"))

	next, ok = fsm.Next("")
	Expect(ok).To(BeFalse())
	Expect(next).To(Equal(""))
	Expect(fsm.CurrentState()).To(Equal("started"))

	next, ok = fsm.Next("hold")
	Expect(ok).To(BeTrue())
	Expect(next).To(Equal("init"))
	Expect(fsm.CurrentState()).To(Equal("init"))

	next, ok = fsm.Next("start")
	Expect(ok).To(BeTrue())
	Expect(next).To(Equal("started"))
	Expect(fsm.CurrentState()).To(Equal("started"))

	next, ok = fsm.Next("in_review")
	Expect(ok).To(BeTrue())
	Expect(next).To(Equal("in_review"))
	Expect(fsm.CurrentState()).To(Equal("in_review"))

	next, ok = fsm.Next("done")
	Expect(ok).To(BeTrue())
	Expect(next).To(Equal("done"))
	Expect(fsm.CurrentState()).To(Equal("done"))
}
