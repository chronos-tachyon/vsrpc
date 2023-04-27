package vsrpc

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type ServerCallObserver interface {
	OnRequest(call *ServerCall, payload *anypb.Any)
	OnResponse(call *ServerCall, payload *anypb.Any)
	OnHalfClose(call *ServerCall)
	OnCancel(call *ServerCall)
	OnEnd(call *ServerCall, status *Status)
}

type ServerCall struct {
	ctx    context.Context
	cancel func()
	sc     *ServerConn
	id     uint32
	method string

	mu        sync.Mutex
	cv1       *sync.Cond
	cv2       *sync.Cond
	status    *Status
	observers map[ServerCallObserver]void
	queue     []*anypb.Any
	state     State
}

func newServerCall(sc *ServerConn, id uint32, method string) *ServerCall {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	call := &ServerCall{ctx: ctx, cancel: cancel, sc: sc, id: id, method: method}
	call.cv1 = sync.NewCond(&call.mu)
	call.cv2 = sync.NewCond(&call.mu)
	return call
}

func (call *ServerCall) isValid() bool {
	return call != nil && call.sc != nil
}

func (call *ServerCall) Context() context.Context {
	if !call.isValid() {
		return nil
	}
	return call.ctx
}

func (call *ServerCall) Conn() *ServerConn {
	if !call.isValid() {
		return nil
	}
	return call.sc
}

func (call *ServerCall) ID() uint32 {
	if !call.isValid() {
		return 0
	}
	return call.id
}

func (call *ServerCall) Method() string {
	if !call.isValid() {
		return ""
	}
	return call.method
}

func (call *ServerCall) AddObserver(o ServerCallObserver) {
	if o == nil || !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	for _, payload := range call.queue {
		o.OnRequest(call, payload)
	}

	switch {
	case call.state >= StateClosed:
		o.OnEnd(call, call.status)
		return

	case call.state >= StateGoingAway:
		o.OnCancel(call)

	case call.state >= StateShuttingDown:
		o.OnHalfClose(call)
	}

	if call.observers == nil {
		call.observers = make(map[ServerCallObserver]void, 16)
	}
	call.observers[o] = void{}
}

func (call *ServerCall) RemoveObserver(o ServerCallObserver) {
	if o == nil || !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()
	if call.observers != nil {
		delete(call.observers, o)
	}
}

func (call *ServerCall) WaitRecv(min uint) {
	if !call.isValid() {
		return
	}

	call.mu.Lock()
	for call.state < StateShuttingDown && uint(len(call.queue)) < min {
		call.cv1.Wait()
	}
	call.mu.Unlock()
}

func (call *ServerCall) Recv() (queue []*anypb.Any, done bool) {
	if call.isValid() {
		call.mu.Lock()
		queue = call.queue
		done = (call.state >= StateShuttingDown)
		call.queue = nil
		call.mu.Unlock()
	} else {
		done = true
	}
	return
}

func (call *ServerCall) Peek() (count uint, done bool) {
	if call.isValid() {
		call.mu.Lock()
		count = uint(len(call.queue))
		done = (call.state >= StateShuttingDown)
		call.mu.Unlock()
	} else {
		done = true
	}
	return
}

func (call *ServerCall) SendMessage(payload proto.Message) error {
	payloadAny, err := asAny(payload)
	if err != nil {
		return err
	}
	return call.Send(payloadAny)
}

func (call *ServerCall) Send(payload *anypb.Any) error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	sc := call.sc
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateClosed {
		return call.lockedAbort(ErrClosed)
	}

	err := WriteResponse(call.ctx, sc.pc, call.id, payload)
	if err != nil {
		return sc.lockedGotWriteError(err)
	}

	for o := range call.observers {
		neverPanic(func() { o.OnResponse(call, payload) })
	}

	return nil
}

