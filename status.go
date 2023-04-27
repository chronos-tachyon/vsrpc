package vsrpc

import (
	"errors"

	anypb "google.golang.org/protobuf/types/known/anypb"
)

func (status *Status) IsOK() bool {
	return status == nil || status.Code == Status_OK
}

func (status *Status) AsError() error {
	if status.IsOK() {
		return nil
	}
	return StatusError{Status: status}
}

func (status *Status) CopyFrom(src *Status) {
	status.Reset()

	if src.IsOK() {
		return
	}

	status.Code = src.Code
	status.Text = src.Text
	if n := len(src.Details); n > 0 {
		status.Details = make([]*anypb.Any, n)
		copy(status.Details, src.Details)
	}
	status.CanRetry = src.CanRetry
}

func (status *Status) FromError(err error) {
	status.Reset()

	if err == nil {
		return
	}

	if xerr, ok := err.(StatusError); ok {
		status.CopyFrom(xerr.Status)
		return
	}

	var serr StatusError
	if errors.As(err, &serr) {
		status.CopyFrom(serr.Status)
		return
	}

	var pserr *StatusError
	if errors.As(err, &pserr) {
		status.CopyFrom(pserr.Status)
		return
	}

	status.Code = Status_UNKNOWN
	status.Text = err.Error()
	status.Details = AppendDetails(nil, err)
}

func StatusFromError(err error) *Status {
	status := &Status{}
	status.FromError(err)
	return status
}

func Abort(err error) *Status {
	status := &Status{Code: Status_ABORTED}
	if err != nil {
		status.Text = err.Error()
		status.Details = AppendDetails(nil, err)
	}
	return status
}
