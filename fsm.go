package ssfsm

import "fmt"

// Transition represents a transition
type Transition struct {
	Event, From, To string
}

// FSM is a finite state machine
type FSM struct {
	mx    *chan struct{}
	state string
	graph map[string]map[string]next
}

type next struct {
	transition Transition
	callback   func(*FSM, Transition)
}

// Trigger triggers an event. If async is set to true, then Trigger would block.
// And if any other transitions are in progress, it will return an error.
func (sm *FSM) Trigger(event string) error {
	if sm.mx != nil {
		select {
		case *sm.mx <- struct{}{}:
			defer func() { <-*sm.mx }()
		default:
			return fmt.Errorf("another transition is in progress")
		}
	}
	vx, ok := sm.graph[event]
	if !ok {
		return fmt.Errorf("event not found: " + event)
	}
	nx, ok := vx[sm.state]
	if !ok {
		return fmt.Errorf("no state found for event: " + event + " current state: " + sm.state)
	}
	starting := sm.state
	defer func() {
		if starting != sm.state {
			return
		}
		sm.state = nx.transition.To
	}()
	if nx.callback == nil {
		return nil
	}
	nx.callback(sm, nx.transition)
	return nil
}

// Table represents a transition table between states
type Table map[Transition]func(*FSM, Transition)

// NewFSM creates an instance of FSM. If async is set to true, then Trigger would block.
// And if any other transitions are in progress, it will return an error.
func NewFSM(async bool, state string, table Table) *FSM {
	// event -> from -> (to, callback)
	var graph = make(map[string]map[string]next)

	for gk, gv := range table {
		vx, ok := graph[gk.Event]
		if !ok {
			vx = make(map[string]next)
		}
		vx[gk.From] = next{gk, gv}
		graph[gk.Event] = vx
	}

	res := &FSM{
		state: state,
		graph: graph,
	}
	if async {
		cb := make(chan struct{}, 1)
		res.mx = &cb
	}

	return res
}