func (call *ServerCall) End(status *Status) error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	sc := call.sc
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateClosed {
		return call.lockedAbort(ErrClosed)
	}

	if err := WriteEnd(call.ctx, sc.pc, call.id, status); err != nil {
		return sc.lockedGotWriteError(err)
	}

	return call.lockedEnd(status, nil)
}

func (call *ServerCall) Wait() *Status {
	if !call.isValid() {
		return &Status{Code: Status_CANCELLED}
	}

	call.mu.Lock()
	for call.state < StateClosed {
		call.cv2.Wait()
	}
	status := call.status
	call.mu.Unlock()
	return status
}

func (call *ServerCall) Close() error {
	call.gotCancel()
	return call.Wait().AsError()
}

func (call *ServerCall) gotRequest(payload *anypb.Any) bool {
	if !call.isValid() {
		return false
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return false
	}

	if call.state >= StateShuttingDown {
		call.sc.gotReadError(ProtocolViolationError{
			Err: fmt.Errorf("REQUEST after HALF_CLOSE or CANCEL"),
		})
		return true
	}

	call.queue = append(call.queue, payload)
	call.cv1.Signal()
	return false
}

func (call *ServerCall) gotHalfClose() bool {
	if !call.isValid() {
		return false
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateShuttingDown {
		return false
	}

	call.state = StateShuttingDown
	call.cv1.Broadcast()
	return false
}

func (call *ServerCall) gotCancel() bool {
	if !call.isValid() {
		return false
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateGoingAway {
		return false
	}

	call.cancel()

	call.state = StateGoingAway
	call.cv1.Broadcast()
	return false
}

func (call *ServerCall) abort(err error) error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	return call.lockedAbort(err)
}

func (call *ServerCall) lockedAbort(err error) error {
	status := Abort(err)
	return call.lockedEnd(status, err)
}

func (call *ServerCall) lockedEnd(status *Status, err error) error {
	if call.state < StateGoingAway {
		call.cancel()
	}

	if status == nil {
		status = &Status{Code: Status_OK}
	}

	for o := range call.observers {
		neverPanic(func() { o.OnEnd(call, status) })
	}

	call.state = StateClosed
	call.status = status
	call.cv1.Broadcast()
	call.cv2.Broadcast()
	go call.sc.forgetCall(call)
	return err
}

type ServerCallNoOpObserver struct{}

func (ServerCallNoOpObserver) OnRequest(call *ServerCall, payload *anypb.Any)  {}
func (ServerCallNoOpObserver) OnResponse(call *ServerCall, payload *anypb.Any) {}
func (ServerCallNoOpObserver) OnHalfClose(call *ServerCall)                    {}
func (ServerCallNoOpObserver) OnCancel(call *ServerCall)                       {}
func (ServerCallNoOpObserver) OnEnd(call *ServerCall, status *Status)          {}

var _ ServerCallObserver = ServerCallNoOpObserver{}

type ServerCallFuncObserver struct {
	Request   func(call *ServerCall, payload *anypb.Any)
	Response  func(call *ServerCall, payload *anypb.Any)
	HalfClose func(call *ServerCall)
	Cancel    func(call *ServerCall)
	End       func(call *ServerCall, status *Status)
}

func (o ServerCallFuncObserver) OnRequest(call *ServerCall, payload *anypb.Any) {
	if o.Request != nil {
		o.Request(call, payload)
	}
}

func (o ServerCallFuncObserver) OnResponse(call *ServerCall, payload *anypb.Any) {
	if o.Response != nil {
		o.Response(call, payload)
	}
}

func (o ServerCallFuncObserver) OnHalfClose(call *ServerCall) {
	if o.HalfClose != nil {
		o.HalfClose(call)
	}
}

func (o ServerCallFuncObserver) OnCancel(call *ServerCall) {
	if o.Cancel != nil {
		o.Cancel(call)
	}
}

func (o ServerCallFuncObserver) OnEnd(call *ServerCall, status *Status) {
	if o.End != nil {
		o.End(call, status)
	}
}

var _ ServerCallObserver = ServerCallFuncObserver{}
