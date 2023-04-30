package vsrpc

import (
	"fmt"
)

type InappropriateError struct {
	Op   string
	Role Role
}

func (err InappropriateError) Error() string {
	return fmt.Sprintf("operation %v is inappropriate for context %v", err.Op, err.Role)
}

var _ error = InappropriateError{}
