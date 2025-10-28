package downloader

import (
	http_server "wget-go/internal/http-server"
	"wget-go/internal/service"
	"wget-go/internal/storage"
)

// WebDownloader реализует сервис загрузки
type WebDownloader struct {
	httpClient   http_server.Client
	fileManager  storage.FileManager
	pathResolver storage.PathResolver
	linkRewriter storage.LinkRewriter
	extractor    service.Extractor
}
