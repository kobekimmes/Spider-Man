

package q

import (
	"errors"
	"sync"
)


type Q[T any] struct {
	Elements []T
	Mutex sync.Mutex
}

func (q *Q[T]) Dequeue() (T, error) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()

	var val T
	if len(q.Elements) <= 0 {
		return val, errors.New("queue underflow")
	}
	val = q.Elements[0]
	q.Elements = q.Elements[1:]

	return val, nil
}

func (q *Q[T]) Enqueue(val T) {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	q.Elements = append(q.Elements, val)
}

func (q *Q[T]) Size() int {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	return len(q.Elements)
}

func (q *Q[T]) Capacity() int {
	q.Mutex.Lock()
	defer q.Mutex.Unlock()
	return cap(q.Elements)
}





