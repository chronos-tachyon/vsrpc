package vsrpc

import (
	"strings"
	"sync"
)

type Handler interface {
	Handle(*Call) error
}

type HandlerFunc func(*Call) error

func (fn HandlerFunc) Handle(call *Call) error {
	if fn == nil {
		method := call.Method()
		return NoSuchMethodError{Method: method}
	}
	return fn(call)
}

var _ Handler = HandlerFunc(nil)

type HandlerMux struct {
	mu sync.Mutex
	db map[Method]Handler
}

func (mux *HandlerMux) Reset() {
	if mux == nil {
		return
	}

	mux.mu.Lock()
	if mux.db != nil {
		for method := range mux.db {
			delete(mux.db, method)
		}
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) AddFunc(fn func(*Call) error, methods ...Method) {
	mux.Add(HandlerFunc(fn), methods...)
}

func (mux *HandlerMux) Add(h Handler, methods ...Method) {
	if mux == nil || len(methods) <= 0 {
		return
	}

	if h == nil {
		h = HandlerFunc(nil)
	}

	mux.mu.Lock()
	if mux.db == nil {
		mux.db = make(map[Method]Handler, 64)
	}
	for _, method := range methods {
		mux.db[method] = h
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) Remove(methods ...Method) {
	if mux == nil || len(methods) <= 0 {
		return
	}

	mux.mu.Lock()
	if mux.db != nil {
		for _, method := range methods {
			delete(mux.db, method)
		}
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) Find(method Method, inexact bool) Handler {
	if mux == nil {
		return nil
	}

	mux.mu.Lock()
	h := mux.db[method]
	for inexact && h == nil {
		i := strings.LastIndex(string(method), ".")
		if i < 0 {
			break
		}
		method = method[:i]
		h = mux.db[method+".*"]
	}
	if inexact && h == nil {
		h = mux.db["*"]
	}
	mux.mu.Unlock()
	return h
}

func (mux *HandlerMux) Handle(call *Call) error {
	method := call.Method()
	h := mux.Find(method, true)
	if h == nil {
		return NoSuchMethodError{Method: method}
	}
	return h.Handle(call)
}

var _ Handler = (*HandlerMux)(nil)
