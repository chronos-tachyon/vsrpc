package vsrpc

import (
	"context"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type ClientCallObserver interface {
	OnRequest(call *ClientCall, payload *anypb.Any)
	OnResponse(call *ClientCall, payload *anypb.Any)
	OnHalfClose(call *ClientCall)
	OnCancel(call *ClientCall)
	OnEnd(call *ClientCall, status *Status)
}

type ClientCall struct {
	ctx    context.Context
	cc     *ClientConn
	id     uint32
	method string

	mu        sync.Mutex
	cv1       *sync.Cond
	cv2       *sync.Cond
	status    *Status
	observers map[ClientCallObserver]void
	queue     []*anypb.Any
	state     State
}

func newClientCall(ctx context.Context, cc *ClientConn, id uint32, method string) *ClientCall {
	call := &ClientCall{ctx: ctx, cc: cc, id: id, method: method}
	call.cv1 = sync.NewCond(&call.mu)
	call.cv2 = sync.NewCond(&call.mu)
	return call
}

func (call *ClientCall) isValid() bool {
	return call != nil && call.ctx != nil && call.cc != nil
}

func (call *ClientCall) Context() context.Context {
	if !call.isValid() {
		return context.Background()
	}
	return call.ctx
}

func (call *ClientCall) Conn() *ClientConn {
	if !call.isValid() {
		return nil
	}
	return call.cc
}

func (call *ClientCall) ID() uint32 {
	if !call.isValid() {
		return 0
	}
	return call.id
}

func (call *ClientCall) Method() string {
	if !call.isValid() {
		return ""
	}
	return call.method
}

func (call *ClientCall) AddObserver(o ClientCallObserver) {
	if o == nil || !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	for _, payload := range call.queue {
		o.OnResponse(call, payload)
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
		call.observers = make(map[ClientCallObserver]void, 16)
	}
	call.observers[o] = void{}
}

func (call *ClientCall) RemoveObserver(o ClientCallObserver) {
	if o == nil || !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()
	if call.observers != nil {
		delete(call.observers, o)
	}
}

func (call *ClientCall) WaitRecv(min uint) {
	if !call.isValid() {
		return
	}

	call.mu.Lock()
	for call.state < StateClosed && uint(len(call.queue)) < min {
		call.cv1.Wait()
	}
	call.mu.Unlock()
}

func (call *ClientCall) Recv() (queue []*anypb.Any, done bool) {
	if call.isValid() {
		call.mu.Lock()
		queue = call.queue
		done = (call.state >= StateClosed)
		call.queue = nil
		call.mu.Unlock()
	} else {
		done = true
	}
	return
}

func (call *ClientCall) Peek() (count uint, done bool) {
	if call.isValid() {
		call.mu.Lock()
		count = uint(len(call.queue))
		done = (call.state >= StateClosed)
		call.mu.Unlock()
	} else {
		done = true
	}
	return
}

func (call *ClientCall) SendMessage(payload proto.Message) error {
	payloadAny, err := asAny(payload)
	if err != nil {
		return err
	}
	return call.Send(payloadAny)
}

func (call *ClientCall) Send(payload *anypb.Any) error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	if call.state >= StateShuttingDown {
		return ErrHalfClosed
	}

	cc := call.cc
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return call.lockedAbort(ErrClosed)
	}

	err := WriteRequest(call.ctx, cc.pc, call.id, payload)
	if err != nil {
		return cc.lockedGotWriteError(err)
	}

	for o := range call.observers {
		neverPanic(func() { o.OnRequest(call, payload) })
	}

	return nil
}

func (call *ClientCall) CloseSend() error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	if call.state >= StateShuttingDown {
		return nil
	}

	cc := call.cc
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return call.lockedAbort(ErrClosed)
	}

	for o := range call.observers {
		neverPanic(func() { o.OnHalfClose(call) })
	}

	err := WriteHalfClose(call.ctx, cc.pc, call.id)
	if err != nil {
		return cc.lockedGotWriteError(err)
	}

	call.state = StateShuttingDown
	return nil
}

func (call *ClientCall) Cancel() error {
	if !call.isValid() {
		return ErrClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return ErrClosed
	}

	if call.state >= StateGoingAway {
		return nil
	}

	cc := call.cc
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return call.lockedAbort(ErrClosed)
	}

	for o := range call.observers {
		neverPanic(func() { o.OnCancel(call) })
	}

	err := WriteCancel(call.ctx, cc.pc, call.id)
	if err != nil {
		return cc.lockedGotWriteError(err)
	}

	call.state = StateGoingAway
	return nil
}

func (call *ClientCall) Wait() *Status {
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

func (call *ClientCall) Close() error {
	if err := call.Cancel(); err != nil {
		return err
	}
	return call.Wait().AsError()
}

func (call *ClientCall) gotResponse(payload *anypb.Any) {
	if !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= StateClosed {
		return
	}

	for o := range call.observers {
		neverPanic(func() { o.OnResponse(call, payload) })
	}

	call.queue = append(call.queue, payload)
	call.cv1.Signal()
}

func (call *ClientCall) gotEnd(status *Status) {
	if !call.isValid() {
		return
	}

	call.mu.Lock()
	defer call.mu.Unlock()
	call.lockedEnd(status)
}

func (call *ClientCall) lockedAbort(err error) error {
	if call.state < StateClosed {
		call.lockedEnd(Abort(err))
	}
	return err
}

func (call *ClientCall) lockedEnd(status *Status) {
	if call.state >= StateClosed {
		return
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
	go call.cc.forgetCall(call)
}

type ClientCallNoOpObserver struct{}

func (ClientCallNoOpObserver) OnRequest(call *ClientCall, payload *anypb.Any)  {}
func (ClientCallNoOpObserver) OnResponse(call *ClientCall, payload *anypb.Any) {}
func (ClientCallNoOpObserver) OnHalfClose(call *ClientCall)                    {}
func (ClientCallNoOpObserver) OnCancel(call *ClientCall)                       {}
func (ClientCallNoOpObserver) OnEnd(call *ClientCall, status *Status)          {}

var _ ClientCallObserver = ClientCallNoOpObserver{}

type ClientCallFuncObserver struct {
	Request   func(call *ClientCall, payload *anypb.Any)
	Response  func(call *ClientCall, payload *anypb.Any)
	HalfClose func(call *ClientCall)
	Cancel    func(call *ClientCall)
	End       func(call *ClientCall, status *Status)
}

func (o ClientCallFuncObserver) OnRequest(call *ClientCall, payload *anypb.Any) {
	if o.Request != nil {
		o.Request(call, payload)
	}
}

func (o ClientCallFuncObserver) OnResponse(call *ClientCall, payload *anypb.Any) {
	if o.Response != nil {
		o.Response(call, payload)
	}
}

func (o ClientCallFuncObserver) OnHalfClose(call *ClientCall) {
	if o.HalfClose != nil {
		o.HalfClose(call)
	}
}

func (o ClientCallFuncObserver) OnCancel(call *ClientCall) {
	if o.Cancel != nil {
		o.Cancel(call)
	}
}

func (o ClientCallFuncObserver) OnEnd(call *ClientCall, status *Status) {
	if o.End != nil {
		o.End(call, status)
	}
}

var _ ClientCallObserver = ClientCallFuncObserver{}
