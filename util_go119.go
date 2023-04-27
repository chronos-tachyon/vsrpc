//go:build !go1.20
// +build !go1.20

package vsrpc

func joinErrors(errs []error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
