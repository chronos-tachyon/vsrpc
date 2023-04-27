package prettyprinter

import (
	"strings"
	"sync"
	"sync/atomic"

	"google.golang.org/protobuf/types/known/anypb"
)

type PrettyPrinter interface {
	PrettyPrintTo(buf []byte, detail *anypb.Any) []byte
}

type Registry struct {
	mu sync.Mutex
	db map[string]PrettyPrinter
}

func (reg *Registry) Add(typeURL string, pp PrettyPrinter) {
	if reg == nil || pp == nil {
		return
	}

	reg.mu.Lock()
	if reg.db == nil {
		reg.db = make(map[string]PrettyPrinter, 16)
	}
	reg.db[typeURL] = pp
	reg.mu.Unlock()
}

func (reg *Registry) Remove(typeURL string) {
	if reg == nil {
		return
	}

	reg.mu.Lock()
	if reg.db != nil {
		delete(reg.db, typeURL)
	}
	reg.mu.Unlock()
}

func (reg *Registry) Find(typeURL string) PrettyPrinter {
	if reg == nil {
		return nil
	}

	reg.mu.Lock()
	pp := reg.db[typeURL]
	reg.mu.Unlock()
	return pp
}

func (reg *Registry) PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	if detail == nil {
		return buf
	}

	url := detail.TypeUrl
	pp := reg.Find(url)
	if pp == nil {
		if i := strings.LastIndex(url, "/"); i >= 0 {
			url = url[:i] + "/*"
			pp = reg.Find(url)
		}
	}
	if pp == nil {
		pp = reg.Find("*")
	}
	if pp == nil {
		return buf
	}
	return pp.PrettyPrintTo(buf, detail)
}

var _ PrettyPrinter = (*Registry)(nil)

var GlobalRegistry atomic.Pointer[Registry]

func PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	if detail == nil {
		return buf
	}
	return GlobalRegistry.Load().PrettyPrintTo(buf, detail)
}
