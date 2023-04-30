package vsrpc

import (
	"io/fs"
)

var ErrClosed = fs.ErrClosed

var ErrClientClosed = fs.ErrClosed

var ErrServerClosed = fs.ErrClosed

var ErrConnClosed = fs.ErrClosed

type ClientClosingError struct{}

func (err ClientClosingError) Error() string {
	return "connection is closing: client has already sent SHUTDOWN frame"
}

var ErrClientClosing error = ClientClosingError{}

type ServerClosingError struct{}

func (err ServerClosingError) Error() string {
	return "connection is closing: server has already sent GO_AWAY frame"
}

var ErrServerClosing error = ServerClosingError{}

type HalfClosedError struct{}

func (err HalfClosedError) Error() string {
	return "RPC call is half-closed: client has already sent HALF_CLOSE or CANCEL frame"
}

var ErrHalfClosed error = HalfClosedError{}

type CallCompleteError struct{}

func (err CallCompleteError) Error() string {
	return "RPC call has completed: server has already sent END frame"
}

var ErrCallComplete error = CallCompleteError{}
