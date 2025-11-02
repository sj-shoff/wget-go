package ratelimiter

import (
	"context"
	"sync"
	"time"
)

// TokenBucketRateLimiter реализует ограничение скорости с использованием алгоритма "бакет с токенами"
type TokenBucketRateLimiter struct {
	mu      sync.Mutex         // Мьютекс для защиты доступа к общим данным
	tokens  chan struct{}      // Канал, представляющий бакет с токенами
	ticker  *time.Ticker       // Тикер для периодического добавления токенов
	rate    int                // Количество токенов, добавляемых в секунду
	ctx     context.Context    // Контекст для управления жизненным циклом ограничителя
	cancel  context.CancelFunc // Функция отмены для контекста
	stopped bool               // Флаг, указывающий, остановлен ли ограничитель
}

// New создает новый TokenBucketRateLimiter
func New(rate int) *TokenBucketRateLimiter {
	// Если заданная скорость некорректна (меньше или равна нулю), устанавливаем минимальную скорость
	if rate <= 0 {
		rate = 1
	}

	// Создаем контекст с отменой для управления жизненным циклом ограничителя
	ctx, cancel := context.WithCancel(context.Background())

	limiter := &TokenBucketRateLimiter{
		tokens:  make(chan struct{}, rate),                         // Создаем канал с буфером, равным скорости
		rate:    rate,                                              // Сохраняем заданную скорость
		ticker:  time.NewTicker(time.Second / time.Duration(rate)), // Создаем тикер, срабатывающий rate раз в секунду
		ctx:     ctx,                                               // Сохраняем контекст
		cancel:  cancel,                                            // Сохраняем функцию отмены
		stopped: false,                                             // Изначально ограничитель не остановлен
	}

	// Запускаем горутину, которая будет непрерывно заполнять бакет токенами
	go limiter.fillBucket()
	return limiter
}

// fillBucket непрерывно заполняет бакет токенами
func (rl *TokenBucketRateLimiter) fillBucket() {
	// Бесконечный цикл, пока ограничитель не остановлен или контекст не отменен
	for {
		select {
		case <-rl.ticker.C: // Когда срабатывает тикер
			rl.mu.Lock()     // Блокируем мьютекс для безопасного доступа к полям
			if !rl.stopped { // Если ограничитель не остановлен
				select {
				case rl.tokens <- struct{}{}: // Пытаемся добавить токен в канал (бакет)
				default:
					// Бакет уже полон, ничего не делаем
				}
			}
			rl.mu.Unlock() // Разблокируем мьютекс
		case <-rl.ctx.Done(): // Если контекст отменен (например, при вызове Stop())
			return // Завершаем горутину
		}
	}
}

// Wait ожидает доступный токен
func (rl *TokenBucketRateLimiter) Wait(ctx context.Context) error {
	// Используем select для ожидания либо доступного токена, либо отмены контекста
	select {
	case <-rl.tokens: // Если удалось получить токен из канала
		return nil // Операция разрешена
	case <-ctx.Done(): // Если внешний контекст пользователя отменен (таймаут, явная отмена)
		return ctx.Err() // Возвращаем ошибку внешнего контекста
	case <-rl.ctx.Done(): // Если внутренний контекст ограничителя отменен (например, при вызове Stop())
		return context.Canceled // Возвращаем ошибку отмены внутреннего контекста
	}
}

// SetRate изменяет скорость ограничителя
func (rl *TokenBucketRateLimiter) SetRate(rate int) {
	// Если заданная скорость некорректна, устанавливаем минимальную
	if rate <= 0 {
		rate = 1
	}

	rl.mu.Lock()         // Блокируем мьютекс для безопасного изменения состояния
	defer rl.mu.Unlock() // Гарантируем разблокировку мьютекса

	// Если новая скорость совпадает с текущей, или ограничитель уже остановлен, ничего не делаем
	if rate == rl.rate || rl.stopped {
		return
	}

	// Останавливаем старый тикер
	rl.ticker.Stop()
	// Создаем новый тикер с новой скоростью
	rl.ticker = time.NewTicker(time.Second / time.Duration(rate))
	// Обновляем значение скорости
	rl.rate = rate

	// Создаем новый канал токенов с новой емкостью
	newTokens := make(chan struct{}, rate)

	// Переносим существующие токены (не более новой емкости)
	transfer := true            // Флаг для контроля переноса
	for i := 0; i < rate; i++ { // Итерируемся до новой скорости (максимальная емкость нового бакета)
		select {
		case token := <-rl.tokens: // Пытаемся неблокирующе извлечь токен из старого бакета
			select {
			case newTokens <- token: // Пытаемся неблокирующе записать токен в новый бакет
			default:
				// Новый бакет заполнен, прекращаем перенос
				transfer = false
			}
		default:
			// Больше нет токенов для переноса из старого бакета
			transfer = false
		}

		// Если перенос прекращен, выходим из цикла
		if !transfer {
			break
		}
	}

	// Атомарно заменяем старый канал токенов новым
	rl.tokens = newTokens
}

// Stop останавливает ограничитель
func (rl *TokenBucketRateLimiter) Stop() {
	rl.mu.Lock()         // Блокируем мьютекс для безопасного изменения состояния
	defer rl.mu.Unlock() // Гарантируем разблокировку мьютекса

	// Если ограничитель уже остановлен, ничего не делаем
	if rl.stopped {
		return
	}

	rl.stopped = true // Устанавливаем флаг остановки
	rl.ticker.Stop()  // Останавливаем тикер
	rl.cancel()       // Отменяем контекст, сигнализируя всем ожидающим горутинам завершиться
	close(rl.tokens)  // Закрываем канал токенов, чтобы все, кто ожидает в Wait, вернулись с nil
}
