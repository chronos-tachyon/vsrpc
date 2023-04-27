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

type PrettyPrinterRegistry struct {
	mu sync.Mutex
	db map[string]PrettyPrinter
}

func (reg *PrettyPrinterRegistry) Add(typeURL string, pp PrettyPrinter) {
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

func (reg *PrettyPrinterRegistry) Remove(typeURL string) {
	if reg == nil {
		return
	}

	reg.mu.Lock()
	if reg.db != nil {
		delete(reg.db, typeURL)
	}
	reg.mu.Unlock()
}

func (reg *PrettyPrinterRegistry) Find(typeURL string) PrettyPrinter {
	if reg == nil {
		return nil
	}

	reg.mu.Lock()
	pp := reg.db[typeURL]
	reg.mu.Unlock()
	return pp
}

func (reg *PrettyPrinterRegistry) PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
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

var _ PrettyPrinter = (*PrettyPrinterRegistry)(nil)

var Registry atomic.Pointer[PrettyPrinterRegistry]

func PrettyPrintTo(buf []byte, detail *anypb.Any) []byte {
	if detail == nil {
		return buf
	}
	return Registry.Load().PrettyPrintTo(buf, detail)
}
