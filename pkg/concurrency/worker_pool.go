package concurrency

import (
	"context"
	"log"
	"sync"
)

// TaskProcessor - функция для обработки задачи
type TaskProcessor func(task interface{}) interface{}

// WorkerPool управляет пулом воркеров
type WorkerPool struct {
	workerCount int
	taskQueue   chan interface{}
	resultChan  chan interface{}
	processor   TaskProcessor
	wg          sync.WaitGroup
}

// NewWorkerPool создает новый пул воркеров
func NewWorkerPool(workerCount int, processor TaskProcessor) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskQueue:   make(chan interface{}, workerCount*2),
		resultChan:  make(chan interface{}, workerCount*2),
		processor:   processor,
	}
}

// worker обрабатывает задачи из очереди
func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()

	for {
		select {
		case task, ok := <-wp.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue closed, shutting down", id)
				return
			}

			result := wp.processor(task)
			wp.resultChan <- result
		case <-ctx.Done():
			log.Printf("Worker %d: context cancelled, shutting down", id)
			return
		}
	}
}

// Start запускает воркеров и возвращает канал результатов
func (wp *WorkerPool) Start(ctx context.Context) <-chan interface{} {
	wp.wg.Add(wp.workerCount)

	for i := 0; i < wp.workerCount; i++ {
		go wp.worker(ctx, i)
	}

	go func() {
		wp.wg.Wait()
		close(wp.resultChan)
	}()

	return wp.resultChan
}

// Submit добавляет задачу в очередь
func (wp *WorkerPool) Submit(task interface{}) {
	wp.taskQueue <- task
}

// Close закрывает пул воркеров
func (wp *WorkerPool) Close() {
	close(wp.taskQueue)
}
