package vsrpc

import (
	"fmt"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtocolViolationError struct {
	Err error
}

func (err ProtocolViolationError) Error() string {
	return fmt.Sprintf("protocol violation: %v", err.Err)
}

func (err ProtocolViolationError) Unwrap() error {
	return err.Err
}

func (err ProtocolViolationError) IsRecoverable() bool {
	return false
}

var (
	_ error                  = ProtocolViolationError{}
	_ unwrapInterface        = ProtocolViolationError{}
	_ isRecoverableInterface = ProtocolViolationError{}
)

type FrameTypeError struct {
	Type Frame_Type
}

func (err FrameTypeError) Error() string {
	return fmt.Sprintf("unexpected %v frame", err.Type)
}

var (
	_ error = FrameTypeError{}
)

type CallIdError struct {
	Type Frame_Type
	ID   ID
}

func (err CallIdError) Error() string {
	return fmt.Sprintf("%v frame with unexpected call_id %d", err.Type, err.ID)
}

var (
	_ error = CallIdError{}
)

type DuplicateCallError struct {
	ID  ID
	Old Method
	New Method
}

func (err DuplicateCallError) Error() string {
	return fmt.Sprintf("%v frame with duplicate call_id %d: old method %q, new method %q", Frame_BEGIN, err.ID, err.Old, err.New)
}

var (
	_ error = DuplicateCallError{}
)

type MessageTypeError struct {
	Actual protoreflect.FullName
	Expect protoreflect.FullName
}

func (err MessageTypeError) Error() string {
	return fmt.Sprintf("expected message of type %q, but got message of type %q", err.Expect, err.Actual)
}

func (err MessageTypeError) As(out any) bool {
	switch x := out.(type) {
	case *StatusError:
		text := err.Error()
		x.Status = &Status{Code: Status_INTERNAL, Text: text}
		return true

	default:
		return false
	}
}

var (
	_ error       = MessageTypeError{}
	_ asInterface = MessageTypeError{}
)

type UnmarshalError struct {
	Type protoreflect.FullName
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
		text := err.Error()
		x.Status = &Status{Code: Status_INTERNAL, Text: text}
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
