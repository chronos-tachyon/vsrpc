package vsrpc

import (
	"context"
)

type Picker interface {
	Pick(ctx context.Context, conns []*Conn) (*Conn, error)
}
