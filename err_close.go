package vsrpc

import (
	"io/fs"
)

type CloseError struct {
	Type CloseType
}

func (err CloseError) Error() string {
	return err.Type.Message()
}

func (err CloseError) Is(target error) bool {
	return target == fs.ErrClosed
}

var (
	_ error       = CloseError{}
	_ isInterface = CloseError{}
)

var (
	ErrClientClosing    error = CloseError{Type: CloseType_ClientShutdown}
	ErrClientClosed     error = CloseError{Type: CloseType_ClientClose}
	ErrServerClosing    error = CloseError{Type: CloseType_ServerShutdown}
	ErrServerClosed     error = CloseError{Type: CloseType_ServerClose}
	ErrConnShuttingDown error = CloseError{Type: CloseType_ConnShutdown}
	ErrConnGoingAway    error = CloseError{Type: CloseType_ConnGoAway}
	ErrConnClosed       error = CloseError{Type: CloseType_ConnClose}
	ErrHalfClosed       error = CloseError{Type: CloseType_CallHalfClose}
	ErrCallClosed       error = CloseError{Type: CloseType_CallClose}
)
