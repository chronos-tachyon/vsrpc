package vsrpc

import (
	"fmt"
)

type ProtocolViolationError struct {
	Err error
}

func (err ProtocolViolationError) Error() string {
	return fmt.Sprintf("protocol violation: %v", err.Err)
}

func (err ProtocolViolationError) Unwrap() error {
	return err.Err
}

func (err ProtocolViolationError) IsRecoverable() bool {
	return false
}

var (
	_ error                  = ProtocolViolationError{}
	_ unwrapInterface        = ProtocolViolationError{}
	_ isRecoverableInterface = ProtocolViolationError{}
)
