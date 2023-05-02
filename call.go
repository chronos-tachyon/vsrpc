package vsrpc

import (
	"context"
	"sync"
	"time"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Call struct {
	options   []Option
	observers []Observer
	ctxOuter  context.Context
	ctxInner  context.Context
	cancel    context.CancelFunc
	deadline  time.Time
	method    Method
	conn      *Conn
	queue     *Queue
	id        ID
	role      Role

	mu     sync.Mutex
	cv     *sync.Cond
	status *Status
	state  state
}

func newCall(
	ctx context.Context,
	role Role,
	conn *Conn,
	id ID,
	method Method,
	deadline *timestamppb.Timestamp,
	options []Option,
) *Call {
	call := &Call{
		options: options,
		method:  method,
		conn:    conn,
		id:      id,
		role:    role,
	}
	call.cv = sync.NewCond(&call.mu)
	call.queue = NewQueue()

	for _, opt := range options {
		opt.applyToCall(call)
	}

	if deadline != nil {
		t := deadline.AsTime()
		if call.deadline.IsZero() || t.Before(call.deadline) {
			call.deadline = t
		}
	}

	ctx = WithContextCall(ctx, call)
	call.ctxOuter = ctx

	var cancel context.CancelFunc = func() {}
	if !call.deadline.IsZero() {
		ctx, cancel = context.WithDeadline(ctx, call.deadline)
	}

	call.ctxInner = ctx
	call.cancel = cancel
	return call
}

func (call *Call) Role() Role {
	if call == nil {
		return UnknownRole
	}
	return call.role
}

func (call *Call) Conn() *Conn {
	if call == nil {
		return nil
	}
	return call.conn
}

func (call *Call) Queue() *Queue {
	if call == nil {
		return nil
	}
	return call.queue
}

func (call *Call) Context() context.Context {
	if call == nil {
		return context.Background()
	}
	return call.ctxInner
}

func (call *Call) ID() ID {
	if call == nil {
		return 0
	}
	return call.id
}

func (call *Call) Method() Method {
	if call == nil {
		return ""
	}
	return call.method
}

func (call *Call) Send(payload *anypb.Any) error {
	if call == nil {
		return ErrCallClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return ErrCallClosed
	}
	if call.state >= stateShuttingDown {
		return ErrHalfClosed
	}

	conn := call.conn
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return call.lockedAbort(ErrConnClosed)
	}
	switch call.role {
	case ClientRole:
		if err := WriteRequest(call.ctxOuter, conn.pc, call.id, payload); err != nil {
			return conn.lockedGotWriteError(err)
		}
		onRequest(call.observers, call, payload)

	case ServerRole:
		if err := WriteResponse(call.ctxOuter, conn.pc, call.id, payload); err != nil {
			return conn.lockedGotWriteError(err)
		}
		onResponse(call.observers, call, payload)

	default:
		panic("unreachable")
	}
	return nil
}

func (call *Call) CloseSend() error {
	if call == nil {
		return ErrCallClosed
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return ErrCallClosed
	}
	if call.state >= stateShuttingDown || call.role != ClientRole {
		return nil
	}

	conn := call.conn
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return call.lockedAbort(ErrConnClosed)
	}
	if err := WriteHalfClose(call.ctxOuter, conn.pc, call.id); err != nil {
		return conn.lockedGotWriteError(err)
	}
	call.state = stateShuttingDown
	onHalfClose(call.observers, call)
	return nil
}

func (call *Call) Cancel() error {
	if call == nil {
		return InappropriateError{Op: "Cancel", Role: UnknownRole}
	}
	if call.role != ClientRole {
		return InappropriateError{Op: "Cancel", Role: call.role}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return ErrCallClosed
	}
	if call.state >= stateGoingAway {
		return nil
	}

	conn := call.conn
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return call.lockedAbort(ErrConnClosed)
	}
	if err := WriteCancel(call.ctxOuter, conn.pc, call.id); err != nil {
		return conn.lockedGotWriteError(err)
	}
	call.state = stateGoingAway
	call.cancel()
	onCancel(call.observers, call)
	return nil
}

func (call *Call) End(status *Status) error {
	if call == nil {
		return InappropriateError{Op: "End", Role: UnknownRole}
	}
	if call.role != ServerRole {
		return InappropriateError{Op: "End", Role: call.role}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return ErrCallClosed
	}

	conn := call.conn
	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return call.lockedAbort(ErrConnClosed)
	}

	if status == nil {
		status = &Status{Code: Status_OK}
	}

	if err := WriteEnd(call.ctxOuter, conn.pc, call.id, status); err != nil {
		return conn.lockedGotWriteError(err)
	}

	call.lockedEnd(status)
	return nil
}

func (call *Call) Wait() *Status {
	if call == nil {
		return &Status{Code: Status_CANCELLED}
	}

	call.mu.Lock()
	for call.state < stateClosed {
		call.cv.Wait()
	}
	status := call.status
	call.mu.Unlock()
	return status

}

func (call *Call) Close() error {
	if call == nil {
		return nil
	}
	if call.role == ClientRole {
		if err := call.Cancel(); err != nil {
			return err
		}
	}
	return call.Wait().AsError()
}

func (call *Call) gotRequest(payload *anypb.Any) error {
	if call == nil || call.role != ServerRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_REQUEST}}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateShuttingDown {
		return nil
	}

	call.queue.Push(payload)
	onRequest(call.observers, call, payload)
	return nil
}

func (call *Call) gotResponse(payload *anypb.Any) error {
	if call == nil || call.role != ClientRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_RESPONSE}}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return nil
	}

	call.queue.Push(payload)
	onResponse(call.observers, call, payload)
	return nil
}

func (call *Call) gotHalfClose() error {
	if call == nil || call.role != ServerRole {
		return nil
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateShuttingDown {
		return nil
	}

	call.queue.Done()
	call.state = stateShuttingDown
	onHalfClose(call.observers, call)
	return nil
}

func (call *Call) gotCancel() error {
	if call == nil || call.role != ServerRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_CANCEL}}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateGoingAway {
		return nil
	}

	call.queue.Done()
	call.state = stateGoingAway
	call.cancel()
	onCancel(call.observers, call)
	return nil
}

func (call *Call) gotEnd(status *Status) error {
	if call == nil || call.role != ClientRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_END}}
	}

	call.mu.Lock()
	defer call.mu.Unlock()

	if call.state >= stateClosed {
		return ProtocolViolationError{Err: ErrCallClosed}
	}

	call.lockedEnd(status)
	return nil
}

func (call *Call) lockedAbort(err error) error {
	if call.state < stateClosed {
		call.lockedEnd(Abort(err))
	}
	return err
}

func (call *Call) lockedEnd(status *Status) {
	if call.state >= stateClosed {
		return
	}

	if status == nil {
		status = &Status{Code: Status_OK}
	}

	call.queue.Done()
	call.state = stateClosed
	call.status = status
	call.cv.Broadcast()
	call.cancel()
	onEnd(call.observers, call, status)
	go call.conn.forgetCall(call)
}
