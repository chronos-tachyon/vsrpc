package vsrpc

import (
	"strings"
	"sync"
)

type Handler interface {
	Handle(*ServerCall) error
}

type HandlerFunc func(*ServerCall) error

func NoSuchMethod(call *ServerCall) error {
	method := call.Method()
	return NoSuchMethodError{Method: method}
}

func (fn HandlerFunc) Handle(call *ServerCall) error {
	if fn == nil {
		return NoSuchMethod(call)
	}
	return fn(call)
}

var _ Handler = HandlerFunc(nil)

type HandlerMux struct {
	mu sync.Mutex
	db map[string]Handler
}

func (mux *HandlerMux) Reset() {
	mux.mu.Lock()
	if mux.db != nil {
		for method := range mux.db {
			delete(mux.db, method)
		}
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) AddFunc(fn func(*ServerCall) error, methods ...string) {
	mux.Add(HandlerFunc(fn), methods...)
}

func (mux *HandlerMux) Add(h Handler, methods ...string) {
	if mux == nil || len(methods) <= 0 {
		return
	}

	if h == nil {
		h = HandlerFunc(nil)
	}

	mux.mu.Lock()
	if mux.db == nil {
		mux.db = make(map[string]Handler, 64)
	}
	for _, method := range methods {
		mux.db[method] = h
	}
	mux.mu.Unlock()
}

func (mux *HandlerMux) Remove(methods ...string) {
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

func (mux *HandlerMux) Find(method string) Handler {
	if mux == nil {
		return nil
	}

	mux.mu.Lock()
	h := mux.db[method]
	mux.mu.Unlock()
	return h
}

func (mux *HandlerMux) Handle(call *ServerCall) error {
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
		return NoSuchMethod(call)
	}
	return h.Handle(call)
}

var _ Handler = (*HandlerMux)(nil)
