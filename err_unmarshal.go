package vsrpc

import (
	"fmt"
)

type UnmarshalError struct {
	Type string
	Err  error
}

func (err UnmarshalError) Error() string {
	return fmt.Sprintf("failed to decode message of type %q: %v", err.Type, err.Err)
}

func (err UnmarshalError) Unwrap() error {
	return err.Err
}

func (err UnmarshalError) As(out any) bool {
	switch x := out.(type) {
	case *StatusError:
		x.Status = &Status{
			Code: Status_INTERNAL,
			Text: err.Error(),
		}
		return true

	default:
		return false
	}
}

var (
	_ error           = UnmarshalError{}
	_ unwrapInterface = UnmarshalError{}
	_ asInterface     = UnmarshalError{}
)
