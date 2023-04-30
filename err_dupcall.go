package vsrpc

import (
	"fmt"
)

type DuplicateCallError struct {
	ID  ID
	Old Method
	New Method
}

func (err DuplicateCallError) Error() string {
	return fmt.Sprintf("%v frame with duplicate call_id %d: old method %q, new method %q", Frame_BEGIN, err.ID, err.Old, err.New)
}

var _ error = DuplicateCallError{}
