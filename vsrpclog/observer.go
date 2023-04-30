package vsrpclog

import (
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/chronos-tachyon/vsrpc"
)

type Observer struct {
	Logger *zerolog.Logger
}

func (o Observer) GetLogger() *zerolog.Logger {
	if o.Logger == nil {
		nopLogger := zerolog.Nop()
		return &nopLogger
	}
	return o.Logger
}

func (o Observer) OnAccept(conn *vsrpc.Conn) {
	o.GetLogger().Info().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Msg("accept")
}

func (o Observer) OnAcceptError(err error) {
	o.GetLogger().Error().
		Err(err).
		Msg("accept error")
}

func (o Observer) OnDial(conn *vsrpc.Conn) {
	o.GetLogger().Info().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Msg("dial")
}

func (o Observer) OnDialError(err error) {
	o.GetLogger().Error().
		Err(err).
		Msg("dial error")
}

func (o Observer) OnGlobalShutdown() {
	o.GetLogger().Info().Msg("global shutdown")
}

func (o Observer) OnGlobalClose() {
	o.GetLogger().Info().Msg("global close")
}

func (o Observer) OnBegin(call *vsrpc.Call) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Msg("RPC begin")
}

func (o Observer) OnRequest(call *vsrpc.Call, payload *anypb.Any) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Str("requestType", payload.GetTypeUrl()).
		Msg("RPC request payload")
}

func (o Observer) OnResponse(call *vsrpc.Call, payload *anypb.Any) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Str("responseType", payload.GetTypeUrl()).
		Msg("RPC response payload")
}

func (o Observer) OnHalfClose(call *vsrpc.Call) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Msg("RPC half-close")
}

func (o Observer) OnCancel(call *vsrpc.Call) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Msg("RPC cancel")
}

func (o Observer) OnEnd(call *vsrpc.Call, status *vsrpc.Status) {
	o.GetLogger().Info().
		Uint32("rpcID", uint32(call.ID())).
		Str("rpcMethod", string(call.Method())).
		Int32("statusCode", int32(status.GetCode())).
		Str("statusText", status.GetText()).
		Msg("RPC end")
}

func (o Observer) OnShutdown(conn *vsrpc.Conn) {
	o.GetLogger().Info().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Msg("connection client-side shutdown")
}

func (o Observer) OnGoAway(conn *vsrpc.Conn) {
	o.GetLogger().Info().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Msg("connection server-side shutdown")
}

func (o Observer) OnReadError(conn *vsrpc.Conn, err error) {
	o.GetLogger().Error().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Err(err).
		Msg("connection read error")
}

func (o Observer) OnWriteError(conn *vsrpc.Conn, err error) {
	o.GetLogger().Error().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Err(err).
		Msg("connection write error")
}

func (o Observer) OnClose(conn *vsrpc.Conn, err error) {
	if err != nil {
		o.GetLogger().Error().
			Stringer("localAddr", conn.LocalAddr()).
			Stringer("remoteAddr", conn.RemoteAddr()).
			Err(err).
			Msg("connection close error")
		return
	}
	o.GetLogger().Info().
		Stringer("localAddr", conn.LocalAddr()).
		Stringer("remoteAddr", conn.RemoteAddr()).
		Msg("connection close")
}

var _ vsrpc.Observer = Observer{}
