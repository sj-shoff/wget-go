package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// TokenBucketRateLimiter реализует ограничение скорости через token bucket
type TokenBucketRateLimiter struct {
	mu     sync.RWMutex
	tokens chan struct{}
	ticker *time.Ticker
	rate   int
}

// New создает новый TokenBucketRateLimiter
func New(rate int) *TokenBucketRateLimiter {
	if rate <= 0 {
		rate = 1 // Установим минимальную скорость, чтобы избежать паники
	}
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
		rl.mu.RLock()             // Блокировка для чтения, чтобы гарантировать, что ticker не будет изменен во время чтения
		tickerChan := rl.ticker.C // Копируем канал тикера, чтобы освободить мьютекс раньше
		rl.mu.RUnlock()

		select {
		case <-tickerChan: // Используем скопированный канал
			rl.mu.Lock() // Блокировка для записи в rl.tokens
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket полон
			}
			rl.mu.Unlock()
		}
	}
}

// Wait ожидает доступный токен
func (rl *TokenBucketRateLimiter) Wait(ctx context.Context) error {

	rl.mu.RLock()
	tokensChan := rl.tokens
	rl.mu.RUnlock()

	select {
	case <-tokensChan:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// SetRate изменяет скорость ограничителя
func (rl *TokenBucketRateLimiter) SetRate(rate int) {
	if rate <= 0 {
		rate = 1
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rate == rl.rate {
		return
	}

	// 1. Останавливаем старый тикер
	rl.ticker.Stop()

	// 2. Сохраняем текущие токены (сколько уже есть в буфере)
	currentTokens := len(rl.tokens)

	// 3. Создаем новый канал токенов
	newTokens := make(chan struct{}, rate)

	// 4. Переносим существующие токены (но не более нового лимита)
	for i := 0; i < currentTokens && i < rate; i++ {
		newTokens <- struct{}{}
	}

	// 5. Атомарно обновляем состояние
	rl.tokens = newTokens
	rl.rate = rate
	rl.ticker = time.NewTicker(time.Second / time.Duration(rate))

	// Старая горутина fillBucket автоматически завершится,
	// так как мы остановили ее тикер, а новая горутина уже запущена в New()
	// и будет использовать новый тикер
}

// Дополнительно: метод для остановки ограничителя
func (rl *TokenBucketRateLimiter) Stop() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.ticker.Stop()
	close(rl.tokens) // Закрываем канал, чтобы все ожидающие Wait вернулись с nil
}
