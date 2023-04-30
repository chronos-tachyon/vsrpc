package vsrpc

import (
	"errors"

	anypb "google.golang.org/protobuf/types/known/anypb"
)

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
