// Code generated by protoc-gen-go-vsrpc. DO NOT EDIT.
// Versions:
// - protoc-gen-go-vsrpc: (unknown)
// - protoc: v4.22.3
// Source: example.proto

package example

import (
	context "context"
	assert "github.com/chronos-tachyon/assert"
	vsrpc "github.com/chronos-tachyon/vsrpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	vsrpcMethodName_ExampleApi_ZeroInZeroOut vsrpc.Method = "vsrpc.ExampleApi.ZeroInZeroOut"
	vsrpcMethodName_ExampleApi_ZeroInOneOut  vsrpc.Method = "vsrpc.ExampleApi.ZeroInOneOut"
	vsrpcMethodName_ExampleApi_ZeroInManyOut vsrpc.Method = "vsrpc.ExampleApi.ZeroInManyOut"
	vsrpcMethodName_ExampleApi_OneInZeroOut  vsrpc.Method = "vsrpc.ExampleApi.OneInZeroOut"
	vsrpcMethodName_ExampleApi_OneInOneOut   vsrpc.Method = "vsrpc.ExampleApi.OneInOneOut"
	vsrpcMethodName_ExampleApi_OneInManyOut  vsrpc.Method = "vsrpc.ExampleApi.OneInManyOut"
	vsrpcMethodName_ExampleApi_ManyInZeroOut vsrpc.Method = "vsrpc.ExampleApi.ManyInZeroOut"
	vsrpcMethodName_ExampleApi_ManyInOneOut  vsrpc.Method = "vsrpc.ExampleApi.ManyInOneOut"
	vsrpcMethodName_ExampleApi_ManyInManyOut vsrpc.Method = "vsrpc.ExampleApi.ManyInManyOut"
)

// ExampleApiClient is the client API for ExampleApi service.
type ExampleApiClient interface {
	ZeroInZeroOut(ctx context.Context, options ...vsrpc.Option) error
	ZeroInOneOut(ctx context.Context, resp *ExampleResponse, options ...vsrpc.Option) error
	ZeroInManyOut(ctx context.Context, fn func(stream vsrpc.RecvStream[*ExampleResponse]) error, options ...vsrpc.Option) error
	OneInZeroOut(ctx context.Context, req *ExampleRequest, options ...vsrpc.Option) error
	OneInOneOut(ctx context.Context, req *ExampleRequest, resp *ExampleResponse, options ...vsrpc.Option) error
	OneInManyOut(ctx context.Context, req *ExampleRequest, fn func(stream vsrpc.RecvStream[*ExampleResponse]) error, options ...vsrpc.Option) error
	ManyInZeroOut(ctx context.Context, fn func(stream vsrpc.SendStream[*ExampleRequest]) error, options ...vsrpc.Option) error
	ManyInOneOut(ctx context.Context, resp *ExampleResponse, fn func(stream vsrpc.SendStream[*ExampleRequest]) error, options ...vsrpc.Option) error
	ManyInManyOut(ctx context.Context, fn func(stream vsrpc.BiStream[*ExampleRequest, *ExampleResponse]) error, options ...vsrpc.Option) error
}

func NewExampleApiClient(conn *vsrpc.Conn) ExampleApiClient {
	return vsrpcClientImpl_ExampleApi{conn: conn}
}

type vsrpcClientImpl_ExampleApi struct {
	conn *vsrpc.Conn
}

func (client vsrpcClientImpl_ExampleApi) Conn() *vsrpc.Conn {
	return client.conn
}

