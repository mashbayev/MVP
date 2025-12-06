.PHONY: help build run dev test clean

help:
	@echo "WhatsApp Analytics MVP - Development Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make dev              - Run server with development environment"
	@echo "  make build            - Build the project"
	@echo "  make run              - Run the built binary"
	@echo "  make test             - Run tests"
	@echo "  make test-telegram    - Test Telegram sender (requires TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID)"
	@echo "  make clean            - Remove build artifacts"
	@echo "  make lint             - Run linters"

build:
	@echo "Building project..."
	go build -o main ./cmd/main.go

run: build
	@echo "Running server..."
	./main

dev:
	@echo "Starting development server..."
	@./run-dev.sh

test:
	@echo "Running tests..."
	go test -v ./...

test-telegram:
	@echo "Testing Telegram sender..."
	@if [ -z "$(TELEGRAM_BOT_TOKEN)" ] || [ -z "$(TELEGRAM_CHAT_ID)" ]; then \
		echo "Error: TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID environment variables required"; \
		exit 1; \
	fi
	TELEGRAM_BOT_TOKEN=$(TELEGRAM_BOT_TOKEN) TELEGRAM_CHAT_ID=$(TELEGRAM_CHAT_ID) \
		go run -run TestTelegram ./tools

clean:
	@echo "Cleaning..."
	rm -f main
	go clean

lint:
	@echo "Linting..."
	golangci-lint run ./...
