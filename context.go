package vsrpc

import (
	"context"
)

type clientKey struct{}
type serverKey struct{}
type connKey struct{}
type callKey struct{}

func ContextClient(ctx context.Context) *Client {
	if value := ctx.Value(clientKey{}); value != nil {
		return value.(*Client)
	}
	return nil
}

func ContextServer(ctx context.Context) *Server {
	if value := ctx.Value(serverKey{}); value != nil {
		return value.(*Server)
	}
	return nil
}

func ContextConn(ctx context.Context) *Conn {
	if value := ctx.Value(connKey{}); value != nil {
		return value.(*Conn)
	}
	return nil
}

func ContextCall(ctx context.Context) *Call {
	if value := ctx.Value(callKey{}); value != nil {
		return value.(*Call)
	}
	return nil
}

func WithContextClient(ctx context.Context, c *Client) context.Context {
	return context.WithValue(ctx, clientKey{}, c)
}

func WithContextServer(ctx context.Context, s *Server) context.Context {
	return context.WithValue(ctx, serverKey{}, s)
}

func WithContextConn(ctx context.Context, conn *Conn) context.Context {
	return context.WithValue(ctx, connKey{}, conn)
}

func WithContextCall(ctx context.Context, call *Call) context.Context {
	return context.WithValue(ctx, callKey{}, call)
}
