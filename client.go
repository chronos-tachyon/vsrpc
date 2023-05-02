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

	mu       sync.Mutex
	connSet  map[*Conn]void
	connList []*Conn
	state    state
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

	if c.state >= stateClosed {
		return nil, ErrClientClosed
	}

	if c.state >= stateShuttingDown {
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

	if c.state >= stateClosed {
		return nil, ErrClientClosed
	}

	if c.state >= stateShuttingDown {
		return nil, ErrClientClosing
	}

	return c.lockedDial(pc, options)
}

func (c *Client) Pick(ctx context.Context, picker Picker) (*Conn, error) {
	assert.NotNil(&ctx)
	assert.NotNil(&picker)

	if c == nil {
		return nil, ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= stateClosed {
		return nil, ErrClientClosed
	}

	if c.state >= stateShuttingDown {
		return nil, ErrClientClosing
	}

	tmp := make([]*Conn, len(c.connList))
	copy(tmp, c.connList)
	return picker.Pick(ctx, tmp)
}

func (c *Client) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if c == nil {
		return ErrClientClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= stateClosed {
		return ErrClientClosed
	}

	if c.state >= stateShuttingDown {
		return nil
	}

	onGlobalShutdown(c.observers)

	c.state = stateShuttingDown
	for _, conn := range c.connList {
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

	if c.state >= stateClosed {
		return ErrClientClosed
	}

	onGlobalClose(c.observers)

	c.state = stateClosed
	for _, conn := range c.connList {
		_ = conn.Close()
	}
	c.connSet = nil
	c.connList = nil
	return nil
}

func (c *Client) forgetConn(conn *Conn) {
	if c == nil || conn == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, found := c.connSet[conn]; !found {
		return
	}

	delete(c.connSet, conn)

	n := uint(len(c.connList))
	i := uint(0)
	j := uint(0)
	for i < n {
		item := c.connList[i]
		c.connList[i] = nil
		i++

		if item != conn {
			c.connList[j] = item
			j++
		}
	}
	c.connList = c.connList[:j]
}

func (c *Client) lockedDial(pc PacketConn, options []Option) (*Conn, error) {
	options = ConcatOptions(c.options, options...)
	conn := newConn(ClientRole, c, nil, pc, options)
	if c.connSet == nil {
		c.connList = make([]*Conn, 0, 16)
		c.connSet = make(map[*Conn]void, 16)
	}
	c.connList = append(c.connList, conn)
	c.connSet[conn] = void{}
	onDial(c.observers, conn)
	conn.start()
	return conn, nil
}
