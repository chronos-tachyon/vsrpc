package bufferpool

import (
	"sync"
)

var g sync.Map

func pool(size uint) *sync.Pool {
	pool, found := g.Load(size)
	if !found {
		pool, _ = g.LoadOrStore(size, &sync.Pool{
			New: func() any { return make([]byte, size) },
		})
	}
	return pool.(*sync.Pool)
}

func Allocate(size uint) []byte {
	return pool(size).Get().([]byte)
}

func Free(size uint, buffer []byte) {
	if uint(cap(buffer)) == size {
		buffer = buffer[:size]
		for i := range buffer {
			buffer[i] = 0
		}
		pool(size).Put(buffer)
	}
}
