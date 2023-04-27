package vsrpc

import (
	"encoding"
	"fmt"
)

type State uint32

const (
	StateRunning State = iota
	StateShuttingDown
	StateGoingAway
	StateClosed
)

var stateGoNames = [...]string{
	"vsrpc.StateRunning",
	"vsrpc.StateShuttingDown",
	"vsrpc.StateGoingAway",
	"vsrpc.StateClosed",
}

var stateNames = [...]string{
	"Running",
	"ShuttingDown",
	"GoingAway",
	"Closed",
}

func (enum State) GoString() string {
	if enum < State(len(stateGoNames)) {
		return stateGoNames[enum]
	}
	return fmt.Sprintf("vsrpc.State(%d)", uint32(enum))
}

func (enum State) String() string {
	if enum < State(len(stateNames)) {
		return stateNames[enum]
	}
	return fmt.Sprintf("#%d", uint32(enum))
}

func (enum State) MarshalText() ([]byte, error) {
	str := enum.String()
	return []byte(str), nil
}

var (
	_ fmt.GoStringer         = State(0)
	_ fmt.Stringer           = State(0)
	_ encoding.TextMarshaler = State(0)
)