// Package FSM implements a simple finite state machine framework.
//
// States are implemented via constants. It is up to the using object to define
// states and how to use them.
package fsm

import (

)

// States should typically be declared using a const iota-type declaration.
type State string

// Object holding a state, implements methods for manipulating the state.
type FSM struct {
	state State
}

func MakeFSM(s State) *FSM {
	return &FSM{s}
}

// Set the state of the FSM
func (fsm *FSM) SetState(s State) {
	fsm.state = s
}

// Get the state of the FSM
func (fsm *FSM) GetState() State {
	return fsm.state
}

// Check if we're in a state
func (fsm *FSM) GetStateString() string {
	return string(fsm.state)
}