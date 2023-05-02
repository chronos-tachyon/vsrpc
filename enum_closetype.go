package vsrpc

import (
	"encoding"
	"fmt"
)

type CloseType uint

const (
	CloseType_Unknown CloseType = iota
	CloseType_ClientShutdown
	CloseType_ClientClose
	CloseType_ServerShutdown
	CloseType_ServerClose
	CloseType_ConnShutdown
	CloseType_ConnGoAway
	CloseType_ConnClose
	CloseType_CallHalfClose
	CloseType_CallClose
)

var closeTypeGoNames = [...]string{
	"vsrpc.CloseType_Unknown",
	"vsrpc.CloseType_ClientShutdown",
	"vsrpc.CloseType_ClientClose",
	"vsrpc.CloseType_ServerShutdown",
	"vsrpc.CloseType_ServerClose",
	"vsrpc.CloseType_ConnShutdown",
	"vsrpc.CloseType_ConnGoAway",
	"vsrpc.CloseType_ConnClose",
	"vsrpc.CloseType_CallHalfClose",
	"vsrpc.CloseType_CallClose",
}

var closeTypeNames = [...]string{
	"unknown",
	"clientShutdown",
	"clientClose",
	"serverShutdown",
	"serverClose",
	"connShutdown",
	"connGoAway",
	"connClose",
	"callHalfClose",
	"callClose",
}

var closeTypeMessages = [...]string{
	"already closed",
	"client is shutting down",
	"client is closed",
	"server is shutting down",
	"server is closed",
	"connection is closing, client has sent SHUTDOWN frame",
	"connection is closing, server has sent GO_AWAY frame",
	"connection is closed",
	"RPC call is half-closed, client has sent HALF_CLOSE frame",
	"RPC call is closed, server has sent END frame",
}

func (enum CloseType) GoString() string {
	if enum < CloseType(len(closeTypeGoNames)) {
		return closeTypeGoNames[enum]
	}
	return fmt.Sprintf("vsrpc.CloseType(%d)", uint(enum))
}

func (enum CloseType) String() string {
	if enum < CloseType(len(closeTypeNames)) {
		return closeTypeNames[enum]
	}
	return fmt.Sprintf("#%d", uint(enum))
}

func (enum CloseType) Message() string {
	if enum < CloseType(len(closeTypeMessages)) {
		return closeTypeMessages[enum]
	}
	return closeTypeMessages[0]
}

func (enum CloseType) MarshalText() ([]byte, error) {
	str := enum.String()
	return []byte(str), nil
}

var (
	_ fmt.GoStringer         = CloseType(0)
	_ fmt.Stringer           = CloseType(0)
	_ encoding.TextMarshaler = CloseType(0)
)
