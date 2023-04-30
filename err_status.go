package vsrpc

import (
	"strconv"

	"github.com/chronos-tachyon/vsrpc/prettyprinter"
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
