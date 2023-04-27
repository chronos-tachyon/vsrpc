package vsrpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type ClientConnObserver interface {
	OnCall(cc *ClientConn, call *ClientCall)
	OnShutdown(cc *ClientConn)
	OnGoAway(cc *ClientConn)
	OnReadError(cc *ClientConn, err error)
	OnWriteError(cc *ClientConn, err error)
	OnClose(cc *ClientConn, err error)
}

type ClientConn struct {
	pc PacketConn
	c  *Client

	mu        sync.Mutex
	calls     map[uint32]*ClientCall
	observers map[ClientConnObserver]void
	state     State
	id        uint32
}

func newClientConn(pc PacketConn, c *Client) *ClientConn {
	cc := &ClientConn{pc: pc, c: c}
	go cc.readThread()
	return cc
}

func (cc *ClientConn) isValid() bool {
	return cc != nil && cc.pc != nil && cc.c != nil
}

func (cc *ClientConn) Conn() PacketConn {
	if !cc.isValid() {
		return nil
	}
	return cc.pc
}

func (cc *ClientConn) Client() *Client {
	if !cc.isValid() {
		return nil
	}
	return cc.c
}

func (cc *ClientConn) AddObserver(o ClientConnObserver) {
	if o == nil || !cc.isValid() {
		return
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		o.OnClose(cc, nil)
		return
	}

	for _, call := range cc.calls {
		o.OnCall(cc, call)
	}

	switch {
	case cc.state >= StateGoingAway:
		o.OnGoAway(cc)
	case cc.state >= StateShuttingDown:
		o.OnShutdown(cc)
	}

	if cc.observers == nil {
		cc.observers = make(map[ClientConnObserver]void, 16)
	}
	cc.observers[o] = void{}
}

func (cc *ClientConn) RemoveObserver(o ClientConnObserver) {
	if o == nil || !cc.isValid() {
		return
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.observers != nil {
		delete(cc.observers, o)
	}
}

func (cc *ClientConn) Call(ctx context.Context, method string) (*ClientCall, error) {
	assert.NotNil(&ctx)

	if !cc.isValid() {
		return nil, ErrClosed
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return nil, ErrClosed
	}

	if cc.state >= StateGoingAway {
		return nil, ErrServerClosing
	}

	if cc.state >= StateShuttingDown {
		return nil, ErrClientClosing
	}

	id := cc.lockedNextID()
	err := WriteBegin(ctx, cc.pc, id, method)
	if err != nil {
		return nil, cc.lockedGotWriteError(err)
	}

	call := newClientCall(ctx, cc, id, method)
	for o := range cc.observers {
		neverPanic(func() { o.OnCall(cc, call) })
	}

	if cc.calls == nil {
		cc.calls = make(map[uint32]*ClientCall, 16)
	}
	cc.calls[id] = call
	return call, nil
}

func (cc *ClientConn) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if !cc.isValid() {
		return ErrClosed
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return ErrClosed
	}

	if cc.state >= StateShuttingDown {
		return nil
	}

	for o := range cc.observers {
		neverPanic(func() { o.OnShutdown(cc) })
	}

	err := WriteShutdown(ctx, cc.pc)
	if err != nil {
		return cc.lockedGotWriteError(err)
	}

	cc.state = StateShuttingDown
	return nil
}

func (cc *ClientConn) Close() error {
	if !cc.isValid() {
		return ErrClosed
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateClosed {
		return ErrClosed
	}

	return cc.lockedClose()
}

func (cc *ClientConn) forgetCall(call *ClientCall) {
	if call == nil || !cc.isValid() {
		return
	}

	cc.mu.Lock()
	defer cc.mu.Unlock()
	if cc.calls != nil {
		delete(cc.calls, call.id)
	}
}

func (cc *ClientConn) readThread() {
	ctx := context.Background()
	for {
		var frame Frame
		if err := ReadFrame(ctx, cc.pc, &frame); err != nil {
			cc.gotReadError(err)
			return
		}

		switch frame.Type {
		case Frame_NO_OP:
			// pass

		case Frame_GO_AWAY:
			cc.gotGoAway()

		case Frame_RESPONSE:
			if frame.CallId == 0 {
				cc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("RESPONSE with call_id 0"),
				})
				return
			}

			cc.findCall(frame.CallId).gotResponse(frame.Payload)

		case Frame_END:
			if frame.CallId == 0 {
				cc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("END with call_id 0"),
				})
				return
			}

			cc.findCall(frame.CallId).gotEnd(frame.Status)

		default:
			cc.gotReadError(ProtocolViolationError{
				Err: UnexpectedFrameTypeError{
					Actual: frame.Type,
					Expected: []Frame_Type{
						Frame_GO_AWAY,
						Frame_RESPONSE,
						Frame_END,
					},
				},
			})
			return
		}
	}
}

