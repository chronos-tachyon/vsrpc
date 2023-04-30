package vsrpc

import (
	"context"
	"io"
	"net"
)

var DefaultPacketDialer PacketDialer

type PacketDialer interface {
	DialPacket(ctx context.Context, addr net.Addr) (PacketConn, error)
	ListenPacket(ctx context.Context, addr net.Addr) (PacketListener, error)
}

type PacketListener interface {
	io.Closer

	AcceptPacket(ctx context.Context) (PacketConn, error)

	Addr() net.Addr
}

type PacketConn interface {
	io.Closer
	PacketReader
	PacketWriter

	LocalAddr() net.Addr

	RemoteAddr() net.Addr
}
