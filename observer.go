package vsrpc

import (
	"google.golang.org/protobuf/types/known/anypb"
)

type Observer interface {
	OnAccept(conn *Conn)
	OnAcceptError(err error)
	OnDial(conn *Conn)
	OnDialError(err error)
	OnGlobalShutdown()
	OnGlobalClose()

	OnBegin(call *Call)
	OnRequest(call *Call, payload *anypb.Any)
	OnResponse(call *Call, payload *anypb.Any)
	OnHalfClose(call *Call)
	OnCancel(call *Call)
	OnEnd(call *Call, status *Status)

	OnShutdown(conn *Conn)
	OnGoAway(conn *Conn)

	OnReadError(conn *Conn, err error)
	OnWriteError(conn *Conn, err error)
	OnClose(conn *Conn, err error)
}

type BaseObserver struct{}

func (BaseObserver) OnAccept(conn *Conn)     {}
func (BaseObserver) OnAcceptError(err error) {}
func (BaseObserver) OnDial(conn *Conn)       {}
func (BaseObserver) OnDialError(err error)   {}
func (BaseObserver) OnGlobalShutdown()       {}
func (BaseObserver) OnGlobalClose()          {}

func (BaseObserver) OnBegin(call *Call)                        {}
func (BaseObserver) OnRequest(call *Call, payload *anypb.Any)  {}
func (BaseObserver) OnResponse(call *Call, payload *anypb.Any) {}
func (BaseObserver) OnHalfClose(call *Call)                    {}
func (BaseObserver) OnCancel(call *Call)                       {}
func (BaseObserver) OnEnd(call *Call, status *Status)          {}

func (BaseObserver) OnShutdown(conn *Conn) {}
func (BaseObserver) OnGoAway(conn *Conn)   {}

func (BaseObserver) OnReadError(conn *Conn, err error)  {}
func (BaseObserver) OnWriteError(conn *Conn, err error) {}
func (BaseObserver) OnClose(conn *Conn, err error)      {}

var _ Observer = BaseObserver{}

type FuncObserver struct {
	Accept         func(conn *Conn)
	AcceptError    func(err error)
	Dial           func(conn *Conn)
	DialError      func(err error)
	GlobalShutdown func()
	GlobalClose    func()

	Begin     func(call *Call)
	Request   func(call *Call, payload *anypb.Any)
	Response  func(call *Call, payload *anypb.Any)
	HalfClose func(call *Call)
	Cancel    func(call *Call)
	End       func(call *Call, status *Status)

	Shutdown func(conn *Conn)
	GoAway   func(conn *Conn)

	ReadError  func(conn *Conn, err error)
	WriteError func(conn *Conn, err error)
	Close      func(conn *Conn, err error)
}

func (o *FuncObserver) OnAccept(conn *Conn) {
	if o != nil && o.Accept != nil {
		o.Accept(conn)
	}
}

func (o *FuncObserver) OnAcceptError(err error) {
	if o != nil && o.AcceptError != nil {
		o.AcceptError(err)
	}
}

func (o *FuncObserver) OnDial(conn *Conn) {
	if o != nil && o.Dial != nil {
		o.Dial(conn)
	}
}

func (o *FuncObserver) OnDialError(err error) {
	if o != nil && o.DialError != nil {
		o.DialError(err)
	}
}

func (o *FuncObserver) OnGlobalShutdown() {
	if o != nil && o.GlobalShutdown != nil {
		o.GlobalShutdown()
	}
}

func (o *FuncObserver) OnGlobalClose() {
	if o != nil && o.GlobalClose != nil {
		o.GlobalClose()
	}
}

func (o *FuncObserver) OnBegin(call *Call) {
	if o != nil && o.Begin != nil {
		o.Begin(call)
	}
}

func (o *FuncObserver) OnRequest(call *Call, payload *anypb.Any) {
	if o != nil && o.Request != nil {
		o.Request(call, payload)
	}
}

func (o *FuncObserver) OnResponse(call *Call, payload *anypb.Any) {
	if o != nil && o.Response != nil {
		o.Response(call, payload)
	}
}

func (o *FuncObserver) OnHalfClose(call *Call) {
	if o != nil && o.HalfClose != nil {
		o.HalfClose(call)
	}
}

