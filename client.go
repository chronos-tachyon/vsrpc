package vsrpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type Client struct {
	options   []Option
	observers []Observer
	pd        PacketDialer

	mu    sync.Mutex
	conns map[*Conn]void
	state State
}

func NewClient(pd PacketDialer, options ...Option) *Client {
	options = ConcatOptions(nil, options...)
	c := &Client{options: options, pd: pd}
	for _, opt := range options {
		opt.applyToClient(c)
	}
	return c
}

func (c *Client) Dialer() PacketDialer {
	if c == nil {
		return nil
	}
	return c.pd
}

func (c *Client) Dial(ctx context.Context, addr net.Addr, options ...Option) (*Conn, error) {
	assert.NotNil(&ctx)
	assert.NotNil(&addr)

	if c == nil {
		return nil, ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return nil, ErrClientClosed
	}

	if c.state >= StateShuttingDown {
		return nil, ErrClientClosing
	}

	pd := c.pd
	if pd == nil {
		pd = DefaultPacketDialer
	}
	if pd == nil {
		panic(fmt.Errorf("vsrpc.DefaultPacketDialer is nil"))
	}

	pc, err := pd.DialPacket(ctx, addr)
	if err != nil {
		onDialError(c.observers, err)
		return nil, err
	}

	return c.lockedDial(pc, options)
}

func (c *Client) DialExisting(pc PacketConn, options ...Option) (*Conn, error) {
	assert.NotNil(&pc)

	if c == nil {
		return nil, ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return nil, ErrClientClosed
	}

	if c.state >= StateShuttingDown {
		return nil, ErrClientClosing
	}

	return c.lockedDial(pc, options)
}

func (c *Client) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if c == nil {
		return ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return ErrClientClosed
	}

	if c.state >= StateShuttingDown {
		return nil
	}

	onGlobalShutdown(c.observers)

	c.state = StateShuttingDown
	for conn := range c.conns {
		_ = conn.Shutdown(ctx)
	}
	return nil
}

func (c *Client) Close() error {
	if c == nil {
		return ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return ErrClientClosed
	}

	onGlobalClose(c.observers)

	c.state = StateClosed
	for conn := range c.conns {
		_ = conn.Close()
	}
	c.conns = nil
	return nil
}

func (c *Client) forgetConn(conn *Conn) {
	if c == nil || conn == nil {
		return
	}

	c.mu.Lock()
	if c.conns != nil {
		delete(c.conns, conn)
	}
	c.mu.Unlock()
}

func (c *Client) lockedDial(pc PacketConn, options []Option) (*Conn, error) {
	options = ConcatOptions(c.options, options...)
	conn := newConn(ClientRole, c, nil, pc, options)
	if c.conns == nil {
		c.conns = make(map[*Conn]void, 16)
	}
	c.conns[conn] = void{}
	onDial(c.observers, conn)
	conn.start()
	return conn, nil
}
