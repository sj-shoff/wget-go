# WGET-GO

Консольная утилита для рекурсивного скачивания веб-сайтов, написанная на Go.

## Возможности

- Рекурсивное скачивание веб-страниц с указанием глубины обхода
- Многопоточная загрузка с настраиваемым количеством воркеров
- Ограничение скорости запросов (rate limiting)
- Поддержка robots.txt
- Перезапись ссылок в скачанных файлах для локального просмотра
- Сохранение структуры сайта в локальной файловой системе
- Настраиваемые таймауты запросов
- Кастомный User-Agent

## Особенности реализации

- Используется конкурентная модель с worker pool
- Поддержка graceful shutdown при получении сигналов OS
- Автоматическое создание необходимых директорий
- Интеллектуальное определение типов контента (HTML, CSS, бинарные файлы)
- Перезапись относительных ссылок для локальной навигации
- Подробное логирование процесса скачивания

## Установка

```bash
git clone https://github.com/sj-shoff/wget-go.git
```

## Сборка

### Сборка бинарного файла

```bash
make build
```

### Запуск без сборки

```bash
make run
```

## Использование

### Базовое использование

```bash
./wget-go -url https://httpbin.org -depth 2 -workers 5
```

### Все параметры

```bash
./wget-go -url https://httpbin.org \
    -depth 3 \
    -workers 10 \
    -rate-limit 5 \
    -timeout 30s \
    -output ./downloads \
    -user-agent "CustomBot/1.0" \
    -respect-robots true
```

### Параметры командной строки

- `-url` (обязательный) - URL для скачивания
- `-depth` - максимальная глубина рекурсии (по умолчанию: 1)
- `-workers` - количество параллельных воркеров (по умолчанию: 5)
- `-rate-limit` - максимальное количество запросов в секунду (по умолчанию: 10)
- `-timeout` - таймаут для HTTP запросов (по умолчанию: 30s)
- `-output` - директория для сохранения файлов (по умолчанию: ./download)
- `-user-agent` - User-Agent для HTTP запросов (по умолчанию: Wget-Go/1.0)
- `-respect-robots` - соблюдать правила robots.txt (по умолчанию: true)

## Примеры

### Скачивание сайта с ограничением скорости

```bash
./wget-go -url https://httpbin.org -depth 2 -workers 3 -rate-limit 2 -output ./my_site
```

### Быстрое скачивание без ограничений

```bash
./wget-go -url https://httpbin.org -depth 1 -workers 20 -rate-limit 50
```

### Скачивание с кастомными настройками

```bash
./wget-go -url https://example.com \
    -depth 3 \
    -workers 8 \
    -rate-limit 3 \
    -timeout 60s \
    -user-agent "MyCrawler/1.0" \
    -respect-robots false
```

## Структура проекта

```
wget-go/
├── cmd/
│   └── wget/
│       └── main.go                 # Точка входа приложения
├── internal/
│   ├── app/
│   │   └── app.go                  # Composition Root (сборка всех зависимостей)
│   ├── config/
│   │   ├── config.go               # Загрузка конфига
│   │   └── flag_parser/
│   │       └── flagparser.go       # Парсинг аргументов командной строки
│   ├── delivery/
│   │   └── http-server/
│   │       ├── client/
│   │       │   └── client.go       # HTTP клиент
│   │       ├── ratelimiter/
│   │       │   └── ratelimiter.go  # Ограничитель запросов
│   │       ├── robots/
│   │       │   └── robots.go       # Проверка robots.txt
│   │       └── http.go             # HTTP интерфейсы
│   ├── domain/
│   │   └── types.go                # Доменные типы и структуры
│   ├── service/
│   │   ├── downloader/
│   │   │   └── downloader.go       # Сервис загрузки контента
│   │   ├── extractor/
│   │   │   └── extractor.go        # Извлечение ссылок из контента
│   │   ├── html_parser/
│   │   │   └── html_parser.go      # Парсинг HTML
│   │   ├── scheduler/
│   │   │   └── scheduler.go        # Планировщик задач загрузки
│   │   └── service.go              # Интерфейсы сервисов
│   └── storage/
│       ├── file_manager/
│       │   └── file_manager.go     # Управление файловой системой
│       ├── link_rewriter/
│       │   └── link_rewriter.go    # Перезапись ссылок
│       ├── path_resolver/
│       │   └── path_resolver.go    # Разрешение путей
│       └── storage.go              # Интерфейсы хранилищ
├── pkg/
│   ├── concurrency/
│   │   ├── queue.go                # Потокобезопасная очередь
│   │   ├── set.go                  # Потокобезопасное множество
│   │   └── worker_pool.go          # Пул воркеров
│   └── utils/
│       └── url.go                  # Утилиты для работы с URL
├── go.mod
└── go.sum
```