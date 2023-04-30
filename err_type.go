package vsrpc

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type TypeError struct {
	Actual protoreflect.FullName
	Expect protoreflect.FullName
}

func (err TypeError) Error() string {
	return fmt.Sprintf("expected message of type %q, but got message of type %q", err.Expect, err.Actual)
}

func (err TypeError) As(out any) bool {
	switch x := out.(type) {
	case *StatusError:
		x.Status = &Status{Code: Status_INTERNAL, Text: err.Error()}
		return true

	default:
		return false
	}
}

var (
	_ error       = TypeError{}
	_ asInterface = TypeError{}
)
