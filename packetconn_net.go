package vsrpc

import (
	"context"
	"net"
	"time"

	"github.com/chronos-tachyon/assert"
	"github.com/chronos-tachyon/vsrpc/bufferpool"
)

const DefaultMaxPacketSize = (1 << 24)

type NetworkPacketDialer struct {
	Dialer               *net.Dialer
	ListenConfig         *net.ListenConfig
	Now                  func() time.Time
	MaxPacketSize        uint
	AcceptTimeout        time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	AcceptTimeoutEnabled bool
	ReadTimeoutEnabled   bool
	WriteTimeoutEnabled  bool
	UnlinkOnClose        bool
}

func (pd *NetworkPacketDialer) DialPacket(ctx context.Context, addr net.Addr) (PacketConn, error) {
	var zeroDialer net.Dialer
	dialer := &zeroDialer
	if pd != nil && pd.Dialer != nil {
		dialer = pd.Dialer
	}

	conn, err := dialer.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	pc := &NetworkPacketConn{Conn: conn}
	if pd != nil {
		pc.Now = pd.Now
		pc.MaxPacketSize = pd.MaxPacketSize
		pc.ReadTimeout = pd.ReadTimeout
		pc.WriteTimeout = pd.WriteTimeout
		pc.ReadTimeoutEnabled = pd.ReadTimeoutEnabled
		pc.WriteTimeoutEnabled = pd.WriteTimeoutEnabled
	}
	return pc, nil
}

func (pd *NetworkPacketDialer) ListenPacket(ctx context.Context, addr net.Addr) (PacketListener, error) {
	var zeroConfig net.ListenConfig
	config := &zeroConfig
	if pd != nil && pd.ListenConfig != nil {
		config = pd.ListenConfig
	}

	listener, err := config.Listen(ctx, addr.Network(), addr.String())
	if err != nil {
		return nil, err
	}

	if pd != nil && pd.UnlinkOnClose {
		if x, ok := listener.(*net.UnixListener); ok {
			x.SetUnlinkOnClose(true)
		}
	}

	pl := &NetworkPacketListener{Listener: listener}
	if pd != nil {
		pl.Now = pd.Now
		pl.MaxPacketSize = pd.MaxPacketSize
		pl.AcceptTimeout = pd.AcceptTimeout
		pl.ReadTimeout = pd.ReadTimeout
		pl.WriteTimeout = pd.WriteTimeout
		pl.AcceptTimeoutEnabled = pd.AcceptTimeoutEnabled
		pl.ReadTimeoutEnabled = pd.ReadTimeoutEnabled
		pl.WriteTimeoutEnabled = pd.WriteTimeoutEnabled
	}
	return pl, nil
}

var _ PacketDialer = (*NetworkPacketDialer)(nil)

type NetworkPacketListener struct {
	Listener             net.Listener
	Now                  func() time.Time
	MaxPacketSize        uint
	AcceptTimeout        time.Duration
	ReadTimeout          time.Duration
	WriteTimeout         time.Duration
	AcceptTimeoutEnabled bool
	ReadTimeoutEnabled   bool
	WriteTimeoutEnabled  bool
}

func (pl *NetworkPacketListener) AcceptPacket(ctx context.Context) (pc PacketConn, err error) {
	if pl == nil || pl.Listener == nil {
		return nil, ErrClosed
	}

	var now time.Time
	s, ok := pl.Listener.(deadlineSetter)
	if ok {
		var deadline time.Time
		now = pl.now()
		if pl.AcceptTimeoutEnabled {
			deadline = now.Add(pl.AcceptTimeout)
		}
		if t, ok := ctx.Deadline(); ok {
			if deadline.IsZero() || t.Before(deadline) {
				deadline = t
			}
		}

		err = s.SetDeadline(deadline)
		if err != nil {
			return
		}
	}

	err = Watch(ctx, func() {
		if ok {
			_ = s.SetDeadline(now)
		}
	}, func() error {
		conn, err := pl.Listener.Accept()
		if err != nil {
			return err
		}

		pc = &NetworkPacketConn{
			Conn:                conn,
			Now:                 pl.Now,
			MaxPacketSize:       pl.MaxPacketSize,
			ReadTimeout:         pl.ReadTimeout,
			WriteTimeout:        pl.WriteTimeout,
			ReadTimeoutEnabled:  pl.ReadTimeoutEnabled,
			WriteTimeoutEnabled: pl.WriteTimeoutEnabled,
		}
		return nil
	})
	return
}

