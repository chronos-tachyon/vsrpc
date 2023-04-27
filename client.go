package vsrpc

import (
	"context"
	"net"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type ClientObserver interface {
	OnDial(c *Client, cc *ClientConn)
	OnDialError(c *Client, err error)
	OnGlobalShutdown(c *Client)
	OnGlobalClose(c *Client)
}

type Client struct {
	pd PacketDialer

	mu        sync.Mutex
	conns     map[*ClientConn]void
	observers map[ClientObserver]void
	state     State
}

func NewClient(pd PacketDialer) *Client {
	if pd == nil {
		return nil
	}

	c := &Client{pd: pd}
	return c
}

func (c *Client) isValid() bool {
	return c != nil && c.pd != nil
}

func (c *Client) Dialer() PacketDialer {
	if !c.isValid() {
		return nil
	}
	return c.pd
}

func (c *Client) AddObserver(o ClientObserver) {
	if o == nil || !c.isValid() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		o.OnGlobalClose(c)
		return
	}

	for cc := range c.conns {
		o.OnDial(c, cc)
	}

	if c.state >= StateShuttingDown {
		o.OnGlobalShutdown(c)
	}

	if c.observers == nil {
		c.observers = make(map[ClientObserver]void, 16)
	}
	c.observers[o] = void{}
}

func (c *Client) RemoveObserver(o ClientObserver) {
	if o == nil || !c.isValid() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.observers != nil {
		delete(c.observers, o)
	}
}

func (c *Client) Dial(ctx context.Context, addr net.Addr) (*ClientConn, error) {
	assert.NotNil(&ctx)
	assert.NotNil(&addr)

	if !c.isValid() {
		return nil, ErrClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return nil, ErrClosed
	}

	if c.state >= StateShuttingDown {
		return nil, ErrClientClosing
	}

	pc, err := c.pd.DialPacket(ctx, addr)
	if err != nil {
		for o := range c.observers {
			neverPanic(func() { o.OnDialError(c, err) })
		}
		return nil, err
	}

	return c.lockedDial(ctx, pc)
}

func (c *Client) DialExisting(ctx context.Context, pc PacketConn) (*ClientConn, error) {
	assert.NotNil(&ctx)
	assert.NotNil(&pc)

	if !c.isValid() {
		return nil, ErrClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return nil, ErrClosed
	}

	if c.state >= StateShuttingDown {
		return nil, ErrClientClosing
	}

	return c.lockedDial(ctx, pc)
}

func (c *Client) Shutdown(ctx context.Context) error {
	assert.NotNil(&ctx)

	if !c.isValid() {
		return ErrClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return ErrClosed
	}

	if c.state >= StateShuttingDown {
		return nil
	}

	for o := range c.observers {
		neverPanic(func() { o.OnGlobalShutdown(c) })
	}

	var errs []error
	for cc := range c.conns {
		if err := cc.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	c.state = StateShuttingDown
	return joinErrors(errs)
}

func (c *Client) Close() error {
	if !c.isValid() {
		return ErrClosed
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state >= StateClosed {
		return ErrClosed
	}

	for o := range c.observers {
		neverPanic(func() { o.OnGlobalClose(c) })
	}

	var errs []error
	for cc := range c.conns {
		if err := cc.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	c.state = StateClosed
	c.conns = nil
	c.observers = nil
	return joinErrors(errs)
}

func (c *Client) forgetConn(cc *ClientConn) {
	if cc == nil || !c.isValid() {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns != nil {
		delete(c.conns, cc)
	}
}

func (c *Client) lockedDial(ctx context.Context, pc PacketConn) (*ClientConn, error) {
	cc := newClientConn(pc, c)

	for o := range c.observers {
		neverPanic(func() { o.OnDial(c, cc) })
	}

	if c.conns == nil {
		c.conns = make(map[*ClientConn]void, 16)
	}
	c.conns[cc] = void{}
	return cc, nil
}

type ClientNoOpObserver struct{}

func (ClientNoOpObserver) OnDial(c *Client, cc *ClientConn) {}
func (ClientNoOpObserver) OnDialError(c *Client, err error) {}
func (ClientNoOpObserver) OnGlobalShutdown(c *Client)       {}
func (ClientNoOpObserver) OnGlobalClose(c *Client)          {}

var _ ClientObserver = ClientNoOpObserver{}

type ClientFuncObserver struct {
	Dial      func(c *Client, cc *ClientConn)
	DialError func(c *Client, err error)
	Shutdown  func(c *Client)
	Close     func(c *Client)
}

func (o ClientFuncObserver) OnDial(c *Client, cc *ClientConn) {
	if o.Dial != nil {
		o.Dial(c, cc)
	}
}

func (o ClientFuncObserver) OnDialError(c *Client, err error) {
	if o.DialError != nil {
		o.DialError(c, err)
	}
}

func (o ClientFuncObserver) OnGlobalShutdown(c *Client) {
	if o.Shutdown != nil {
		o.Shutdown(c)
	}
}

func (o ClientFuncObserver) OnGlobalClose(c *Client) {
	if o.Close != nil {
		o.Close(c)
	}
}

var _ ClientObserver = ClientFuncObserver{}
