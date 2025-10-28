package config

import (
	"fmt"
	"time"
	"wget-go/internal/domain"
)

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *domain.Config {
	return &domain.Config{
		OutputDir:     "./download",
		MaxDepth:      1,
		Workers:       5,
		RateLimit:     10,
		UserAgent:     "Wget-Go/1.0",
		Timeout:       30 * time.Second,
		RespectRobots: true,
	}
}

// Validate проверяет корректность конфигурации
func Validate(cfg *domain.Config) error {
	if cfg.URL == "" {
		return fmt.Errorf("URL is required")
	}
	if cfg.MaxDepth < 0 {
		return fmt.Errorf("depth cannot be negative")
	}
	if cfg.Workers < 1 {
		return fmt.Errorf("workers must be at least 1")
	}
	if cfg.RateLimit < 1 {
		return fmt.Errorf("rate limit must be at least 1")
	}
	return nil
}
