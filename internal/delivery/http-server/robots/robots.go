package robots

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"
	httpserver "wget-go/internal/delivery/http-server"
)

// RobotsTxt представляет правила из robots.txt
type RobotsTxt struct {
	rules map[string][]string // user-agent -> disallow paths
}

// RobotsCheckerImpl реализация RobotsChecker
type RobotsCheckerImpl struct {
	client    httpserver.Client
	userAgent string
	cache     map[string]*RobotsTxt
	mu        sync.RWMutex
}

// New создает новый проверщик robots.txt
func New(client httpserver.Client) *RobotsCheckerImpl {
	return &RobotsCheckerImpl{
		client: client,
		cache:  make(map[string]*RobotsTxt),
	}
}

// SetUserAgent устанавливает User-Agent для проверки
func (r *RobotsCheckerImpl) SetUserAgent(ua string) {
	r.userAgent = ua
}

func (r *RobotsCheckerImpl) IsAllowed(rawUrl string) bool {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return true // При ошибке парсинга разрешаем доступ
	}

	domain := parsedUrl.Host
	path := parsedUrl.Path

	// Получаем robots.txt для домена
	robotsTxt, err := r.getRobotsTxt(domain)
	if err != nil {
		return true // При ошибке загрузки разрешаем доступ
	}

	return robotsTxt.IsAllowed(r.userAgent, path)
}

func (r *RobotsCheckerImpl) getRobotsTxt(domain string) (*RobotsTxt, error) {
	r.mu.RLock()
	robotsTxt, exists := r.cache[domain]
	r.mu.RUnlock()

	if exists {
		return robotsTxt, nil
	}

	// Загружаем robots.txt
	robotsURL := fmt.Sprintf("https://%s/robots.txt", domain)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	content, _, err := r.client.Get(ctx, robotsURL)
	if err != nil {
		// Если не удалось загрузить, создаем пустой robots.txt
		robotsTxt = &RobotsTxt{rules: make(map[string][]string)}
	} else {
		// Парсим robots.txt
		robotsTxt = parseRobotsTxt(content)
	}

	r.mu.Lock()
	r.cache[domain] = robotsTxt
	r.mu.Unlock()

	return robotsTxt, nil
}

// parseRobotsTxt парсит содержимое robots.txt
func parseRobotsTxt(content []byte) *RobotsTxt {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	rules := make(map[string][]string)
	var currentAgent string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пропускаем комментарии и пустые строки
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "user-agent":
			currentAgent = value
			if _, exists := rules[currentAgent]; !exists {
				rules[currentAgent] = []string{}
			}
		case "disallow":
			if currentAgent != "" && value != "" {
				rules[currentAgent] = append(rules[currentAgent], value)
			}
		case "allow":
			// Можно добавить поддержку allow директив при необходимости
		}
	}

	return &RobotsTxt{rules: rules}
}

// IsAllowed проверяет, разрешен ли путь для указанного User-Agent
func (r *RobotsTxt) IsAllowed(userAgent string, path string) bool {
	// Проверяем правила для конкретного User-Agent
	if disallows, exists := r.rules[userAgent]; exists {
		if r.isPathDisallowed(path, disallows) {
			return false
		}
	}

	// Проверяем правила для агентов (*)
	if disallows, exists := r.rules["*"]; exists {
		if r.isPathDisallowed(path, disallows) {
			return false
		}
	}

	return true
}

// isPathDisallowed проверяет, является ли путь запрещенным
func (r *RobotsTxt) isPathDisallowed(path string, disallows []string) bool {
	for _, disallow := range disallows {
		if disallow == "" {
			continue
		}

		if strings.HasPrefix(path, disallow) {
			return true
		}
	}
	return false
}
