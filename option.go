package vsrpc

type Option interface {
	applyToClient(c *Client)
	applyToServer(s *Server)
	applyToConn(conn *Conn)
	applyToCall(call *Call)
}

func ConcatOptions(a []Option, b ...Option) []Option {
	var n uint
	for _, opt := range a {
		if opt != nil {
			n++
		}
	}
	for _, opt := range b {
		if opt != nil {
			n++
		}
	}
	if n <= 0 {
		return nil
	}

	out := make([]Option, 0, n)
	for _, opt := range a {
		if opt != nil {
			out = append(out, opt)
		}
	}
	for _, opt := range b {
		if opt != nil {
			out = append(out, opt)
		}
	}
	return out
}
