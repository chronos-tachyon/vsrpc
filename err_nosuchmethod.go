package vsrpc

import (
	"fmt"
)

type NoSuchMethodError struct {
	Method Method
}

func (err NoSuchMethodError) Error() string {
	return fmt.Sprintf("method %q is not implemented", err.Method)
}

func (err NoSuchMethodError) As(out any) bool {
	switch x := out.(type) {
	case *StatusError:
		x.Status = &Status{
			Code: Status_UNIMPLEMENTED,
			Text: err.Error(),
		}
		return true

	default:
		return false
	}
}

var (
	_ error       = NoSuchMethodError{}
	_ asInterface = NoSuchMethodError{}
)