func (cc *ClientConn) gotReadError(err error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.lockedGotReadError(err)
}

func (cc *ClientConn) gotGoAway() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if cc.state >= StateGoingAway {
		return
	}

	for o := range cc.observers {
		neverPanic(func() { o.OnGoAway(cc) })
	}

	cc.state = StateGoingAway
}

func (cc *ClientConn) findCall(id uint32) *ClientCall {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.calls[id]
}

func (cc *ClientConn) lockedNextID() uint32 {
	id := cc.id

	if n := uint(len(cc.calls)); n == 0 || id == 0 || uint(id) >= 2*n {
		id = 1
	}

	for {
		if _, found := cc.calls[id]; !found {
			cc.id = id
			return id
		}
		id++
	}
}

func (cc *ClientConn) lockedGotReadError(err error) {
	if err == nil || cc.state >= StateClosed {
		return
	}

	for o := range cc.observers {
		neverPanic(func() { o.OnReadError(cc, err) })
	}

	if !IsRecoverable(err) {
		_ = cc.lockedClose()
	}
}

func (cc *ClientConn) lockedGotWriteError(err error) error {
	if err == nil || cc.state >= StateClosed {
		return err
	}

	for o := range cc.observers {
		neverPanic(func() { o.OnWriteError(cc, err) })
	}

	if !IsRecoverable(err) {
		_ = cc.lockedClose()
	}
	return err
}

func (cc *ClientConn) lockedClose() error {
	err := try(cc.pc.Close)

	for o := range cc.observers {
		neverPanic(func() { o.OnClose(cc, err) })
	}

	for _, call := range cc.calls {
		call.gotEnd(Abort(ErrClosed))
	}

	cc.state = StateClosed
	cc.calls = nil
	cc.observers = nil
	go cc.c.forgetConn(cc)
	return err
}

type ClientConnNoOpObserver struct{}

func (ClientConnNoOpObserver) OnCall(cc *ClientConn, call *ClientCall) {}
func (ClientConnNoOpObserver) OnShutdown(cc *ClientConn)               {}
func (ClientConnNoOpObserver) OnGoAway(cc *ClientConn)                 {}
func (ClientConnNoOpObserver) OnReadError(cc *ClientConn, err error)   {}
func (ClientConnNoOpObserver) OnWriteError(cc *ClientConn, err error)  {}
func (ClientConnNoOpObserver) OnClose(cc *ClientConn, err error)       {}

var _ ClientConnObserver = ClientConnNoOpObserver{}

type ClientConnFuncObserver struct {
	Call         func(cc *ClientConn, call *ClientCall)
	ShuttingDown func(cc *ClientConn)
	GoingAway    func(cc *ClientConn)
	ReadError    func(cc *ClientConn, err error)
	WriteError   func(cc *ClientConn, err error)
	Close        func(cc *ClientConn, err error)
}

func (o ClientConnFuncObserver) OnCall(cc *ClientConn, call *ClientCall) {
	if o.Call != nil {
		o.Call(cc, call)
	}
}

func (o ClientConnFuncObserver) OnShutdown(cc *ClientConn) {
	if o.ShuttingDown != nil {
		o.ShuttingDown(cc)
	}
}

func (o ClientConnFuncObserver) OnGoAway(cc *ClientConn) {
	if o.GoingAway != nil {
		o.GoingAway(cc)
	}
}

func (o ClientConnFuncObserver) OnReadError(cc *ClientConn, err error) {
	if o.ReadError != nil {
		o.ReadError(cc, err)
	}
}

func (o ClientConnFuncObserver) OnWriteError(cc *ClientConn, err error) {
	if o.WriteError != nil {
		o.WriteError(cc, err)
	}
}

func (o ClientConnFuncObserver) OnClose(cc *ClientConn, err error) {
	if o.Close != nil {
		o.Close(cc, err)
	}
}

var _ ClientConnObserver = ClientConnFuncObserver{}
