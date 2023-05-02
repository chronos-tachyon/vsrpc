package vsrpc

import (
	"context"
	"net"
	"sync"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type Conn struct {
	options   []Option
	observers []Observer
	pc        PacketConn
	c         *Client
	s         *Server
	role      Role

	mu    sync.Mutex
	calls map[ID]*Call
	id    ID
	state state
}

func newConn(role Role, c *Client, s *Server, pc PacketConn, options []Option) *Conn {
	conn := &Conn{
		options: options,
		pc:      pc,
		c:       c,
		s:       s,
		role:    role,
	}
	for _, opt := range options {
		opt.applyToConn(conn)
	}
	return conn
}

func (conn *Conn) Role() Role {
	if conn == nil {
		return UnknownRole
	}
	return conn.role
}

func (conn *Conn) Client() *Client {
	if conn == nil {
		return nil
	}
	return conn.c
}

func (conn *Conn) Server() *Server {
	if conn == nil {
		return nil
	}
	return conn.s
}

func (conn *Conn) PacketConn() PacketConn {
	if conn == nil {
		return nil
	}
	return conn.pc
}

func (conn *Conn) LocalAddr() net.Addr {
	if conn == nil || conn.pc == nil {
		return nil
	}
	return conn.pc.LocalAddr()
}

func (conn *Conn) RemoteAddr() net.Addr {
	if conn == nil || conn.pc == nil {
		return nil
	}
	return conn.pc.RemoteAddr()
}

func (conn *Conn) Begin(ctx context.Context, method Method, options ...Option) (*Call, error) {
	if conn == nil {
		return nil, InappropriateError{Op: "Begin", Role: UnknownRole}
	}
	if conn.role != ClientRole {
		return nil, InappropriateError{Op: "Begin", Role: conn.role}
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return nil, ErrConnClosed
	}
	if conn.state >= stateGoingAway {
		return nil, ErrConnGoingAway
	}
	if conn.state >= stateShuttingDown {
		return nil, ErrConnShuttingDown
	}

	id := conn.id
	if numCalls := uint(len(conn.calls)); id == 0 || numCalls == 0 || uint(id) > (2*numCalls) {
		id = 1
	}
	_, found := conn.calls[id]
	for found {
		id++
		_, found = conn.calls[id]
	}
	conn.id = id

	ctx = WithContextClient(ctx, conn.c)
	ctx = WithContextConn(ctx, conn)

	if err := WriteBegin(ctx, conn.pc, id, method); err != nil {
		return nil, conn.lockedGotWriteError(err)
	}

	options = ConcatOptions(conn.options, options...)
	call := newCall(ctx, ClientRole, conn, id, method, nil, options)
	if conn.calls == nil {
		conn.calls = make(map[ID]*Call, 16)
	}
	conn.calls[id] = call
	return call, nil
}

func (conn *Conn) Shutdown(ctx context.Context) error {
	if conn == nil {
		return ErrConnClosed
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.state >= stateClosed {
		return ErrConnClosed
	}

	switch conn.role {
	case ClientRole:
		if conn.state >= stateShuttingDown {
			return nil
		}

		ctx = WithContextClient(ctx, conn.c)
		ctx = WithContextConn(ctx, conn)

		if err := WriteShutdown(ctx, conn.pc); err != nil {
			return conn.lockedGotWriteError(err)
		}

		conn.state = stateShuttingDown
		onShutdown(conn.observers, conn)

	case ServerRole:
		if conn.state >= stateGoingAway {
			return nil
		}

		ctx = WithContextServer(ctx, conn.s)
		ctx = WithContextConn(ctx, conn)

		if err := WriteGoAway(ctx, conn.pc); err != nil {
			return conn.lockedGotWriteError(err)
		}

		conn.state = stateGoingAway
		onGoAway(conn.observers, conn)

	default:
		panic("unreachable")
	}
	return nil
}

func (conn *Conn) Close() error {
	if conn == nil {
		return ErrConnClosed
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	return conn.lockedClose()
}

func (conn *Conn) start() {
	go conn.readThread()
}

func (conn *Conn) readThread() {
	ctx := context.Background()
	ctx = WithContextClient(ctx, conn.c)
	ctx = WithContextServer(ctx, conn.s)
	ctx = WithContextConn(ctx, conn)
	for {
		var frame Frame
		if err := ReadFrame(ctx, conn.pc, &frame); err != nil {
			conn.gotReadError(err)
			return
		}
		if err := conn.dispatch(ctx, &frame); err != nil {
			conn.gotReadError(err)
			return
		}
	}
}

func (conn *Conn) dispatch(ctx context.Context, frame *Frame) error {
	frameType := frame.Type
	id := ID(frame.CallId)
	method := Method(frame.Method)
	payload := frame.Payload
	status := frame.Status
	deadline := frame.Deadline

	if expectZeroCallId(frameType) && id != 0 {
		return ProtocolViolationError{Err: CallIdError{Type: frameType, ID: id}}
	}

	if expectNonZeroCallId(frameType) && id == 0 {
		return ProtocolViolationError{Err: CallIdError{Type: frameType, ID: id}}
	}

	switch frameType {
	case Frame_NO_OP:
		return conn.gotNoOp()
	case Frame_SHUTDOWN:
		return conn.gotShutdown()
	case Frame_GO_AWAY:
		return conn.gotGoAway()
	case Frame_BEGIN:
		return conn.gotBegin(ctx, id, method, deadline)
	case Frame_REQUEST:
		return conn.findCall(id).gotRequest(payload)
	case Frame_RESPONSE:
		return conn.findCall(id).gotResponse(payload)
	case Frame_HALF_CLOSE:
		return conn.findCall(id).gotHalfClose()
	case Frame_CANCEL:
		return conn.findCall(id).gotCancel()
	case Frame_END:
		return conn.findCall(id).gotEnd(status)
	default:
		return ProtocolViolationError{Err: FrameTypeError{Type: frameType}}
	}
}

func (conn *Conn) findCall(id ID) *Call {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	return conn.calls[id]
}

func (conn *Conn) gotNoOp() error {
	return nil
}

func (conn *Conn) gotShutdown() error {
	if conn.role != ServerRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_SHUTDOWN}}
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.state < stateShuttingDown {
		conn.state = stateShuttingDown
		onShutdown(conn.observers, conn)
	}
	return nil
}

func (conn *Conn) gotGoAway() error {
	if conn.role != ClientRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_GO_AWAY}}
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()
	if conn.state < stateGoingAway {
		conn.state = stateGoingAway
		onGoAway(conn.observers, conn)
	}
	return nil
}

