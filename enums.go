package vsrpc

import (
	"encoding"
	"fmt"
)

type Role byte

const (
	UnknownRole Role = iota
	ClientRole
	ServerRole
)

var roleGoNames = [...]string{
	"vsrpc.UnknownRole",
	"vsrpc.ClientRole",
	"vsrpc.ServerRole",
}

var roleNames = [...]string{
	"unknown",
	"client",
	"server",
}

func (enum Role) GoString() string {
	if enum < Role(len(roleGoNames)) {
		return roleGoNames[enum]
	}
	return fmt.Sprintf("vsrpc.Role(%d)", uint32(enum))
}

func (enum Role) String() string {
	if enum < Role(len(roleNames)) {
		return roleNames[enum]
	}
	return fmt.Sprintf("#%d", uint32(enum))
}

func (enum Role) MarshalText() ([]byte, error) {
	str := enum.String()
	return []byte(str), nil
}

var (
	_ fmt.GoStringer         = Role(0)
	_ fmt.Stringer           = Role(0)
	_ encoding.TextMarshaler = Role(0)
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
