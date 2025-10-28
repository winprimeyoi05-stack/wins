# Makefile for Telegram Premium Store Bot

.PHONY: build run clean test deps docker-build docker-run help

# Variables
BINARY_NAME=telegram-store-bot
DOCKER_IMAGE=telegram-store-bot
DOCKER_TAG=latest

# Default target
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
deps: ## Install Go dependencies
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

build: deps ## Build the application
	@echo "🔨 Building application..."
	@mkdir -p bin
	go build -o bin/$(BINARY_NAME) cmd/bot/main.go

run: build ## Run the application
	@echo "🚀 Starting bot..."
	./bin/$(BINARY_NAME)

dev: ## Run in development mode with auto-reload (requires air)
	@echo "🔄 Starting development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "❌ Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run 'make run' instead"; \
	fi

# Testing
test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Database
db-reset: ## Reset database (delete and recreate)
	@echo "🗄️ Resetting database..."
	rm -f store.db
	@echo "Database reset complete"

# Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run with Docker
	@echo "🐳 Running with Docker..."
	docker run --rm -it \
		--env-file .env \
		-v $(PWD)/data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-compose-up: ## Start with docker-compose
	@echo "🐳 Starting with docker-compose..."
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	@echo "🐳 Stopping docker-compose..."
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo "📋 Viewing logs..."
	docker-compose logs -f

# Deployment
deploy-build: ## Build for production deployment
	@echo "🏭 Building for production..."
	@mkdir -p bin
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/bot/main.go

# Utilities
clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean

format: ## Format Go code
	@echo "✨ Formatting code..."
	go fmt ./...

lint: ## Run linter
	@echo "🔍 Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "❌ golangci-lint not installed. Install with:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi

# Setup
setup: ## Setup development environment
	@echo "⚙️ Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "📄 Creating .env file from template..."; \
		cp .env.example .env; \
		echo "✅ .env file created. Please edit it with your configuration."; \
	else \
		echo "⚠️ .env file already exists"; \
	fi
	make deps

# Admin tools
admin: build ## Run admin CLI tools
	@echo "🔧 Starting admin tools..."
	go run cmd/admin/main.go

# QRIS tools
qris-test: ## Build and run QRIS test tool
	@echo "🔧 Building QRIS test tool..."
	@mkdir -p bin
	go build -o bin/qris-test cmd/qris-test/main.go
	@echo "✅ QRIS test tool built: bin/qris-test"
	@echo "Usage: ./bin/qris-test [upload|generate|status|test]"

qris-upload: qris-test ## Upload QRIS static image (requires image path)
	@if [ -z "$(IMAGE)" ]; then \
		echo "❌ Usage: make qris-upload IMAGE=path/to/qr.png"; \
	else \
		./bin/qris-test upload $(IMAGE); \
	fi

qris-status: qris-test ## Show QRIS configuration status
	@./bin/qris-test status

qris-generate: qris-test ## Generate test QRIS
	@./bin/qris-test test

# Monitoring
logs: ## Show application logs (if running with systemd)
	@echo "📋 Showing logs..."
	journalctl -u telegram-store-bot -f

status: ## Check application status (if running with systemd)
	@echo "📊 Checking status..."
	systemctl status telegram-store-bot

# Quick start
quick-start: setup build ## Quick start for new users
	@echo ""
	@echo "🎉 Setup complete!"
	@echo ""
	@echo "📋 Next steps:"
	@echo "1. Edit .env file with your bot token and admin IDs"
	@echo "2. Run 'make run' to start the bot"
	@echo "3. Setup QRIS dinamis dengan /qrissetup di Telegram"
	@echo ""
	@echo "🤖 To get bot token:"
	@echo "   - Chat with @BotFather on Telegram"
	@echo "   - Send /newbot and follow instructions"
	@echo ""
	@echo "🆔 To get your user ID:"
	@echo "   - Chat with @userinfobot on Telegram"
	@echo ""
	@echo "💳 To setup QRIS:"
	@echo "   - Start bot with 'make run'"
	@echo "   - Use /qrissetup command in Telegram"
	@echo "   - Upload static QR from your bank/e-wallet"
	@echo ""

# Release
release: clean test build ## Prepare release build
	@echo "📦 Preparing release..."
	@mkdir -p release
	@cp bin/$(BINARY_NAME) release/
	@cp .env.example release/
	@cp README.md release/
	@echo "✅ Release prepared in ./release/"

# Install system service (Linux only)
install-service: build ## Install as systemd service (Linux only)
	@echo "📥 Installing systemd service..."
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	@sudo cp scripts/telegram-store-bot.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable telegram-store-bot
	@echo "✅ Service installed. Configure .env and start with: sudo systemctl start telegram-store-bot"

# Backup
backup: ## Backup database and configuration
	@echo "💾 Creating backup..."
	@mkdir -p backups
	@cp store.db backups/store-$(shell date +%Y%m%d_%H%M%S).db 2>/dev/null || echo "No database found"
	@cp .env backups/env-$(shell date +%Y%m%d_%H%M%S).backup 2>/dev/null || echo "No .env found"
	@echo "✅ Backup created in ./backups/"