func (conn *Conn) gotBegin(ctx context.Context, id ID, method Method, deadline *timestamppb.Timestamp) error {
	if conn.role != ServerRole {
		return ProtocolViolationError{Err: FrameTypeError{Type: Frame_BEGIN}}
	}

	if deadline != nil {
		if err := deadline.CheckValid(); err != nil {
			return ProtocolViolationError{Err: err}
		}
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if call := conn.calls[id]; call != nil {
		return ProtocolViolationError{Err: DuplicateCallError{ID: id, Old: call.method, New: method}}
	}
	if conn.state >= stateShuttingDown {
		return nil
	}

	call := newCall(ctx, ServerRole, conn, id, method, deadline, conn.options)
	if conn.calls == nil {
		conn.calls = make(map[ID]*Call, 16)
	}
	conn.calls[id] = call
	onBegin(conn.observers, call)
	go conn.handle(call)
	return nil
}

func (conn *Conn) handle(call *Call) {
	h := conn.Server().Handler()
	err := try(func() error { return h.Handle(call) })
	_ = call.End(StatusFromError(err))
}

func (conn *Conn) gotReadError(err error) {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.lockedGotReadError(err)
}

func (conn *Conn) lockedGotReadError(err error) {
	if err == nil || conn.state >= stateClosed {
		return
	}
	onReadError(conn.observers, conn, err)
	if !IsRecoverable(err) {
		_ = conn.lockedClose()
	}
}

func (conn *Conn) lockedGotWriteError(err error) error {
	if err == nil || conn.state >= stateClosed {
		return err
	}
	onWriteError(conn.observers, conn, err)
	if !IsRecoverable(err) {
		_ = conn.lockedClose()
	}
	return err
}

func (conn *Conn) lockedClose() error {
	if conn.state >= stateClosed {
		return ErrConnClosed
	}

	err := try(conn.pc.Close)
	conn.state = stateClosed
	for _, call := range conn.calls {
		call.gotEnd(Abort(ErrConnClosed))
	}
	conn.calls = nil
	onClose(conn.observers, conn, err)
	if conn.c != nil {
		go conn.c.forgetConn(conn)
	}
	if conn.s != nil {
		go conn.s.forgetConn(conn)
	}
	return err
}

func (conn *Conn) forgetCall(call *Call) {
	if conn == nil || call == nil {
		return
	}

	conn.mu.Lock()
	if conn.calls != nil {
		delete(conn.calls, call.id)
	}
	conn.mu.Unlock()
}
