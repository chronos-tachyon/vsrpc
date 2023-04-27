package vsrpc

import (
	"context"
	"fmt"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type ServerConnObserver interface {
	OnCall(sc *ServerConn, call *ServerCall)
	OnShutdown(sc *ServerConn)
	OnGoAway(sc *ServerConn)
	OnReadError(sc *ServerConn, err error)
	OnWriteError(sc *ServerConn, err error)
	OnClose(sc *ServerConn, err error)
}

type ServerConn struct {
	pc PacketConn
	s  *Server

	mu        sync.Mutex
	calls     map[uint32]*ServerCall
	observers map[ServerConnObserver]void
	state     State
}

func newServerConn(pc PacketConn, s *Server) *ServerConn {
	sc := &ServerConn{pc: pc, s: s}
	go sc.readThread()
	return sc
}

func (sc *ServerConn) isValid() bool {
	return sc != nil && sc.pc != nil && sc.s != nil
}

func (sc *ServerConn) Conn() PacketConn {
	if !sc.isValid() {
		return nil
	}
	return sc.pc
}

func (sc *ServerConn) Server() *Server {
	if !sc.isValid() {
		return nil
	}
	return sc.s
}

func (sc *ServerConn) AddObserver(o ServerConnObserver) {
	if o == nil || !sc.isValid() {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateClosed {
		o.OnClose(sc, nil)
		return
	}

	for _, call := range sc.calls {
		o.OnCall(sc, call)
	}

	switch {
	case sc.state >= StateClosed:
		o.OnClose(sc, nil)

	case sc.state >= StateGoingAway:
		o.OnGoAway(sc)

	case sc.state >= StateShuttingDown:
		o.OnShutdown(sc)
	}

	if sc.observers == nil {
		sc.observers = make(map[ServerConnObserver]void, 16)
	}
	sc.observers[o] = void{}
}

func (sc *ServerConn) RemoveObserver(o ServerConnObserver) {
	if o == nil || !sc.isValid() {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.observers != nil {
		delete(sc.observers, o)
	}
}

func (sc *ServerConn) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if !sc.isValid() {
		return ErrClosed
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateClosed {
		return ErrClosed
	}

	if sc.state >= StateGoingAway {
		return nil
	}

	for o := range sc.observers {
		neverPanic(func() { o.OnGoAway(sc) })
	}

	err := WriteGoAway(ctx, sc.pc)
	if err != nil {
		return sc.lockedGotWriteError(err)
	}

	sc.state = StateGoingAway
	return nil
}

func (sc *ServerConn) Close() error {
	if !sc.isValid() {
		return ErrClosed
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateClosed {
		return ErrClosed
	}

	return sc.lockedClose()
}

func (sc *ServerConn) forgetCall(call *ServerCall) {
	if call == nil || !sc.isValid() {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()
	if sc.calls != nil {
		delete(sc.calls, call.id)
	}
}

func (sc *ServerConn) readThread() {
	ctx := context.Background()
	for {
		var frame Frame
		if err := ReadFrame(ctx, sc.pc, &frame); err != nil {
			sc.gotReadError(err)
			return
		}

		switch frame.Type {
		case Frame_NO_OP:
			// pass

		case Frame_SHUTDOWN:
			sc.gotShutdown()

		case Frame_BEGIN:
			if frame.CallId == 0 {
				sc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("BEGIN with call_id 0"),
				})
				return
			}

			if frame.Method == "" {
				sc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("BEGIN with empty method"),
				})
				return
			}

			if sc.gotBegin(frame.CallId, frame.Method) {
				return
			}

		case Frame_REQUEST:
			if frame.CallId == 0 {
				sc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("REQUEST with call_id 0"),
				})
				return
			}

			if sc.findCall(frame.CallId).gotRequest(frame.Payload) {
				return
			}

		case Frame_HALF_CLOSE:
			if frame.CallId == 0 {
				sc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("HALF_CLOSE with call_id 0"),
				})
				return
			}

			if sc.findCall(frame.CallId).gotHalfClose() {
				return
			}

		case Frame_CANCEL:
			if frame.CallId == 0 {
				sc.gotReadError(ProtocolViolationError{
					Err: fmt.Errorf("CANCEL with call_id 0"),
				})
				return
			}

			if sc.findCall(frame.CallId).gotCancel() {
				return
			}

		default:
			sc.gotReadError(ProtocolViolationError{
				Err: UnexpectedFrameTypeError{
					Actual: frame.Type,
					Expected: []Frame_Type{
						Frame_SHUTDOWN,
						Frame_BEGIN,
						Frame_REQUEST,
						Frame_HALF_CLOSE,
						Frame_CANCEL,
					},
				},
			})
			return
		}
	}
}

func (sc *ServerConn) gotReadError(err error) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.lockedGotReadError(err)
}

func (sc *ServerConn) gotShutdown() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.state >= StateShuttingDown {
		return
	}

	for o := range sc.observers {
		neverPanic(func() { o.OnShutdown(sc) })
	}

	sc.state = StateShuttingDown
}

func (sc *ServerConn) gotBegin(id uint32, method string) bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if call := sc.calls[id]; call != nil {
		sc.lockedGotReadError(ProtocolViolationError{
			Err: fmt.Errorf("duplicate BEGIN with call_id %d: old method %q, new method %q", id, call.method, method),
		})
		return true
	}

	if sc.state >= StateShuttingDown {
		return false
	}

	call := newServerCall(sc, id, method)
	h := sc.s.h
	go func() {
		err := h.Accept(call)
		_ = call.End(StatusFromError(err))
	}()

	for o := range sc.observers {
		neverPanic(func() { o.OnCall(sc, call) })
	}

	if sc.calls == nil {
		sc.calls = make(map[uint32]*ServerCall, 16)
	}
	sc.calls[id] = call
	return false
}

func (sc *ServerConn) findCall(id uint32) *ServerCall {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.calls[id]
}

func (sc *ServerConn) lockedGotReadError(err error) {
	if err == nil || sc.state >= StateClosed {
		return
	}

	for o := range sc.observers {
		neverPanic(func() { o.OnReadError(sc, err) })
	}

	if !IsRecoverable(err) {
		_ = sc.lockedClose()
	}
}

func (sc *ServerConn) lockedGotWriteError(err error) error {
	if err == nil || sc.state >= StateClosed {
		return err
	}

	for o := range sc.observers {
		neverPanic(func() { o.OnWriteError(sc, err) })
	}

	if !IsRecoverable(err) {
		_ = sc.lockedClose()
	}
	return err
}

func (sc *ServerConn) lockedClose() error {
	err := try(sc.pc.Close)

	for o := range sc.observers {
		neverPanic(func() { o.OnClose(sc, err) })
	}

	for _, call := range sc.calls {
		go call.abort(nil)
	}

	sc.state = StateClosed
	sc.calls = nil
	sc.observers = nil
	go sc.s.forgetConn(sc)
	return err
}
