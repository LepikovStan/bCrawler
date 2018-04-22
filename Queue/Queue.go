package Queue

import (
	"sync"
)

type Q struct {
	arr []interface{}
	mu  *sync.RWMutex
}

func (q *Q) Unshift(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	newArr := make([]interface{}, len(q.arr)+1)
	newArr[0] = item
	for i := 0; i < len(q.arr); i++ {
		newArr[i+1] = q.arr[i]
	}
	q.arr = newArr
}

func (q *Q) Push(item interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.arr = append(q.arr, item)
}

func (q *Q) Pop() interface{} {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.arr) == 0 {
		return nil
	}
	item := q.arr[0]
	q.arr = q.arr[1:len(q.arr)]
	return item
}

func (q Q) Len() int {
	return len(q.arr)
}

func (q *Q) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.arr = []interface{}{}
}

func NewQ() *Q {
	q := new(Q)
	q.mu = &sync.RWMutex{}
	q.arr = make([]interface{}, 0)
	return q
}
