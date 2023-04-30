package vsrpc

import (
	"fmt"
)

type CallIdError struct {
	Type Frame_Type
	ID   ID
}

func (err CallIdError) Error() string {
	return fmt.Sprintf("%v frame with unexpected call_id %d", err.Type, err.ID)
}

var _ error = CallIdError{}
