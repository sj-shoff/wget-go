package service

import (
	"context"
	"wget-go/internal/domain"
)

// Downloader загружает ресурсы
type Downloader interface {
	Download(ctx context.Context, task domain.DownloadTask) (domain.DownloadResult, error)
	DetermineResourceType(url string, contentType string) domain.ResourceType
}

// Extractor извлекает ссылки из контента
type Extractor interface {
	ExtractLinks(content []byte, baseURL string, contentType string) ([]string, error)
}

// HTMLParser парсит HTML контент
type HTMLParser interface {
	Parse(htmlContent []byte, baseURL string) ([]string, error)
}

// Scheduler управляет процессом скачивания
type Scheduler interface {
	Start(ctx context.Context) error
	Schedule(task domain.DownloadTask)
	Stats() SchedulerStats
}

// SchedulerStats статистика планировщика
type SchedulerStats struct {
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	ActiveWorkers  int
}
