package config

import (
	"fmt"
	"log"
	"time"
)

// Config содержит конфигурацию приложения
type Config struct {
	URL           string
	OutputDir     string
	MaxDepth      int
	Workers       int
	RateLimit     int
	UserAgent     string
	Timeout       time.Duration
	RespectRobots bool
}

func MustLoad() *Config {
	cfg := *defaultConfig()

	// Парсинг флагов
	parsedCfg := parse(cfg)

	// Валидация конфигурации
	if err := validate(parsedCfg); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	return parsedCfg
}

// DefaultConfig возвращает конфигурацию по умолчанию
func defaultConfig() *Config {
	return &Config{
		OutputDir:     "./download",
		MaxDepth:      1,
		Workers:       5,
		RateLimit:     10,
		UserAgent:     "Wget-Go/1.0",
		Timeout:       30 * time.Second,
		RespectRobots: true,
	}
}

// validate проверяет корректность конфигурации
func validate(cfg *Config) error {
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
