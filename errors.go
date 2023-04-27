package vsrpc

import (
	"errors"
	"fmt"
	"io/fs"
	"strconv"

	anypb "google.golang.org/protobuf/types/known/anypb"

	"github.com/chronos-tachyon/vsrpc/prettyprinter"
)

type unwrapInterface interface {
	error
	Unwrap() error
}

type asInterface interface {
	As(any) bool
}

type PanicError struct {
	Err error
}

func (err PanicError) Error() string {
	return fmt.Sprintf("panic! %v", err.Err)
}

func (err PanicError) Unwrap() error {
	return err.Err
}

var (
	_ error           = PanicError{}
	_ unwrapInterface = PanicError{}
)

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

type UnexpectedFrameTypeError struct {
	Actual   Frame_Type
	Expected []Frame_Type
}

func (err UnexpectedFrameTypeError) Error() string {
	return fmt.Sprintf("unexpected frame of type %v; expected one of: %v", err.Actual, err.Expected)
}

var _ error = UnexpectedFrameTypeError{}

type ClientClosingError struct{}

func (err ClientClosingError) Error() string {
	return "connection is closing: client has already sent SHUTDOWN frame"
}

var ErrClientClosing error = ClientClosingError{}

type ServerClosingError struct{}

func (err ServerClosingError) Error() string {
	return "connection is closing: server has already sent GO_AWAY frame"
}

var ErrServerClosing error = ServerClosingError{}

type HalfClosedError struct{}

func (err HalfClosedError) Error() string {
	return "RPC call is half-closed: client has already sent HALF_CLOSE or CANCEL frame"
}

var ErrHalfClosed error = HalfClosedError{}

type CallCompleteError struct{}

func (err CallCompleteError) Error() string {
	return "RPC call has completed: server has already sent END frame"
}

var ErrCallComplete error = CallCompleteError{}

var ErrClosed = fs.ErrClosed

type NoSuchMethodError struct {
	Method string
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

type TypeError struct {
	Actual string
	Expect string
}

func (err TypeError) Error() string {
	return fmt.Sprintf("expected message of type %q, but got message of type %q", err.Expect, err.Actual)
}

func (err TypeError) As(out any) bool {
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
	_ error       = TypeError{}
	_ asInterface = TypeError{}
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

type StatusError struct {
	Status *Status
}

func (err StatusError) Error() string {
	if err.Status.IsOK() {
		return "OK[0]"
	}

	codeNum := int32(err.Status.Code)
	codeName := "???"
	if str, found := Status_Code_name[codeNum]; found {
		codeName = str
	}

	buf := make([]byte, 0, 1024)
	buf = append(buf, codeName...)
	buf = append(buf, '[')
	buf = strconv.AppendUint(buf, uint64(codeNum), 10)
	buf = append(buf, ']')
	if err.Status.Text != "" {
		buf = append(buf, ':', ' ')
		buf = append(buf, err.Status.Text...)
	}
	for _, detail := range err.Status.Details {
		buf = prettyprinter.PrettyPrintTo(buf, detail)
	}
	return string(buf)
}

var _ error = StatusError{}

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
