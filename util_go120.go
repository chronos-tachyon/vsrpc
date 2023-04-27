//go:build go1.20
// +build go1.20

package vsrpc

import (
	"errors"
)

func joinErrors(errs []error) error {
	return errors.Join(errs...)
}
