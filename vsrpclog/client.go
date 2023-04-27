package vsrpclog

import (
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/chronos-tachyon/vsrpc"
)

type ClientLogger struct {
	Logger *zerolog.Logger
}

func (log *ClientLogger) OnDial(c *vsrpc.Client, cc *vsrpc.ClientConn) {
	log.Logger.Info().
		Stringer("localAddr", cc.Conn().LocalAddr()).
		Stringer("remoteAddr", cc.Conn().RemoteAddr()).
		Msg("dial")

	cc.AddObserver(log)
}

func (log *ClientLogger) OnDialError(c *vsrpc.Client, err error) {
	log.Logger.Error().
		Err(err).
		Msg("dial error")
}

func (log *ClientLogger) OnGlobalShutdown(c *vsrpc.Client) {
}

func (log *ClientLogger) OnGlobalClose(c *vsrpc.Client) {
}

func (log *ClientLogger) OnCall(cc *vsrpc.ClientConn, call *vsrpc.ClientCall) {
	call.AddObserver(log)
}

func (log *ClientLogger) OnShutdown(cc *vsrpc.ClientConn) {
}

func (log *ClientLogger) OnGoAway(cc *vsrpc.ClientConn) {
}

func (log *ClientLogger) OnReadError(cc *vsrpc.ClientConn, err error) {
}

func (log *ClientLogger) OnWriteError(cc *vsrpc.ClientConn, err error) {
}

func (log *ClientLogger) OnClose(cc *vsrpc.ClientConn, err error) {
}

func (log *ClientLogger) OnRequest(call *vsrpc.ClientCall, payload *anypb.Any) {
}

func (log *ClientLogger) OnResponse(call *vsrpc.ClientCall, payload *anypb.Any) {
}

func (log *ClientLogger) OnHalfClose(call *vsrpc.ClientCall) {
}

func (log *ClientLogger) OnCancel(call *vsrpc.ClientCall) {
}

func (log *ClientLogger) OnEnd(call *vsrpc.ClientCall, status *vsrpc.Status) {
}

var (
	_ vsrpc.ClientObserver     = (*ClientLogger)(nil)
	_ vsrpc.ClientConnObserver = (*ClientLogger)(nil)
	_ vsrpc.ClientCallObserver = (*ClientLogger)(nil)
)
