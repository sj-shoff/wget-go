package concurrency

import "sync"

// ConcurrentSet - потокобезопасное множество строк, реализованное через map
type ConcurrentSet struct {
	mu    sync.RWMutex
	items map[string]bool
}

// NewConcurrentSet создает новый экземпляр ConcurrentSet
func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{
		items: make(map[string]bool),
	}
}

// Add добавляет элемент в множество
func (cs *ConcurrentSet) Add(item string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.items[item] {
		return false
	}

	cs.items[item] = true
	return true
}

// Contains проверяет, содержится ли элемент в множестве
func (cs *ConcurrentSet) Contains(item string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.items[item]
}

// Remove удаляет элемент из множества
func (cs *ConcurrentSet) Remove(item string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.items, item)
}

// Size возвращает количество элементов в множестве
func (cs *ConcurrentSet) Size() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.items)
}

// Items возвращает все элементы множества в виде списка
func (cs *ConcurrentSet) Items() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if len(cs.items) == 0 {
		return nil
	}

	items := make([]string, 0, len(cs.items))
	for item := range cs.items {
		items = append(items, item)
	}

	return items
}
