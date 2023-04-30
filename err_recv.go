package vsrpc

type EmptyRecvError struct{}

func (EmptyRecvError) Error() string {
	return "empty recv"
}

var ErrEmptyRecv error = EmptyRecvError{}
