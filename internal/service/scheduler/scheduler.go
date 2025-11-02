package scheduler

import (
	"context"
	"log"
	"net/url"
	"sync/atomic"
	"time"

	"wget-go/internal/config"
	"wget-go/internal/domain"
	"wget-go/internal/service"
	"wget-go/internal/storage"
	"wget-go/pkg/concurrency"
	"wget-go/pkg/utils"
)

// DownloadScheduler реализует планировщик
type DownloadScheduler struct {
	config       *config.Config
	downloader   service.Downloader
	pathResolver storage.PathResolver
	visited      *concurrency.ConcurrentSet
	workerPool   *concurrency.WorkerPool
	baseURL      *url.URL

	totalTasks     int32
	completedTasks int32
	failedTasks    int32
	pendingTasks   int32

	stopChan chan struct{}
}

// New создает новый планировщик
func New(
	config *config.Config,
	downloader service.Downloader,
	pathResolver storage.PathResolver,
) *DownloadScheduler {
	baseURL, _ := url.Parse(config.URL)

	return &DownloadScheduler{
		config:       config,
		downloader:   downloader,
		pathResolver: pathResolver,
		visited:      concurrency.NewConcurrentSet(),
		baseURL:      baseURL,
		stopChan:     make(chan struct{}),
	}
}

// Start запускает процесс скачивания
func (s *DownloadScheduler) Start(ctx context.Context) error {

	s.workerPool = concurrency.NewWorkerPool(s.config.Workers, s.processTask)
	results := s.workerPool.Start(ctx)

	s.scheduleInitialTask()
	return s.processResults(ctx, results)
}

// processResults обрабатывает результаты скачивания
func (s *DownloadScheduler) processResults(ctx context.Context, results <-chan interface{}) error {
	progressTicker := time.NewTicker(2 * time.Second)
	defer progressTicker.Stop()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				s.printFinalStats()
				return nil
			}

			s.handleResult(result.(domain.DownloadResult))

			// Проверяем завершение после обработки каждого результата
			if s.shouldStop() {
				s.workerPool.Close()
				s.printFinalStats()
				return nil
			}

		case <-progressTicker.C:
			s.printProgress()

		case <-ctx.Done():
			log.Printf("Download interrupted by user")
			s.workerPool.Close()
			s.printFinalStats()
			return ctx.Err()

		case <-s.stopChan:
			s.workerPool.Close()
			s.printFinalStats()
			return nil
		}
	}
}

// processTask обрабатывает одну задачу
func (s *DownloadScheduler) processTask(task interface{}) interface{} {
	downloadTask := task.(domain.DownloadTask)
	ctx := context.TODO()

	result, err := s.downloader.Download(ctx, downloadTask)
	if err != nil {
		result.Error = err
	}

	return result
}

// handleResult обрабатывает результаты скачивания
func (s *DownloadScheduler) handleResult(result domain.DownloadResult) {
	atomic.AddInt32(&s.pendingTasks, -1)

	if result.Error != nil {
		atomic.AddInt32(&s.failedTasks, 1)
		log.Printf("Failed to download %s: %v", result.Task.URL, result.Error)
	} else {
		atomic.AddInt32(&s.completedTasks, 1)
		log.Printf("Downloaded %s -> %s", result.Task.URL, result.FilePath)

		if result.Task.Depth < s.config.MaxDepth {
			s.scheduleNewTasks(result)
		}
	}
}

// sheduleNewTasks добавляет новые задачи на основе найденных ссылок
func (s *DownloadScheduler) scheduleNewTasks(result domain.DownloadResult) {
	for _, link := range result.Links {
		absoluteURL, err := s.pathResolver.ResolveAbsoluteURL(result.Task.URL, link)
		if err != nil {
			continue
		}

		if !s.shouldDownload(absoluteURL) {
			continue
		}

		newTask := domain.DownloadTask{
			URL:       absoluteURL,
			Depth:     result.Task.Depth + 1,
			ParentURL: result.Task.URL,
		}

		s.Schedule(newTask)
	}
}

