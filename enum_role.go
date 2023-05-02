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
