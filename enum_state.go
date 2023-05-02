package vsrpc

import (
	"encoding"
	"fmt"
)

type state uint32

const (
	stateRunning state = iota
	stateShuttingDown
	stateGoingAway
	stateClosed
)

var stateGoNames = [...]string{
	"vsrpc.stateRunning",
	"vsrpc.stateShuttingDown",
	"vsrpc.stateGoingAway",
	"vsrpc.stateClosed",
}

var stateNames = [...]string{
	"running",
	"shuttingDown",
	"goingAway",
	"closed",
}

func (enum state) GoString() string {
	if enum < state(len(stateGoNames)) {
		return stateGoNames[enum]
	}
	return fmt.Sprintf("vsrpc.state(%d)", uint32(enum))
}

func (enum state) String() string {
	if enum < state(len(stateNames)) {
		return stateNames[enum]
	}
	return fmt.Sprintf("#%d", uint32(enum))
}

func (enum state) MarshalText() ([]byte, error) {
	str := enum.String()
	return []byte(str), nil
}

var (
	_ fmt.GoStringer         = state(0)
	_ fmt.Stringer           = state(0)
	_ encoding.TextMarshaler = state(0)
)
