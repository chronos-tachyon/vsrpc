package vsrpc

import (
	"errors"

	"google.golang.org/protobuf/types/known/anypb"
)

type unwrapInterface interface {
	error
	Unwrap() error
}

type isInterface interface {
	error
	Is(error) bool
}

type asInterface interface {
	error
	As(any) bool
}

type isRecoverableInterface interface {
	error
	IsRecoverable() bool
}

func IsRecoverable(err error) bool {
	for err != nil {
		if result, ok := isErrRecoverable(err); ok {
			return result
		}
		err = errors.Unwrap(err)
	}
	return false
}

func isErrRecoverable(err error) (bool, bool) {
	if xerr, ok := err.(isRecoverableInterface); ok {
		return xerr.IsRecoverable(), true
	}

	return false, false
}

type appendDetailsInterface interface {
	error
	AppendDetails([]*anypb.Any) []*anypb.Any
}

func AppendDetails(details []*anypb.Any, err error) []*anypb.Any {
	var scratch [8]appendDetailsInterface
	stack := scratch[:0]
	for err != nil {
		if xerr, ok := err.(appendDetailsInterface); ok {
			stack = append(stack, xerr)
		}
		err = errors.Unwrap(err)
	}
	i := uint(len(stack))
	for i > 0 {
		i--
		details = stack[i].AppendDetails(details)
	}
	return details
}
