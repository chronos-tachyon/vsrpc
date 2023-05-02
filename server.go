package vsrpc

import (
	"context"
	"net"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type Server struct {
	options   []Option
	observers []Observer
	pl        PacketListener
	h         Handler

	mu      sync.Mutex
	connSet map[*Conn]void
	state   state
}

func NewServer(pl PacketListener, h Handler, options ...Option) *Server {
	if h == nil {
		h = &HandlerMux{}
	}

	options = ConcatOptions(nil, options...)
	s := &Server{options: options, pl: pl, h: h}
	for _, opt := range options {
		opt.applyToServer(s)
	}
	if pl != nil {
		go s.acceptThread()
	}
	return s
}

func (s *Server) isValid() bool {
	return s != nil && s.pl != nil
}

func (s *Server) Listener() PacketListener {
	if s == nil {
		return nil
	}
	return s.pl
}

func (s *Server) Addr() net.Addr {
	if s == nil || s.pl == nil {
		return nil
	}
	return s.Listener().Addr()
}

func (s *Server) Handler() Handler {
	if s == nil {
		return nil
	}
	return s.h
}

func (s *Server) AcceptExisting(pc PacketConn, options ...Option) error {
	assert.NotNil(&pc)

	if s == nil {
		return ErrServerClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= stateClosed {
		return pc.Close()
	}

	options = ConcatOptions(s.options, options...)
	conn := newConn(ServerRole, nil, s, pc, options)
	if s.connSet == nil {
		s.connSet = make(map[*Conn]void, 16)
	}
	s.connSet[conn] = void{}
	onAccept(s.observers, conn)
	conn.start()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if s == nil {
		return ErrServerClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= stateClosed {
		return ErrServerClosed
	}

	if s.state >= stateShuttingDown {
		return nil
	}

	onGlobalShutdown(s.observers)

	var err error
	if s.pl != nil {
		err = try(s.pl.Close)
	}

	s.state = stateShuttingDown
	for conn := range s.connSet {
		_ = conn.Shutdown(ctx)
	}
	return err
}

func (s *Server) Close() error {
	if s == nil {
		return ErrServerClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= stateClosed {
		return ErrServerClosed
	}

	onGlobalClose(s.observers)

	var err error
	if s.pl != nil && s.state < stateShuttingDown {
		err = try(s.pl.Close)
	}

	s.state = stateClosed
	for conn := range s.connSet {
		_ = conn.Close()
	}
	s.connSet = nil
	return err
}

func (s *Server) forgetConn(conn *Conn) {
	if conn == nil || s == nil {
		return
	}

	s.mu.Lock()
	if s.connSet != nil {
		delete(s.connSet, conn)
	}
	s.mu.Unlock()
}

func (s *Server) acceptThread() {
	ctx := context.Background()
	looping := true
	for looping {
		looping = s.dispatch(s.pl.AcceptPacket(ctx))
	}
}

func (s *Server) dispatch(pc PacketConn, err error) bool {
	if err != nil {
		onAcceptError(s.observers, err)
		return IsRecoverable(err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= stateClosed {
		_ = try(pc.Close)
		return false
	}

	conn := newConn(ServerRole, nil, s, pc, s.options)
	if s.connSet == nil {
		s.connSet = make(map[*Conn]void, 16)
	}
	s.connSet[conn] = void{}
	onAccept(s.observers, conn)
	conn.start()
	return true
}
