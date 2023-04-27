package vsrpc

import (
	"strings"
	"sync"

	"github.com/chronos-tachyon/assert"
)

type Handler interface {
	Accept(*ServerCall) error
}

type HandlerFunc func(*ServerCall) error

func (h HandlerFunc) Accept(call *ServerCall) error {
	if h == nil {
		method := call.Method()
		return NoSuchMethodError{Method: method}
	}
	return h(call)
}

type HandlerMux struct {
	mu       sync.Mutex
	handlers map[string]Handler
}

func (mux *HandlerMux) AddFunc(method string, fn func(*ServerCall) error) {
	mux.Add(method, HandlerFunc(fn))
}

func (mux *HandlerMux) Add(method string, h Handler) {
	assert.NotNil(&h)

	if mux == nil {
		return
	}

	mux.mu.Lock()
	if mux.handlers == nil {
		mux.handlers = make(map[string]Handler, 64)
	}
	mux.handlers[method] = h
	mux.mu.Unlock()
}

func (mux *HandlerMux) AddAll(m map[string]Handler) {
	if mux == nil {
		return
	}

	mLen := len(m)
	if mLen <= 0 {
		return
	}

	mux.mu.Lock()
	if mux.handlers == nil {
		if mLen < 64 {
			mLen = 64
		}
		mux.handlers = make(map[string]Handler, mLen)
	}
	for method, h := range m {
		mux.handlers[method] = h
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) Remove(method string) {
	if mux == nil {
		return
	}

	mux.mu.Lock()
	if mux.handlers != nil {
		delete(mux.handlers, method)
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) RemoveAll(methods []string) {
	if mux == nil {
		return
	}

	if len(methods) <= 0 {
		return
	}

	mux.mu.Lock()
	if mux.handlers != nil {
		for _, method := range methods {
			delete(mux.handlers, method)
		}
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) Find(method string) Handler {
	if mux == nil {
		return nil
	}

	mux.mu.Lock()
	h := mux.handlers[method]
	mux.mu.Unlock()
	return h
}

func (mux *HandlerMux) Accept(call *ServerCall) error {
	method := call.Method()
	h := mux.Find(method)
	for h == nil {
		i := strings.LastIndex(method, ".")
		if i < 0 {
			break
		}
		method = method[:i]
		h = mux.Find(method + ".*")
	}
	if h == nil {
		h = mux.Find("*")
	}
	if h == nil {
		method = call.Method()
		return NoSuchMethodError{Method: method}
	}
	return h.Accept(call)
}

var _ Handler = (*HandlerMux)(nil)