func (client vsrpcClientImpl_ExampleApi) ZeroInZeroOut(ctx context.Context, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ZeroInZeroOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*emptypb.Empty, *emptypb.Empty](call)
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) ZeroInOneOut(ctx context.Context, resp *ExampleResponse, options ...vsrpc.Option) error {
	assert.NotNil(&resp)
	resp.Reset()

	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ZeroInOneOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*emptypb.Empty, *ExampleResponse](call)
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	_, _, err = stream.Recv(true, resp)
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) ZeroInManyOut(ctx context.Context, fn func(stream vsrpc.RecvStream[*ExampleResponse]) error, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ZeroInManyOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*emptypb.Empty, *ExampleResponse](call)
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	err = fn(stream)
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) OneInZeroOut(ctx context.Context, req *ExampleRequest, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_OneInZeroOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *emptypb.Empty](call)
	err = stream.Send(req)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) OneInOneOut(ctx context.Context, req *ExampleRequest, resp *ExampleResponse, options ...vsrpc.Option) error {
	assert.NotNil(&resp)
	resp.Reset()

	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_OneInOneOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *ExampleResponse](call)
	err = stream.Send(req)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	_, _, err = stream.Recv(true, resp)
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) OneInManyOut(ctx context.Context, req *ExampleRequest, fn func(stream vsrpc.RecvStream[*ExampleResponse]) error, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_OneInManyOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *ExampleResponse](call)
	err = stream.Send(req)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	err = fn(stream)
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) ManyInZeroOut(ctx context.Context, fn func(stream vsrpc.SendStream[*ExampleRequest]) error, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ManyInZeroOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *emptypb.Empty](call)
	err = fn(stream)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) ManyInOneOut(ctx context.Context, resp *ExampleResponse, fn func(stream vsrpc.SendStream[*ExampleRequest]) error, options ...vsrpc.Option) error {
	assert.NotNil(&resp)
	resp.Reset()

	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ManyInOneOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *ExampleResponse](call)
	err = fn(stream)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	_, _, err = stream.Recv(true, resp)
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (client vsrpcClientImpl_ExampleApi) ManyInManyOut(ctx context.Context, fn func(stream vsrpc.BiStream[*ExampleRequest, *ExampleResponse]) error, options ...vsrpc.Option) error {
	call, err := client.Conn().Begin(ctx, vsrpcMethodName_ExampleApi_ManyInManyOut, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := vsrpc.NewBiStream[*ExampleRequest, *ExampleResponse](call)
	err = fn(stream)
	if err != nil {
		return err
	}
	err = stream.CloseSend()
	if err != nil {
		return err
	}
	return call.Wait().AsError()
}

var _ ExampleApiClient = (*vsrpcClientImpl_ExampleApi)(nil)

// ExampleApiServer is the server API for ExampleApi service.
type ExampleApiServer interface {
	ZeroInZeroOut(ctx context.Context) error
	ZeroInOneOut(ctx context.Context, resp *ExampleResponse) error
	ZeroInManyOut(ctx context.Context, stream vsrpc.SendStream[*ExampleResponse]) error
	OneInZeroOut(ctx context.Context, req *ExampleRequest) error
	OneInOneOut(ctx context.Context, req *ExampleRequest, resp *ExampleResponse) error
	OneInManyOut(ctx context.Context, req *ExampleRequest, stream vsrpc.SendStream[*ExampleResponse]) error
	ManyInZeroOut(ctx context.Context, stream vsrpc.RecvStream[*ExampleRequest]) error
	ManyInOneOut(ctx context.Context, resp *ExampleResponse, stream vsrpc.RecvStream[*ExampleRequest]) error
	ManyInManyOut(ctx context.Context, stream vsrpc.BiStream[*ExampleResponse, *ExampleRequest]) error
}

func NewExampleApiHandler(impl ExampleApiServer) vsrpc.Handler {
	return vsrpcHandler_ExampleApi{impl: impl}
}

type vsrpcHandler_ExampleApi struct {
	impl ExampleApiServer
}

func (h vsrpcHandler_ExampleApi) Handle(call *vsrpc.Call) error {
	ctx := call.Context()
	method := call.Method()

	if h.impl == nil {
		return vsrpc.NoSuchMethodError{Method: method}
	}

	switch method {
	case vsrpcMethodName_ExampleApi_ZeroInZeroOut:
		if err := h.impl.ZeroInZeroOut(ctx); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_ZeroInOneOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *emptypb.Empty](call)
		var resp ExampleResponse
		if err := h.impl.ZeroInOneOut(ctx, &resp); err != nil {
			return err
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_ZeroInManyOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *emptypb.Empty](call)
		if err := h.impl.ZeroInManyOut(ctx, stream); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_OneInZeroOut:
		stream := vsrpc.NewBiStream[*emptypb.Empty, *ExampleRequest](call)
		var req ExampleRequest
		if _, _, err := stream.Recv(true, &req); err != nil {
			return err
		}
		if err := h.impl.OneInZeroOut(ctx, &req); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_OneInOneOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *ExampleRequest](call)
		var req ExampleRequest
		if _, _, err := stream.Recv(true, &req); err != nil {
			return err
		}
		var resp ExampleResponse
		if err := h.impl.OneInOneOut(ctx, &req, &resp); err != nil {
			return err
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_OneInManyOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *ExampleRequest](call)
		var req ExampleRequest
		if _, _, err := stream.Recv(true, &req); err != nil {
			return err
		}
		if err := h.impl.OneInManyOut(ctx, &req, stream); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_ManyInZeroOut:
		stream := vsrpc.NewBiStream[*emptypb.Empty, *ExampleRequest](call)
		if err := h.impl.ManyInZeroOut(ctx, stream); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_ManyInOneOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *ExampleRequest](call)
		var resp ExampleResponse
		if err := h.impl.ManyInOneOut(ctx, &resp, stream); err != nil {
			return err
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}

	case vsrpcMethodName_ExampleApi_ManyInManyOut:
		stream := vsrpc.NewBiStream[*ExampleResponse, *ExampleRequest](call)
		if err := h.impl.ManyInManyOut(ctx, stream); err != nil {
			return err
		}

	default:
		return vsrpc.NoSuchMethodError{Method: method}
	}
	return nil
}

var _ vsrpc.Handler = vsrpcHandler_ExampleApi{}