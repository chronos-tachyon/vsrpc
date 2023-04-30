package vsrpc

import "errors"

type isRecoverableInterface interface {
	error
	IsRecoverable() bool
}

func IsRecoverable(err error) bool {
	for err != nil {
		if xerr, ok := err.(isRecoverableInterface); ok {
			return xerr.IsRecoverable()
		}
		err = errors.Unwrap(err)
	}
	return false
}

type RecoverableError struct {
	Err error
}

func (err RecoverableError) Error() string {
	return err.Err.Error()
}

func (err RecoverableError) Unwrap() error {
	return err.Err
}

func (err RecoverableError) IsRecoverable() bool {
	return true
}

var (
	_ error                  = RecoverableError{}
	_ unwrapInterface        = RecoverableError{}
	_ isRecoverableInterface = RecoverableError{}
)

type UnrecoverableError struct {
	Err error
}

func (err UnrecoverableError) Error() string {
	return err.Err.Error()
}

func (err UnrecoverableError) Unwrap() error {
	return err.Err
}

func (err UnrecoverableError) IsRecoverable() bool {
	return false
}

var (
	_ error                  = UnrecoverableError{}
	_ unwrapInterface        = UnrecoverableError{}
	_ isRecoverableInterface = UnrecoverableError{}
)
