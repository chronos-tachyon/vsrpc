package vsrpc

import (
	"fmt"
)

type PanicError struct {
	Err error
}

func (err PanicError) Error() string {
	return fmt.Sprintf("panic! %v", err.Err)
}

func (err PanicError) Unwrap() error {
	return err.Err
}

var (
	_ error           = PanicError{}
	_ unwrapInterface = PanicError{}
)
