package concurrency

import "sync"

// ConcurrentQueue - потокобезопасная очередь
type ConcurrentQueue struct {
	mu    sync.Mutex
	items []interface{}
}

// NewConcurrentQueue создает новый экземпляр ConcurrentQueue
func NewConcurrentQueue() *ConcurrentQueue {
	return &ConcurrentQueue{
		items: []interface{}{},
	}
}

// Enqueue добавляет элемент в очередь
func (cq *ConcurrentQueue) Enqueue(item interface{}) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	cq.items = append(cq.items, item)
}

// Dequeue извлекает элемент из очереди
func (cq *ConcurrentQueue) Dequeue() (interface{}, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.items) == 0 {
		return nil, false
	}

	item := cq.items[0]
	cq.items = cq.items[1:]
	return item, true
}

// Size возвращает количество элементов в очереди
func (q *ConcurrentQueue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// IsEmpty возвращает true, если очередь пуста
func (q *ConcurrentQueue) IsEmpty() bool {
	return q.Size() == 0
}
