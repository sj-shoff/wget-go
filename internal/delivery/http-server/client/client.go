package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"wget-go/internal/config"
	httpserver "wget-go/internal/delivery/http-server"
)

// HTTPClient реализация HTTP-клиента
type HTTPClient struct {
	client        *http.Client
	userAgent     string
	rateLimiter   httpserver.RateLimiter
	robotsChecker httpserver.RobotsChecker
}

// New создает новый HTTP клиент
func New(cfg *config.Config, rateLimiter httpserver.RateLimiter, robotsChecker httpserver.RobotsChecker) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout:       cfg.Timeout,
			CheckRedirect: redirectPolicy,
		},
		userAgent:     cfg.UserAgent,
		rateLimiter:   rateLimiter,
		robotsChecker: robotsChecker,
	}
}

// redirectPolicy ограничивает количество редиректов
func redirectPolicy(req *http.Request, redir []*http.Request) error {
	if len(redir) >= 10 {
		return fmt.Errorf("stopped after 10 redirects")
	}
	return nil
}

// Get выполняет HTTP GET запрос
func (c *HTTPClient) Get(ctx context.Context, url string) ([]byte, string, error) {

	// проверка robots.txt если включено
	if c.robotsChecker != nil && !c.robotsChecker.IsAllowed(url) {
		return nil, "", fmt.Errorf("access denied by robots.txt")
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, "", fmt.Errorf("rate limiter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("execute request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read response body: %w", err)
	}

	return content, resp.Header.Get("Content-Type"), nil
}

// Head выполняет HTTP HEAD запрос
func (c *HTTPClient) Head(ctx context.Context, url string) (string, error) {
	// Проверяем robots.txt если включено
	if c.robotsChecker != nil && !c.robotsChecker.IsAllowed(url) {
		return "", fmt.Errorf("access disallowed by robots.txt: %s", url)
	}

	if err := c.rateLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("rate limiter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "HEAD", url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return resp.Header.Get("Content-Type"), nil
}

// setHeaders устанавливает стандартные заголовки
func (c *HTTPClient) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
}