func (pl *NetworkPacketListener) Addr() net.Addr {
	if pl == nil || pl.Listener == nil {
		return nil
	}
	return pl.Listener.Addr()
}

func (pl *NetworkPacketListener) Close() error {
	if pl == nil || pl.Listener == nil {
		return nil
	}
	return pl.Listener.Close()
}

func (pl *NetworkPacketListener) now() time.Time {
	var fn func() time.Time = time.Now
	if pl != nil && pl.Now != nil {
		fn = pl.Now
	}
	return fn()
}

var _ PacketListener = (*NetworkPacketListener)(nil)

type NetworkPacketConn struct {
	Conn                net.Conn
	Now                 func() time.Time
	MaxPacketSize       uint
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	ReadTimeoutEnabled  bool
	WriteTimeoutEnabled bool
}

func (pc *NetworkPacketConn) ReadPacket(ctx context.Context) (packet []byte, dispose func(), err error) {
	assert.NotNil(&ctx)

	if pc == nil || pc.Conn == nil {
		return nil, nil, ErrClosed
	}

	var deadline time.Time
	now := pc.now()
	if pc.ReadTimeoutEnabled {
		deadline = now.Add(pc.ReadTimeout)
	}
	if t, ok := ctx.Deadline(); ok {
		if deadline.IsZero() || t.Before(deadline) {
			deadline = t
		}
	}

	err = pc.Conn.SetReadDeadline(deadline)
	if err != nil {
		return
	}

	size := pc.MaxPacketSize
	if size == 0 {
		size = DefaultMaxPacketSize
	}

	buffer := bufferpool.Allocate(size)
	err = Watch(ctx, func() {
		_ = pc.Conn.SetReadDeadline(now)
	}, func() error {
		n, err := pc.Conn.Read(buffer)
		if err != nil {
			bufferpool.Free(size, buffer)
			return err
		}
		packet = buffer[:n]
		dispose = func() { bufferpool.Free(size, buffer) }
		return nil
	})
	return
}

func (pc *NetworkPacketConn) WritePacket(ctx context.Context, packet []byte) error {
	assert.NotNil(&ctx)

	if pc == nil || pc.Conn == nil {
		return ErrClosed
	}

	var deadline time.Time
	now := pc.now()
	if pc.WriteTimeoutEnabled {
		deadline = now.Add(pc.WriteTimeout)
	}
	if t, ok := ctx.Deadline(); ok {
		if deadline.IsZero() || t.Before(deadline) {
			deadline = t
		}
	}

	err := pc.Conn.SetWriteDeadline(deadline)
	if err != nil {
		return err
	}

	return Watch(ctx, func() {
		_ = pc.Conn.SetWriteDeadline(now)
	}, func() error {
		_, err := pc.Conn.Write(packet)
		return err
	})
}

func (pc *NetworkPacketConn) LocalAddr() net.Addr {
	if pc == nil || pc.Conn == nil {
		return nil
	}
	return pc.Conn.LocalAddr()
}

func (pc *NetworkPacketConn) RemoteAddr() net.Addr {
	if pc == nil || pc.Conn == nil {
		return nil
	}
	return pc.Conn.RemoteAddr()
}

func (pc *NetworkPacketConn) Close() error {
	if pc == nil || pc.Conn == nil {
		return ErrClosed
	}
	return pc.Conn.Close()
}

func (pc *NetworkPacketConn) now() time.Time {
	var fn func() time.Time = time.Now
	if pc != nil && pc.Now != nil {
		fn = pc.Now
	}
	return fn()
}

var _ PacketConn = (*NetworkPacketConn)(nil)