// Schedule добавляет новую задачу в планировщик
func (s *DownloadScheduler) Schedule(task domain.DownloadTask) {
	atomic.AddInt32(&s.totalTasks, 1)
	atomic.AddInt32(&s.pendingTasks, 1)
	s.workerPool.Submit(task)
}

// Stats возвращает статистику планировщика
func (s *DownloadScheduler) scheduleInitialTask() {
	initialTask := domain.DownloadTask{
		URL:   s.config.URL,
		Depth: 0,
		Type:  domain.ResourceHTML,
	}
	s.Schedule(initialTask)
	s.visited.Add(s.normalizeURL(s.config.URL))
}

// shouldDownload проверяет, нужно ли скачать ссылку
func (s *DownloadScheduler) shouldDownload(testURL string) bool {
	if !s.pathResolver.IsSameDomain(s.config.URL, testURL) {
		return false
	}

	normalized := s.normalizeURL(testURL)
	return s.visited.Add(normalized)
}

// normalizeURL нормализует URL
func (s *DownloadScheduler) normalizeURL(rawURL string) string {
	normalized, err := utils.NormalizeURL(rawURL)
	if err != nil {
		return rawURL
	}
	return normalized
}

// shouldStop проверяет, нужно ли остановить процесс
func (s *DownloadScheduler) shouldStop() bool {
	total := atomic.LoadInt32(&s.totalTasks)
	completed := atomic.LoadInt32(&s.completedTasks)
	failed := atomic.LoadInt32(&s.failedTasks)
	pending := atomic.LoadInt32(&s.pendingTasks)

	// Останавливаемся когда все задачи завершены и нет ожидающих
	return total > 0 && pending == 0 && total == completed+failed
}

// printProgress выводит прогресс
func (s *DownloadScheduler) printProgress() {
	total := atomic.LoadInt32(&s.totalTasks)
	completed := atomic.LoadInt32(&s.completedTasks)
	failed := atomic.LoadInt32(&s.failedTasks)
	pending := atomic.LoadInt32(&s.pendingTasks)

	progress := s.calculateProgress(total, completed)

	log.Printf("Progress: %.1f%% | Completed: %d | Failed: %d | Pending: %d | Total: %d",
		progress, completed, failed, pending, total)
}

// printFinalStats вычисляет и выводит прогресс
func (s *DownloadScheduler) calculateProgress(total, completed int32) float32 {
	if total == 0 {
		return 0
	}
	return float32(completed) / float32(total) * 100
}

// printFinalStats выводит финальную статистику
func (s *DownloadScheduler) printFinalStats() {
	total := atomic.LoadInt32(&s.totalTasks)
	completed := atomic.LoadInt32(&s.completedTasks)
	failed := atomic.LoadInt32(&s.failedTasks)

	log.Printf("Download completed!")
	log.Printf("Final stats:")
	log.Printf("  Total URLs processed: %d", s.visited.Size())
	log.Printf("  Tasks completed: %d", completed)
	log.Printf("  Tasks failed: %d", failed)
	log.Printf("  Success rate: %.1f%%", s.calculateSuccessRate(total, completed))
}

// calculateSuccessRate вычисляет успешность
func (s *DownloadScheduler) calculateSuccessRate(total, completed int32) float32 {
	if total == 0 {
		return 0
	}
	return float32(completed) / float32(total) * 100
}

// Stats возвращает статистику планировщика
func (s *DownloadScheduler) Stats() service.SchedulerStats {
	return service.SchedulerStats{
		TotalTasks:     int(atomic.LoadInt32(&s.totalTasks)),
		CompletedTasks: int(atomic.LoadInt32(&s.completedTasks)),
		FailedTasks:    int(atomic.LoadInt32(&s.failedTasks)),
		ActiveWorkers:  s.config.Workers,
	}
}
