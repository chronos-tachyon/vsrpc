package vsrpc

import (
	"context"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type ServerObserver interface {
	OnAccept(s *Server, sc *ServerConn)
	OnAcceptError(s *Server, err error)
	OnShutdown(s *Server)
	OnClose(s *Server)
}

type Server struct {
	pl PacketListener
	h  Handler

	mu        sync.Mutex
	conns     map[*ServerConn]void
	observers map[ServerObserver]void
	state     State
}

func NewServer(pl PacketListener, h Handler) *Server {
	if pl == nil {
		return nil
	}

	if h == nil {
		h = &HandlerMux{}
	}

	s := &Server{pl: pl, h: h}
	go s.acceptThread()
	return s
}

func (s *Server) isValid() bool {
	return s != nil && s.pl != nil
}

func (s *Server) Listener() PacketListener {
	if !s.isValid() {
		return nil
	}
	return s.pl
}

func (s *Server) Handler() Handler {
	if !s.isValid() {
		return nil
	}
	return s.h
}

func (s *Server) AddObserver(o ServerObserver) {
	if o == nil || !s.isValid() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= StateClosed {
		o.OnClose(s)
		return
	}

	for sc := range s.conns {
		o.OnAccept(s, sc)
	}

	if s.state >= StateShuttingDown {
		o.OnShutdown(s)
	}

	if s.observers == nil {
		s.observers = make(map[ServerObserver]void, 16)
	}
	s.observers[o] = void{}
}

func (s *Server) RemoveObserver(o ServerObserver) {
	if o == nil || !s.isValid() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.observers != nil {
		delete(s.observers, o)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if !s.isValid() {
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

	for o := range s.observers {
		neverPanic(func() { o.OnShutdown(s) })
	}

	err := try(s.pl.Close)

	var errs []error
	for sc := range s.conns {
		if err := sc.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if err == nil && len(errs) > 0 {
		err = joinErrors(errs)
	}

	s.state = StateShuttingDown
	return err
}

func (s *Server) Close() error {
	if !s.isValid() {
		return ErrClosed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state >= StateClosed {
		return ErrClosed
	}

	for o := range s.observers {
		neverPanic(func() { o.OnClose(s) })
	}

	var err error
	if s.state < StateShuttingDown {
		err = try(s.pl.Close)
	}

	var errs []error
	for sc := range s.conns {
		if err := sc.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if err == nil && len(errs) > 0 {
		err = joinErrors(errs)
	}

	s.state = StateClosed
	s.conns = nil
	s.observers = nil
	return err
}

func (s *Server) forgetConn(sc *ServerConn) {
	if sc == nil || !s.isValid() {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conns != nil {
		delete(s.conns, sc)
	}
}

func (s *Server) acceptThread() {
	ctx := context.Background()
	for {
		pc, err := s.pl.AcceptPacket(ctx)
		if err != nil {
			s.gotAcceptError(err)
			if IsRecoverable(err) {
				continue
			}
			break
		}

		sc := newServerConn(pc, s)
		s.gotAccept(sc)
	}
}

func (s *Server) gotAccept(sc *ServerConn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for o := range s.observers {
		neverPanic(func() { o.OnAccept(s, sc) })
	}

	if s.conns == nil {
		s.conns = make(map[*ServerConn]void, 16)
	}
	s.conns[sc] = void{}
}

func (s *Server) gotAcceptError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for o := range s.observers {
		neverPanic(func() { o.OnAcceptError(s, err) })
	}
}

type ServerNoOpObserver struct{}

func (ServerNoOpObserver) OnAccept(s *Server, sc *ServerConn) {}
func (ServerNoOpObserver) OnAcceptError(s *Server, err error) {}
func (ServerNoOpObserver) OnShutdown(s *Server)               {}
func (ServerNoOpObserver) OnClose(s *Server)                  {}

var _ ServerObserver = ServerNoOpObserver{}

type ServerFuncObserver struct {
	Accept      func(s *Server, sc *ServerConn)
	AcceptError func(s *Server, err error)
	Shutdown    func(s *Server)
	Close       func(s *Server)
}

func (o ServerFuncObserver) OnAccept(s *Server, sc *ServerConn) {
	if o.Accept != nil {
		o.Accept(s, sc)
	}
}

func (o ServerFuncObserver) OnAcceptError(s *Server, err error) {
	if o.AcceptError != nil {
		o.AcceptError(s, err)
	}
}

func (o ServerFuncObserver) OnShutdown(s *Server) {
	if o.Shutdown != nil {
		o.Shutdown(s)
	}
}

func (o ServerFuncObserver) OnClose(s *Server) {
	if o.Close != nil {
		o.Close(s)
	}
}

var _ ServerObserver = ServerFuncObserver{}
