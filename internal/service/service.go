package service

import (
	"context"
	"wget-go/internal/domain"
)

// Downloader определяет контракт для загрузки ресурсов
type Downloader interface {
	Download(ctx context.Context, task domain.DownloadTask) (domain.DownloadResult, error)
	DetermineResourceType(url string, contentType string) domain.ResourceType
}

// TaskProcessor обрабатывает задачи и возвращает новые
type TaskProcessor interface {
	Process(ctx context.Context, task domain.DownloadTask) (domain.DownloadResult, []domain.DownloadTask, error)
}

// HTMLLinkParser извлекает ссылки из HTML
type HTMLLinkParser interface {
	ExtractLinks(htmlContent []byte, baseURL string) ([]string, error)
}

// ContentParser определяет тип контента
type ContentParser interface {
	ParseContentType(url string, contentTypeHeader string) domain.ResourceType
}
