package vsrpc

type unwrapInterface interface {
	error
	Unwrap() error
}
