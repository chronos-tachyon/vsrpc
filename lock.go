package vsrpc

import (
	"sync"
)

type lockHolder struct {
	mu     sync.Locker
	isHeld bool
}

func newLock(mu sync.Locker) *lockHolder {
	lock := &lockHolder{mu: mu}
	lock.Lock()
	return lock
}

func (lock *lockHolder) Lock() {
	if lock == nil {
		return
	}
	if lock.isHeld {
		panic("lock is already held")
	}
	lock.mu.Lock()
	lock.isHeld = true
}

func (lock *lockHolder) Unlock() {
	if lock == nil {
		return
	}
	if !lock.isHeld {
		panic("lock is not currently held")
	}
	lock.mu.Unlock()
	lock.isHeld = false
}

func (lock *lockHolder) Dispose() {
	if lock == nil || !lock.isHeld {
		return
	}
	lock.mu.Unlock()
	lock.isHeld = false
}

var _ sync.Locker = (*lockHolder)(nil)