func (o *FuncObserver) OnCancel(call *Call) {
	if o != nil && o.Cancel != nil {
		o.Cancel(call)
	}
}

func (o *FuncObserver) OnEnd(call *Call, status *Status) {
	if o != nil && o.End != nil {
		o.End(call, status)
	}
}

func (o *FuncObserver) OnShutdown(conn *Conn) {
	if o != nil && o.Shutdown != nil {
		o.Shutdown(conn)
	}
}

func (o *FuncObserver) OnGoAway(conn *Conn) {
	if o != nil && o.GoAway != nil {
		o.GoAway(conn)
	}
}

func (o *FuncObserver) OnReadError(conn *Conn, err error) {
	if o != nil && o.ReadError != nil {
		o.ReadError(conn, err)
	}
}

func (o *FuncObserver) OnWriteError(conn *Conn, err error) {
	if o != nil && o.WriteError != nil {
		o.WriteError(conn, err)
	}
}

func (o *FuncObserver) OnClose(conn *Conn, err error) {
	if o != nil && o.Close != nil {
		o.Close(conn, err)
	}
}

var _ Observer = (*FuncObserver)(nil)

func WithObserver(o Observer) Option {
	if o == nil {
		return (*withObserver)(nil)
	}
	return &withObserver{o: o}
}

type withObserver struct {
	o Observer
}

func (opt *withObserver) applyToClient(c *Client) {
	if opt == nil || opt.o == nil || c == nil {
		return
	}
	c.observers = append(c.observers, opt.o)
}

func (opt *withObserver) applyToServer(s *Server) {
	if opt == nil || opt.o == nil || s == nil {
		return
	}
	s.observers = append(s.observers, opt.o)
}

func (opt *withObserver) applyToConn(conn *Conn) {
	if opt == nil || opt.o == nil || conn == nil {
		return
	}
}

func (opt *withObserver) applyToCall(call *Call) {
	if opt == nil || opt.o == nil || call == nil {
		return
	}
}

var _ Option = (*withObserver)(nil)

func onAccept(observers []Observer, conn *Conn) {
	for _, o := range observers {
		go o.OnAccept(conn)
	}
}

func onAcceptError(observers []Observer, err error) {
	for _, o := range observers {
		go o.OnAcceptError(err)
	}
}

func onDial(observers []Observer, conn *Conn) {
	for _, o := range observers {
		go o.OnDial(conn)
	}
}

func onDialError(observers []Observer, err error) {
	for _, o := range observers {
		go o.OnDialError(err)
	}
}

func onGlobalShutdown(observers []Observer) {
	for _, o := range observers {
		go o.OnGlobalShutdown()
	}
}

func onGlobalClose(observers []Observer) {
	for _, o := range observers {
		go o.OnGlobalClose()
	}
}

func onBegin(observers []Observer, call *Call) {
	for _, o := range observers {
		go o.OnBegin(call)
	}
}

func onRequest(observers []Observer, call *Call, payload *anypb.Any) {
	for _, o := range observers {
		go o.OnRequest(call, payload)
	}
}

func onResponse(observers []Observer, call *Call, payload *anypb.Any) {
	for _, o := range observers {
		go o.OnResponse(call, payload)
	}
}

func onHalfClose(observers []Observer, call *Call) {
	for _, o := range observers {
		go o.OnHalfClose(call)
	}
}

func onCancel(observers []Observer, call *Call) {
	for _, o := range observers {
		go o.OnCancel(call)
	}
}

func onEnd(observers []Observer, call *Call, status *Status) {
	for _, o := range observers {
		go o.OnEnd(call, status)
	}
}

func onShutdown(observers []Observer, conn *Conn) {
	for _, o := range observers {
		go o.OnShutdown(conn)
	}
}

func onGoAway(observers []Observer, conn *Conn) {
	for _, o := range observers {
		go o.OnGoAway(conn)
	}
}

func onReadError(observers []Observer, conn *Conn, err error) {
	for _, o := range observers {
		go o.OnReadError(conn, err)
	}
}

func onWriteError(observers []Observer, conn *Conn, err error) {
	for _, o := range observers {
		go o.OnWriteError(conn, err)
	}
}

func onClose(observers []Observer, conn *Conn, err error) {
	for _, o := range observers {
		go o.OnClose(conn, err)
	}
}
