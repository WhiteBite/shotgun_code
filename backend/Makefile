# Makefile для ARK CLI

.PHONY: build clean test help

# Переменные
BINARY_NAME=ark
BUILD_DIR=build
VERSION=1.0.0
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

# Основные цели
build: clean
	@echo "$(GREEN)Building ARK CLI...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/ark
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-all: clean
	@echo "$(GREEN)Building ARK CLI for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/ark
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/ark
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/ark
	@echo "$(GREEN)Cross-platform build completed$(NC)"

clean:
	@echo "$(GREEN)Cleaning build directory...$(NC)"
	@rm -rf $(BUILD_DIR)

test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test ./...

install: build
	@echo "$(GREEN)Installing ARK CLI...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)ARK CLI installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

# Цели для разработки
dev-build:
	@echo "$(GREEN)Building for development...$(NC)"
	go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev ./cmd/ark

lint:
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

format:
	@echo "$(GREEN)Formatting code...$(NC)"
	go fmt ./...

# Цели для CLI команд
test-index:
	@echo "$(GREEN)Testing index command...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) index --help

test-solve:
	@echo "$(GREEN)Testing solve command...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) solve --help

test-result:
	@echo "$(GREEN)Testing result command...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) result --help

# Справка
help:
	@echo "$(GREEN)ARK CLI Makefile$(NC)"
	@echo ""
	@echo "Available targets:"
	@echo "  build       - Build ARK CLI for current platform"
	@echo "  build-all   - Build ARK CLI for all platforms (Linux, Windows, macOS)"
	@echo "  clean       - Clean build directory"
	@echo "  test        - Run tests"
	@echo "  install     - Install ARK CLI to /usr/local/bin"
	@echo "  dev-build   - Build with race detection for development"
	@echo "  lint        - Run linter"
	@echo "  format      - Format code"
	@echo "  test-index  - Test index command"
	@echo "  test-solve  - Test solve command"
	@echo "  test-result - Test result command"
	@echo "  help        - Show this help"
