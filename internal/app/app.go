package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"wget-go/internal/config"
	httpserver "wget-go/internal/delivery/http-server"

	"wget-go/internal/delivery/http-server/client"
	"wget-go/internal/delivery/http-server/ratelimiter"
	"wget-go/internal/delivery/http-server/robots"
	"wget-go/internal/service/downloader"
	"wget-go/internal/service/extractor"
	"wget-go/internal/service/html_parser"
	"wget-go/internal/service/scheduler"
	"wget-go/internal/storage/file_manager"
	"wget-go/internal/storage/link_rewriter"
	"wget-go/internal/storage/path_resolver"

	"wget-go/internal/config/flag_parser"
)

// Application основное приложение
type Application struct {
	config    *config.Config
	scheduler *scheduler.DownloadScheduler
}

// New создает и инициализирует приложение
func New() *Application {
	cfg := flag_parser.New().Parse()

	rateLimiter := ratelimiter.New(cfg.RateLimit)

	// Создаем robots checker если включено
	var robotsChecker httpserver.RobotsChecker
	if cfg.RespectRobots {
		// Создаем временный клиент для загрузки robots.txt
		tempClient := client.New(cfg, rateLimiter, nil)
		robotsChecker = robots.New(tempClient)
		robotsChecker.SetUserAgent(cfg.UserAgent)
	}

	httpClient := client.New(cfg, rateLimiter, robotsChecker)
	fileManager := file_manager.New()
	pathResolver := path_resolver.New(cfg.OutputDir)
	linkRewriter := link_rewriter.New(pathResolver)

	htmlParser := html_parser.New()
	linkExtractor := extractor.New(htmlParser)

	webDownloader := downloader.New(
		httpClient,
		fileManager,
		pathResolver,
		linkRewriter,
		linkExtractor,
	)

	downloadScheduler := scheduler.New(
		cfg,
		webDownloader,
		pathResolver,
	)

	return &Application{
		config:    cfg,
		scheduler: downloadScheduler,
	}
}

// Run запускает приложение
func (a *Application) Run() error {
	log.Printf("Wget-Go starting...")
	log.Printf("URL: %s", a.config.URL)
	log.Printf("Output directory: %s", a.config.OutputDir)
	log.Printf("Max depth: %d, Workers: %d, Rate limit: %d/sec",
		a.config.MaxDepth, a.config.Workers, a.config.RateLimit)
	log.Printf("Respect robots.txt: %v", a.config.RespectRobots)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.handleSignals(cancel)

	if err := a.scheduler.Start(ctx); err != nil {
		return err
	}

	log.Printf("Wget-Go finished successfully")
	return nil
}

// handleSignals обрабатывает сигналы OS для graceful shutdown
func (a *Application) handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Printf("Received shutdown signal, stopping gracefully...")
	cancel()
}
