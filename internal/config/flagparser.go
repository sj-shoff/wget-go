package config

import (
	"flag"
	"fmt"
	"os"
)

// Parse извлекает конфигурацию из флагов
func parse(cfg Config) *Config {
	flag.StringVar(&cfg.URL, "url", "", "URL to download (required)")
	flag.StringVar(&cfg.OutputDir, "output", "./download", "Output directory")
	flag.IntVar(&cfg.MaxDepth, "depth", cfg.MaxDepth, "Maximum recursion depth")
	flag.IntVar(&cfg.Workers, "workers", cfg.Workers, "Number of concurrent workers")
	flag.IntVar(&cfg.RateLimit, "rate-limit", cfg.RateLimit, "Maximum requests per second")
	flag.StringVar(&cfg.UserAgent, "user-agent", cfg.UserAgent, "User-Agent header")
	flag.DurationVar(&cfg.Timeout, "timeout", cfg.Timeout, "Request timeout")
	flag.BoolVar(&cfg.RespectRobots, "respect-robots", cfg.RespectRobots, "Respect robots.txt")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "\nExample:")
		fmt.Fprintln(flag.CommandLine.Output(), "  wget-go -url https://example.com -depth 2 -workers 10")
	}

	flag.Parse()

	return &cfg
}
