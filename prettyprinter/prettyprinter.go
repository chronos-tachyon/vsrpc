package prettyprinter

import (
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type PrettyPrinter interface {
	PrettyPrintTo(buf []byte, detail *anypb.Any) []byte
}

type NoOp struct{}

func (NoOp) PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	return buf
}

var _ PrettyPrinter = NoOp{}

type Registry struct {
	mu sync.Mutex
	db map[protoreflect.FullName]PrettyPrinter
}

func (reg *Registry) Add(fullName protoreflect.FullName, pp PrettyPrinter) {
	if reg == nil {
		return
	}
	if pp == nil {
		pp = NoOp{}
	}

	reg.mu.Lock()
	if reg.db == nil {
		reg.db = make(map[protoreflect.FullName]PrettyPrinter, 16)
	}
	reg.db[fullName] = pp
	reg.mu.Unlock()
}

func (reg *Registry) Remove(fullName protoreflect.FullName) {
	if reg == nil {
		return
	}

	reg.mu.Lock()
	if reg.db != nil {
		delete(reg.db, fullName)
	}
	reg.mu.Unlock()
}

func (reg *Registry) Find(fullName protoreflect.FullName, inexact bool) PrettyPrinter {
	var pp PrettyPrinter
	if reg != nil {
		reg.mu.Lock()
		pp = reg.db[fullName]
		for inexact && pp == nil {
			i := strings.LastIndexByte(string(fullName), '.')
			if i < 0 {
				break
			}
			fullName = fullName[:i]
			pp = reg.db[fullName+".*"]
		}
		if inexact && pp == nil {
			pp = reg.db["*"]
		}
		reg.mu.Unlock()
	}
	if pp == nil {
		pp = NoOp{}
	}
	return pp
}

func (reg *Registry) PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	if detail == nil {
		return buf
	}
	fullName := detail.MessageName()
	return reg.Find(fullName, true).PrettyPrintTo(buf, detail)
}

var _ PrettyPrinter = (*Registry)(nil)

var GlobalRegistry atomic.Pointer[Registry]

func PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	return GlobalRegistry.Load().PrettyPrintTo(buf, detail)
}
