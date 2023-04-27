package vsrpc

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

type Case struct {
	Name       string
	Func       func(ctx context.Context, t *testing.T, cc *ClientConn) error
	Timeout    time.Duration
	HasTimeout bool
}

var Cases = []Case{
	{"AlwaysOK", CaseAlwaysOK, 0, false},
	{"SumNone", CaseSumNone, 0, false},
	{"SumOne", CaseSumOne, 0, false},
	{"SumThree", CaseSumThree, 0, false},
	{"Unexpected", CaseUnexpected, 0, false},
	{"Bogus", CaseBogus, 0, false},
}

func MustUnmarshal(out proto.Message, in *anypb.Any) {
	if err := UnmarshalFromAny(out, in); err != nil {
		panic(err)
	}
}

func CaseAlwaysOK(ctx context.Context, t *testing.T, cc *ClientConn) error {
	call, err := cc.Call(ctx, "foo.AlwaysOK")
	if err != nil {
		return err
	}
	defer call.Close()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func CaseSumNone(ctx context.Context, t *testing.T, cc *ClientConn) error {
	call, err := cc.Call(ctx, "foo.Sum")
	if err != nil {
		return err
	}
	defer call.Close()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func CaseSumOne(ctx context.Context, t *testing.T, cc *ClientConn) error {
	var req SumRequest
	var resp SumResponse

	call, err := cc.Call(ctx, "foo.Sum")
	if err != nil {
		return err
	}
	defer call.Close()

	req.Input = []int32{1, 2, 3, 4, 5}
	err = call.SendMessage(&req)
	if err != nil {
		return err
	}

	call.WaitRecv(1)
	queue, _ := call.Recv()

	if len(queue) != 1 {
		return fmt.Errorf("expected len(queue) = 1, got len(queue) = %d", len(queue))
	}

	MustUnmarshal(&resp, queue[0])
	if resp.Output != 15 {
		return fmt.Errorf("expected output = 15, got output = %d", resp.Output)
	}

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func CaseSumThree(ctx context.Context, t *testing.T, cc *ClientConn) error {
	var req SumRequest
	var resp SumResponse

	call, err := cc.Call(ctx, "foo.Sum")
	if err != nil {
		return err
	}
	defer call.Close()

	req.Input = []int32{1, 2, 3, 4, 5}
	err = call.SendMessage(&req)
	if err != nil {
		return err
	}

	req.Input = []int32{2, 3}
	err = call.SendMessage(&req)
	if err != nil {
		return err
	}

	req.Input = []int32{}
	err = call.SendMessage(&req)
	if err != nil {
		return err
	}

	call.WaitRecv(3)
	queue, _ := call.Recv()

	if len(queue) != 3 {
		return fmt.Errorf("expected len(queue) = 3, got len(queue) = %d", len(queue))
	}

	MustUnmarshal(&resp, queue[0])
	if resp.Output != 15 {
		return fmt.Errorf("expected output = 15, got output = %d", resp.Output)
	}

	MustUnmarshal(&resp, queue[1])
	if resp.Output != 5 {
		return fmt.Errorf("expected output = 5, got output = %d", resp.Output)
	}

	MustUnmarshal(&resp, queue[2])
	if resp.Output != 0 {
		return fmt.Errorf("expected output = 0, got output = %d", resp.Output)
	}

	err = call.CloseSend()
	if err != nil {
		return err
	}

	return call.Wait().AsError()
}

func CaseUnexpected(ctx context.Context, t *testing.T, cc *ClientConn) error {
	call, err := cc.Call(ctx, "foo.Unexpected")
	if err != nil {
		return err
	}
	defer call.Close()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	status := call.Wait()
	if status == nil {
		status = &Status{Code: Status_OK}
	}
	if expect := Status_PERMISSION_DENIED; status.Code != expect {
		return fmt.Errorf("wrong status code: expected %v, got %v", expect, status.Code)
	}
	if expect := "foo.* is not permitted"; status.Text != expect {
		return fmt.Errorf("wrong status text: expected %q, got %q", expect, status.Text)
	}
	return nil
}

func CaseBogus(ctx context.Context, t *testing.T, cc *ClientConn) error {
	call, err := cc.Call(ctx, "bar.Bogus")
	if err != nil {
		return err
	}
	defer call.Close()

	err = call.CloseSend()
	if err != nil {
		return err
	}

	status := call.Wait()
	if status == nil {
		status = &Status{Code: Status_OK}
	}
	if expect := Status_UNIMPLEMENTED; status.Code != expect {
		return fmt.Errorf("wrong status code: expected %v, got %v", expect, status.Code)
	}
	if expect := "method \"bar.Bogus\" is not implemented"; status.Text != expect {
		return fmt.Errorf("wrong status text: expected %q, got %q", expect, status.Text)
	}
	return nil
}

func NewTestMux() *HandlerMux {
	mux := &HandlerMux{}
	mux.AddFunc("foo.AlwaysOK", func(call *ServerCall) error {
		return nil
	})
	mux.AddFunc("foo.Sum", func(call *ServerCall) error {
		var queue []*anypb.Any
		var done bool
		for !done {
			call.WaitRecv(1)
			queue, done = call.Recv()
			for _, reqAny := range queue {
				var req SumRequest
				if err := UnmarshalFromAny(&req, reqAny); err != nil {
					return err
				}

				var resp SumResponse
				for _, i32 := range req.Input {
					resp.Output += i32
				}

				_ = call.SendMessage(&resp)
			}
		}
		return nil
	})
	mux.AddFunc("foo.*", func(call *ServerCall) error {
		return StatusError{
			Status: &Status{
				Code: Status_PERMISSION_DENIED,
				Text: "foo.* is not permitted",
			},
		}
	})
	mux.AddFunc("*", func(call *ServerCall) error {
		method := call.Method()
		return NoSuchMethodError{Method: method}
	})
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

func Run(ctx context.Context, t *testing.T, cc *ClientConn, cases []Case) {
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			ctxInner := ctx
			if tc.HasTimeout {
				dctx, cancel := context.WithTimeout(ctx, tc.Timeout)
				defer cancel()
				ctxInner = dctx
			}

			if err := tc.Func(ctxInner, t, cc); err != nil {
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

	var pd NetworkPacketDialer

	pl, err := pd.ListenPacket(ctx, addr)
	if err != nil {
		panic(err)
	}
	addr = pl.Addr().(*net.UnixAddr)

	s := NewServer(pl, NewTestMux())
	defer s.Close()

	c := NewClient(&pd)
	defer c.Close()

	cc, err := c.Dial(ctx, addr)
	if err != nil {
		panic(err)
	}
	defer cc.Close()

	Run(ctx, t, cc, Cases)
}
