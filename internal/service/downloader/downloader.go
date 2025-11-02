package downloader

import (
	"context"
	"log"
	"strings"
	httpserver "wget-go/internal/delivery/http-server"
	"wget-go/internal/domain"
	"wget-go/internal/service"
	"wget-go/internal/storage"
)

// WebDownloader реализует сервис загрузки
type WebDownloader struct {
	httpClient   httpserver.Client
	fileManager  storage.FileManager
	pathResolver storage.PathResolver
	linkRewriter storage.LinkRewriter
	extractor    service.Extractor
}

// New создает новый загрузчик
func New(
	httpClient httpserver.Client,
	fileManager storage.FileManager,
	pathResolver storage.PathResolver,
	linkRewriter storage.LinkRewriter,
	extractor service.Extractor,
) *WebDownloader {
	return &WebDownloader{
		httpClient:   httpClient,
		fileManager:  fileManager,
		pathResolver: pathResolver,
		linkRewriter: linkRewriter,
		extractor:    extractor,
	}
}

// Download загружает ресурс по URL
func (d *WebDownloader) Download(ctx context.Context, task domain.DownloadTask) (domain.DownloadResult, error) {
	result := domain.DownloadResult{Task: task}

	// определяем тип ресурса через HEAD запрос
	resourceType, err := d.determineResourceTypeWithHead(ctx, task.URL)
	if err != nil {
		// Если HEAD не удался, продолжаем с GET
		resourceType = d.determineResourceTypeByURL(task.URL)
	}

	content, contentType, err := d.httpClient.Get(ctx, task.URL)
	if err != nil {
		result.Error = err
		return result, err
	}

	// Уточняем тип ресурса на основе Content-Type
	finalResourceType := d.refineResourceType(resourceType, contentType)

	switch finalResourceType {
	case domain.ResourceHTML:
		return d.processHTML(task, content)
	case domain.ResourceCSS:
		return d.processCSS(task, content)
	default:
		return d.processBinary(task, content)
	}
}

// determineResourceTypeWithHead пытается определить тип ресурса через HEAD запрос
func (d *WebDownloader) determineResourceTypeWithHead(ctx context.Context, url string) (domain.ResourceType, error) {
	contentType, err := d.httpClient.Head(ctx, url)
	if err != nil {
		return domain.ResourceOther, err
	}
	return d.determineResourceTypeByContentType(contentType), nil
}

// determineResourceTypeByContentType определяет тип по Content-Type
func (d *WebDownloader) determineResourceTypeByContentType(contentType string) domain.ResourceType {
	contentType = strings.ToLower(contentType)

	switch {
	case strings.Contains(contentType, "text/html"):
		return domain.ResourceHTML
	case strings.Contains(contentType, "text/css"):
		return domain.ResourceCSS
	case strings.Contains(contentType, "javascript"):
		return domain.ResourceJavaScript
	case strings.HasPrefix(contentType, "image/"):
		return domain.ResourceImage
	case strings.Contains(contentType, "font"):
		return domain.ResourceFont
	default:
		return domain.ResourceOther
	}
}

// determineResourceTypeByURL определяет тип по расширению файла
func (d *WebDownloader) determineResourceTypeByURL(url string) domain.ResourceType {
	url = strings.ToLower(url)

	switch {
	case strings.HasSuffix(url, ".html") || strings.HasSuffix(url, ".htm"):
		return domain.ResourceHTML
	case strings.HasSuffix(url, ".css"):
		return domain.ResourceCSS
	case strings.HasSuffix(url, ".js"):
		return domain.ResourceJavaScript
	case strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".jpg") ||
		strings.HasSuffix(url, ".jpeg") || strings.HasSuffix(url, ".gif") ||
		strings.HasSuffix(url, ".svg") || strings.HasSuffix(url, ".webp"):
		return domain.ResourceImage
	case strings.HasSuffix(url, ".woff") || strings.HasSuffix(url, ".woff2") ||
		strings.HasSuffix(url, ".ttf") || strings.HasSuffix(url, ".eot"):
		return domain.ResourceFont
	default:
		return domain.ResourceOther
	}
}

// refineResourceType уточняет тип ресурса на основе Content-Type
func (d *WebDownloader) refineResourceType(currentType domain.ResourceType, contentType string) domain.ResourceType {
	contentTypeBased := d.determineResourceTypeByContentType(contentType)

	// Если тип по Content-Type более специфичный, используем его
	if contentTypeBased != domain.ResourceOther {
		return contentTypeBased
	}
	return currentType
}

// processHTML обрабатывает HTML контент
// В processHTML метод добавьте логирование:
func (d *WebDownloader) processHTML(task domain.DownloadTask, content []byte) (domain.DownloadResult, error) {
	result := domain.DownloadResult{Task: task}

	// Извлекаем ссылки
	links, err := d.extractor.ExtractLinks(content, task.URL, "text/html")
	if err != nil {
		return result, err
	}
	result.Links = links

	log.Printf("Found %d links in %s", len(links), task.URL)
	for i, link := range links {
		log.Printf("  Link %d: %s", i+1, link)
	}

	// Перезаписываем ссылки
	rewrittenContent, err := d.linkRewriter.RewriteHTML(content, task.URL)
	if err != nil {
		return result, err
	}

	// Сохраняем файл
	localPath, err := d.pathResolver.URLToLocalPath(task.URL)
	if err != nil {
		return result, err
	}

	log.Printf("Saving to: %s", localPath)

	if err := d.fileManager.Save(localPath, rewrittenContent); err != nil {
		return result, err
	}

	result.FilePath = localPath
	result.Content = rewrittenContent
	return result, nil
}

// processCSS обрабатывает CSS контент
func (d *WebDownloader) processCSS(task domain.DownloadTask, content []byte) (domain.DownloadResult, error) {
	result := domain.DownloadResult{Task: task}

	links, err := d.extractor.ExtractLinks(content, task.URL, "text/css")
	if err != nil {
		return result, err
	}
	result.Links = links

	rewrittenContent, err := d.linkRewriter.RewriteCSS(content, task.URL)
	if err != nil {
		return result, err
	}

	localPath, err := d.pathResolver.URLToLocalPath(task.URL)
	if err != nil {
		return result, err
	}

	log.Printf("Saving to: %s", localPath)

	if err := d.fileManager.Save(localPath, rewrittenContent); err != nil {
		return result, err
	}

	result.FilePath = localPath
	result.Content = rewrittenContent
	return result, nil
}

// processBinary обрабатывает бинарный контент
func (d *WebDownloader) processBinary(task domain.DownloadTask, content []byte) (domain.DownloadResult, error) {
	result := domain.DownloadResult{Task: task}

	localPath, err := d.pathResolver.URLToLocalPath(task.URL)
	if err != nil {
		return result, err
	}

	log.Printf("Saving to: %s", localPath)

	if err := d.fileManager.Save(localPath, content); err != nil {
		return result, err
	}

	result.FilePath = localPath
	result.Content = content
	return result, nil
}
