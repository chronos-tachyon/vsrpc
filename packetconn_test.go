package vsrpc

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"
)

const (
	FooServer_AlwaysOK  Method = "foo.AlwaysOK"
	FooServer_Sum       Method = "foo.Sum"
	FooServer_Missing   Method = "foo.Missing"
	FooServer_Forbidden Method = "bar.Forbidden"
)

type FooClient interface {
	AlwaysOK(ctx context.Context, options ...Option) error
	Sum(ctx context.Context, fn func(stream BiStream[*SumRequest, *SumResponse]) error, options ...Option) error
	Missing(ctx context.Context, options ...Option) error
	Forbidden(ctx context.Context, options ...Option) error
}

type FooClientImpl struct {
	Conn *Conn
}

func (c FooClientImpl) AlwaysOK(ctx context.Context, options ...Option) error {
	call, err := c.Conn.Begin(ctx, FooServer_AlwaysOK, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func (c FooClientImpl) Sum(ctx context.Context, fn func(stream BiStream[*SumRequest, *SumResponse]) error, options ...Option) error {
	call, err := c.Conn.Begin(ctx, FooServer_Sum, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	stream := NewBiStream[*SumRequest, *SumResponse](call)
	err = fn(stream)
	if err != nil {
		return err
	}

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func (c FooClientImpl) Missing(ctx context.Context, options ...Option) error {
	call, err := c.Conn.Begin(ctx, FooServer_Missing, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func (c FooClientImpl) Forbidden(ctx context.Context, options ...Option) error {
	call, err := c.Conn.Begin(ctx, FooServer_Forbidden, options...)
	if err != nil {
		return err
	}
	defer func() { _ = call.Close() }()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

var _ FooClient = FooClientImpl{}

type FooServer interface {
	AlwaysOK(ctx context.Context) error
	Sum(ctx context.Context, stream BiStream[*SumResponse, *SumRequest]) error
}

type FooHandler struct {
	Impl FooServer
}

func (h FooHandler) Handle(call *Call) error {
	ctx := call.Context()
	method := call.Method()

	if h.Impl == nil {
		return NoSuchMethodError{Method: method}
	}

	switch method {
	case FooServer_AlwaysOK:
		if err := h.Impl.AlwaysOK(ctx); err != nil {
			return err
		}

	case FooServer_Sum:
		stream := NewBiStream[*SumResponse, *SumRequest](call)
		if err := h.Impl.Sum(ctx, stream); err != nil {
			return err
		}

	default:
		return NoSuchMethodError{Method: method}
	}
	return nil
}

var _ Handler = FooHandler{}

type ForbiddenHandler struct{}

func (h ForbiddenHandler) Handle(call *Call) error {
	status := &Status{
		Code: Status_PERMISSION_DENIED,
		Text: fmt.Sprintf("permission denied for method %q", call.Method()),
	}
	return status.AsError()
}

var _ Handler = ForbiddenHandler{}

type FooServerImpl struct{}

func (FooServerImpl) AlwaysOK(ctx context.Context) error {
	return nil
}

func (FooServerImpl) Sum(ctx context.Context, stream BiStream[*SumResponse, *SumRequest]) error {
	var req SumRequest
	var resp SumResponse
	var err error
	done := false
	for err == nil && !done {
		var ok bool
		ok, done, err = stream.Recv(true, &req)
		if err == nil && ok {
			resp.Reset()
			for _, i32 := range req.Input {
				resp.Output += i32
			}
			err = stream.Send(&resp)
		}
	}
	return err
}

var _ FooServer = FooServerImpl{}

type Case struct {
	Name       string
	Func       func(ctx context.Context, t *testing.T, c FooClient) error
	Timeout    time.Duration
	HasTimeout bool
}

var Cases = []Case{
	{"AlwaysOK", CaseAlwaysOK, 0, false},
	{"SumNone", CaseSumNone, 0, false},
	{"SumOne", CaseSumOne, 0, false},
	{"SumThree", CaseSumThree, 0, false},
	{"Missing", CaseMissing, 0, false},
	{"Forbidden", CaseForbidden, 0, false},
}

func CaseAlwaysOK(ctx context.Context, t *testing.T, c FooClient) error {
	return c.AlwaysOK(ctx)
}

func CaseSumNone(ctx context.Context, t *testing.T, c FooClient) error {
	return c.Sum(ctx, func(stream BiStream[*SumRequest, *SumResponse]) error {
		return nil
	})
}

func CaseSumOne(ctx context.Context, t *testing.T, c FooClient) error {
	return c.Sum(ctx, func(stream BiStream[*SumRequest, *SumResponse]) error {
		var req SumRequest
		var resp SumResponse

		req.Input = []int32{1, 2, 3, 4, 5}
		err := stream.Send(&req)
		if err != nil {
			return err
		}

		ok, _, err := stream.Recv(true, &resp)
		if err != nil {
			return err
		}
		if !ok || resp.Output != 15 {
			return fmt.Errorf("expected ok = true && output = 15, got ok = %t && output = %d", ok, resp.Output)
		}

		return nil
	})
}

func CaseSumThree(ctx context.Context, t *testing.T, c FooClient) error {
	return c.Sum(ctx, func(stream BiStream[*SumRequest, *SumResponse]) error {
		var req SumRequest
		var resp SumResponse

		req.Input = []int32{1, 2, 3, 4, 5}
		err := stream.Send(&req)
		if err != nil {
			return err
		}

		ok, _, err := stream.Recv(true, &resp)
		if err != nil {
			return err
		}
		if !ok || resp.Output != 15 {
			return fmt.Errorf("expected ok = true && output = 15, got ok = %t && output = %d", ok, resp.Output)
		}

		req.Input = []int32{2, 3}
		err = stream.Send(&req)
		if err != nil {
			return err
		}

		ok, _, err = stream.Recv(true, &resp)
		if err != nil {
			return err
		}
		if !ok || resp.Output != 5 {
			return fmt.Errorf("expected ok = true && output = 5, got ok = %t && output = %d", ok, resp.Output)
		}

		req.Input = []int32{}
		err = stream.Send(&req)
		if err != nil {
			return err
		}

		ok, _, err = stream.Recv(true, &resp)
		if err != nil {
			return err
		}
		if !ok || resp.Output != 0 {
			return fmt.Errorf("expected ok = true && output = 0, got ok = %t && output = %d", ok, resp.Output)
		}

		return nil
	})
}

func CaseMissing(ctx context.Context, t *testing.T, c FooClient) error {
	status := StatusFromError(c.Missing(ctx))
	if status == nil {
		status = &Status{Code: Status_OK}
	}
	if expect := Status_UNIMPLEMENTED; status.Code != expect {
		return fmt.Errorf("wrong status code: expected %v, got %v", expect, status.Code)
	}
	if expect := fmt.Sprintf("method %q is not implemented", FooServer_Missing); status.Text != expect {
		return fmt.Errorf("wrong status text: expected %q, got %q", expect, status.Text)
	}
	return nil
}

func CaseForbidden(ctx context.Context, t *testing.T, c FooClient) error {
	status := StatusFromError(c.Forbidden(ctx))
	if status == nil {
		status = &Status{Code: Status_OK}
	}
	if expect := Status_PERMISSION_DENIED; status.Code != expect {
		return fmt.Errorf("wrong status code: expected %v, got %v", expect, status.Code)
	}
	if expect := fmt.Sprintf("permission denied for method %q", FooServer_Forbidden); status.Text != expect {
		return fmt.Errorf("wrong status text: expected %q, got %q", expect, status.Text)
	}
	return nil
}

func NewTestMux() *HandlerMux {
	var h1 Handler = FooHandler{Impl: FooServerImpl{}}
	var h2 Handler = ForbiddenHandler{}
	mux := &HandlerMux{}
	mux.Add(h1, "foo.*")
	mux.Add(h2, "bar.*")
	return mux
}

func ContextFromTest(t *testing.T) (ctx context.Context, cancel context.CancelFunc) {
	ctx = context.Background()
	cancel = func() {}
	if deadline, ok := t.Deadline(); ok {
		ctx, cancel = context.WithDeadline(ctx, deadline)
	}
	return
}

func Run(ctx context.Context, t *testing.T, c FooClient, cases []Case) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ctxInner := ctx
			if tc.HasTimeout {
				dctx, cancel := context.WithTimeout(ctx, tc.Timeout)
				defer cancel()
				ctxInner = dctx
			}

			if err := tc.Func(ctxInner, t, c); err != nil {
				t.Error(err.Error())
			}
		})
	}
}

func TestUnixPacket(t *testing.T) {
	ctx, cancel := ContextFromTest(t)
	defer cancel()

	addr, err := net.ResolveUnixAddr("unixpacket", "")
	if err != nil {
		panic(err)
	}

	var pd UnixPacketDialer

	pl, err := pd.ListenPacket(ctx, addr)
	if err != nil {
		panic(err)
	}

	s := NewServer(pl, NewTestMux())
	defer s.Close()

	c := NewClient(&pd)
	defer c.Close()

	addr = s.Addr().(*net.UnixAddr)
	conn, err := c.Dial(ctx, addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	Run(ctx, t, FooClientImpl{Conn: conn}, Cases)
}
