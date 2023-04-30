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

	mu    sync.Mutex
	conns map[*Conn]void
	state State
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

func (s *Server) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if s == nil {
		return ErrClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= StateClosed {
		return ErrClosed
	}

	if s.state >= StateShuttingDown {
		return nil
	}

	onGlobalShutdown(s.observers)

	var err error
	if s.pl != nil {
		err = try(s.pl.Close)
	}

	s.state = StateShuttingDown
	for conn := range s.conns {
		_ = conn.Shutdown(ctx)
	}
	return err
}

func (s *Server) Close() error {
	if s == nil {
		return ErrClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= StateClosed {
		return ErrClosed
	}

	onGlobalClose(s.observers)

	var err error
	if s.pl != nil && s.state < StateShuttingDown {
		err = try(s.pl.Close)
	}

	s.state = StateClosed
	for conn := range s.conns {
		_ = conn.Close()
	}
	s.conns = nil
	return err
}

func (s *Server) forgetConn(conn *Conn) {
	if conn == nil || s == nil {
		return
	}

	s.mu.Lock()
	if s.conns != nil {
		delete(s.conns, conn)
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

	if s.state >= StateClosed {
		_ = try(pc.Close)
		return false
	}

	conn := newConn(ServerRole, nil, s, pc, s.options)
	if s.conns == nil {
		s.conns = make(map[*Conn]void, 16)
	}
	s.conns[conn] = void{}
	onAccept(s.observers, conn)
	conn.start()
	return true
}
