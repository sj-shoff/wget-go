package http_server

import (
	"context"
)

// Client определяет контракт HTTP клиента
type Client interface {
	Get(ctx context.Context, url string) ([]byte, string, error)
	Head(ctx context.Context, url string) (string, error)
}

// RateLimiter ограничивает частоту запросов
type RateLimiter interface {
	Wait(ctx context.Context) error
	SetRate(rate int)
}

// RobotsChecker проверяет robots.txt
//
// robots.txt — это текстовый файл, который веб-мастера размещают в
// корневом каталоге своего веб-сайта (например, https://example.com/robots.txt).
//
// Его основное назначение — информировать автоматизированные программы-краулеры
// (роботов, таких как поисковые системы или, в нашем случае, наш wget), какие
// части сайта им следует (или не следует) сканировать.
type RobotsChecker interface {
	IsAllowed(url string) bool
	SetUserAgent(ua string)
}
