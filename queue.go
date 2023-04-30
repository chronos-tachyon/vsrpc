package vsrpc

import (
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type Queue struct {
	mu   sync.Mutex
	list []*anypb.Any
	cv1  *sync.Cond
	cv2  *sync.Cond
	done bool
}

func NewQueue() *Queue {
	q := new(Queue)
	q.cv1 = sync.NewCond(&q.mu)
	q.cv2 = sync.NewCond(&q.mu)
	return q
}

func (q *Queue) Push(value *anypb.Any) bool {
	if q == nil {
		return false
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.done {
		return false
	}

	if value != nil {
		value = proto.Clone(value).(*anypb.Any)
	}

	q.list = append(q.list, value)
	if q.cv1 != nil {
		q.cv1.Signal()
	}
	return true
}

func (q *Queue) Done() {
	if q == nil {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.done {
		return
	}

	q.done = true
	if q.cv1 != nil {
		q.cv1.Broadcast()
	}
	if q.cv2 != nil {
		q.cv2.Broadcast()
	}
}

func (q *Queue) Recv(blocking bool) (*anypb.Any, bool, bool) {
	if q == nil {
		return nil, false, true
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if q.done {
		return nil, false, true
	}

	if blocking && len(q.list) <= 0 {
		if q.cv1 == nil {
			q.cv1 = sync.NewCond(&q.mu)
		}
		for len(q.list) <= 0 && !q.done {
			q.cv1.Wait()
		}
	}

	if len(q.list) <= 0 {
		return nil, false, q.done
	}

	item := q.list[0]
	q.list[0] = nil
	q.list = q.list[1:]
	return item, true, q.done
}
