# Makefile для WGET-GO

# Переменные
BINARY_NAME=wget-go
BUILD_DIR=bin
SOURCE_DIR=./cmd/wget

# Сборка бинарного файла
build:
	@echo "Сборка $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(SOURCE_DIR)
	@echo "Бинарный файл создан: $(BUILD_DIR)/$(BINARY_NAME)"

# Запуск без сборки
run:
	@echo "Запуск приложения..."
	@go run $(SOURCE_DIR) -url https://httpbin.org -depth 1 -workers 5
