package vsrpc

import (
	"fmt"
)

type FrameTypeError struct {
	Type Frame_Type
}

func (err FrameTypeError) Error() string {
	return fmt.Sprintf("unexpected %v frame", err.Type)
}

var _ error = FrameTypeError{}
