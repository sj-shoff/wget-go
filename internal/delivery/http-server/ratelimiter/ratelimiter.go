package ratelimiter

import (
	"context"
	"time"
)

// TokenBucketRateLimiter реализует ограничение скорости через token bucket
type TokenBucketRateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
	rate   int
}

// New создает новый TokenBucketRateLimiter
func New(rate int) *TokenBucketRateLimiter {
	limiter := &TokenBucketRateLimiter{
		tokens: make(chan struct{}, rate),
		rate:   rate,
		ticker: time.NewTicker(time.Second / time.Duration(rate)),
	}

	go limiter.fillBucket()
	return limiter
}

// fillBucket непрерывно заполняет bucket токенами
func (rl *TokenBucketRateLimiter) fillBucket() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket полон
		}
	}
}

// Wait ожидает доступный токен
func (rl *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SetRate изменяет скорость ограничителя
func (rl *TokenBucketRateLimiter) SetRate(rate int) {
	if rate == rl.rate {
		return
	}

	rl.ticker.Stop()
	rl.ticker = time.NewTicker(time.Second / time.Duration(rate))
	rl.rate = rate

	// Очищаем и пересоздаем канал токенов
	close(rl.tokens)
	rl.tokens = make(chan struct{}, rate)
	go rl.fillBucket()
}
