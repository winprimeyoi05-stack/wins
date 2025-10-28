# Makefile for Telegram Premium Store Bot

.PHONY: build run clean test deps docker-build docker-run help

# Variables
BINARY_NAME=telegram-store-bot
DOCKER_IMAGE=telegram-store-bot
DOCKER_TAG=latest

# Detect OS and set binary extension
ifeq ($(OS),Windows_NT)
	BINARY_EXT=.exe
	RM=del /Q /F
	RMDIR=rmdir /S /Q
	MKDIR=if not exist
	PATH_SEP=\\
	DETECTED_OS=Windows
	SHELL=cmd
	ECHO=echo
	NULL_DEVICE=NUL
else
	BINARY_EXT=
	RM=rm -f
	RMDIR=rm -rf
	MKDIR=mkdir -p
	PATH_SEP=/
	DETECTED_OS=$(shell uname -s)
	ECHO=echo
	NULL_DEVICE=/dev/null
endif

# Default target
help: ## Show this help message
ifeq ($(OS),Windows_NT)
	@echo Available commands:
	@echo.
	@echo Build and Run:
	@echo   build              - Build the application
	@echo   run                - Run the application
	@echo   quick-start        - Quick start for new users
	@echo.
	@echo Development:
	@echo   deps               - Install Go dependencies
	@echo   dev                - Run in development mode
	@echo   test               - Run tests
	@echo   format             - Format Go code
	@echo   lint               - Run linter
	@echo   clean              - Clean build artifacts
	@echo.
	@echo QRIS Tools:
	@echo   qris-test          - Build QRIS test tool
	@echo   qris-status        - Show QRIS configuration status
	@echo   qris-upload        - Upload QRIS static image
	@echo.
	@echo Windows-Specific:
	@echo   windows-check      - Check Windows build environment
	@echo   windows-build      - Build using Windows build script
	@echo   windows-deps       - Install Windows development tools
	@echo.
	@echo Other:
	@echo   backup             - Backup database and configuration
	@echo   db-reset           - Reset database
	@echo   admin              - Run admin CLI tools
	@echo   help               - Show this help message
	@echo.
else
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
endif

# Development
deps: ## Install Go dependencies
	@echo Installing dependencies...
	go mod download
	go mod tidy

build: deps ## Build the application
	@echo Building application...
ifeq ($(OS),Windows_NT)
	@if not exist bin mkdir bin
else
	@mkdir -p bin
endif
	go build -o bin$(PATH_SEP)$(BINARY_NAME)$(BINARY_EXT) cmd$(PATH_SEP)bot$(PATH_SEP)main.go

run: build ## Run the application
	@echo Starting bot...
ifeq ($(OS),Windows_NT)
	@bin\$(BINARY_NAME)$(BINARY_EXT)
else
	@bin/$(BINARY_NAME)$(BINARY_EXT)
endif

dev: ## Run in development mode with auto-reload (requires air)
	@echo Starting development mode...
ifeq ($(OS),Windows_NT)
	@where air >NUL 2>NUL && air || (echo Air not installed. Install with: go install github.com/cosmtrek/air@latest && echo Or run 'make run' instead)
else
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Or run 'make run' instead"; \
	fi
endif

# Testing
test: ## Run tests
	@echo Running tests...
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo Running tests with coverage...
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo Coverage report generated: coverage.html

# Database
db-reset: ## Reset database (delete and recreate)
	@echo Resetting database...
ifeq ($(OS),Windows_NT)
	@if exist store.db del /Q /F store.db
else
	@rm -f store.db
endif
	@echo Database reset complete

# Docker
docker-build: ## Build Docker image
	@echo Building Docker image...
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run with Docker
	@echo Running with Docker...
ifeq ($(OS),Windows_NT)
	docker run --rm -it --env-file .env -v %CD%/data:/app/data $(DOCKER_IMAGE):$(DOCKER_TAG)
else
	docker run --rm -it --env-file .env -v $(PWD)/data:/app/data $(DOCKER_IMAGE):$(DOCKER_TAG)
endif

docker-compose-up: ## Start with docker-compose
	@echo Starting with docker-compose...
	docker-compose up -d

docker-compose-down: ## Stop docker-compose
	@echo Stopping docker-compose...
	docker-compose down

docker-compose-logs: ## View docker-compose logs
	@echo Viewing logs...
	docker-compose logs -f

# Deployment
deploy-build: ## Build for production deployment
	@echo Building for production...
ifeq ($(OS),Windows_NT)
	@if not exist bin mkdir bin
	set CGO_ENABLED=1&& set GOOS=linux&& go build -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/bot/main.go
else
	@mkdir -p bin
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/$(BINARY_NAME) cmd/bot/main.go
endif

# Utilities
clean: ## Clean build artifacts
	@echo Cleaning...
ifeq ($(OS),Windows_NT)
	@if exist bin rmdir /S /Q bin
	@if exist coverage.out del /Q /F coverage.out
	@if exist coverage.html del /Q /F coverage.html
else
	@rm -rf bin/
	@rm -f coverage.out coverage.html
endif
	go clean

format: ## Format Go code
	@echo Formatting code...
	go fmt ./...

lint: ## Run linter
	@echo Running linter...
ifeq ($(OS),Windows_NT)
	@where golangci-lint >NUL 2>NUL && golangci-lint run || echo golangci-lint not installed. Install from https://golangci-lint.run/usage/install/
else
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.54.2"; \
	fi
endif

# Setup
setup: ## Setup development environment
	@echo Setting up development environment...
ifeq ($(OS),Windows_NT)
	@if not exist .env (echo Creating .env file from template... && copy .env.example .env && echo .env file created. Please edit it with your configuration.) else (echo .env file already exists)
else
	@if [ ! -f .env ]; then \
		echo "Creating .env file from template..."; \
		cp .env.example .env; \
		echo ".env file created. Please edit it with your configuration."; \
	else \
		echo ".env file already exists"; \
	fi
endif
	$(MAKE) deps

# Admin tools
admin: ## Run admin CLI tools
	@echo Starting admin tools...
ifeq ($(OS),Windows_NT)
	go run cmd\admin\main.go
else
	go run cmd/admin/main.go
endif

# QRIS tools
qris-test: ## Build and run QRIS test tool
	@echo Building QRIS test tool...
ifeq ($(OS),Windows_NT)
	@if not exist bin mkdir bin
	go build -o bin\qris-test$(BINARY_EXT) cmd\qris-test\main.go
	@echo QRIS test tool built: bin\qris-test$(BINARY_EXT)
	@echo Usage: bin\qris-test$(BINARY_EXT) [upload^|generate^|status^|test]
else
	@mkdir -p bin
	go build -o bin/qris-test$(BINARY_EXT) cmd/qris-test/main.go
	@echo "QRIS test tool built: bin/qris-test$(BINARY_EXT)"
	@echo "Usage: bin/qris-test$(BINARY_EXT) [upload|generate|status|test]"
endif

qris-upload: qris-test ## Upload QRIS static image (requires image path)
ifeq ($(OS),Windows_NT)
	@if "$(IMAGE)"=="" (echo Usage: make qris-upload IMAGE=path\to\qr.png) else (bin\qris-test$(BINARY_EXT) upload $(IMAGE))
else
	@if [ -z "$(IMAGE)" ]; then \
		echo "Usage: make qris-upload IMAGE=path/to/qr.png"; \
	else \
		bin/qris-test$(BINARY_EXT) upload $(IMAGE); \
	fi
endif

qris-status: qris-test ## Show QRIS configuration status
ifeq ($(OS),Windows_NT)
	@bin\qris-test$(BINARY_EXT) status
else
	@bin/qris-test$(BINARY_EXT) status
endif

qris-generate: qris-test ## Generate test QRIS
ifeq ($(OS),Windows_NT)
	@bin\qris-test$(BINARY_EXT) test
else
	@bin/qris-test$(BINARY_EXT) test
endif

# Monitoring
logs: ## Show application logs (Linux systemd only)
ifeq ($(OS),Windows_NT)
	@echo Logs command is only available on Linux systems with systemd
else
	@echo Showing logs...
	journalctl -u telegram-store-bot -f
endif

status: ## Check application status (Linux systemd only)
ifeq ($(OS),Windows_NT)
	@echo Status command is only available on Linux systems with systemd
else
	@echo Checking status...
	systemctl status telegram-store-bot
endif

# Quick start
quick-start: setup build ## Quick start for new users
	@echo.
	@echo Setup complete!
	@echo.
	@echo Next steps:
	@echo 1. Edit .env file with your bot token and admin IDs
	@echo 2. Run 'make run' to start the bot
	@echo 3. Setup QRIS dinamis dengan /qrissetup di Telegram
	@echo.
	@echo To get bot token:
	@echo    - Chat with @BotFather on Telegram
	@echo    - Send /newbot and follow instructions
	@echo.
	@echo To get your user ID:
	@echo    - Chat with @userinfobot on Telegram
	@echo.
	@echo To setup QRIS:
	@echo    - Start bot with 'make run'
	@echo    - Use /qrissetup command in Telegram
	@echo    - Upload static QR from your bank/e-wallet
	@echo.

# Release
release: clean test build ## Prepare release build
	@echo Preparing release...
ifeq ($(OS),Windows_NT)
	@if not exist release mkdir release
	@copy bin\$(BINARY_NAME)$(BINARY_EXT) release\
	@copy .env.example release\
	@copy README.md release\
else
	@mkdir -p release
	@cp bin/$(BINARY_NAME)$(BINARY_EXT) release/
	@cp .env.example release/
	@cp README.md release/
endif
	@echo Release prepared in ./release/

# Install system service (Linux only)
install-service: build ## Install as systemd service (Linux only)
ifeq ($(OS),Windows_NT)
	@echo System service installation is only available on Linux
	@echo For Windows, use Task Scheduler or NSSM to run as a service
else
	@echo Installing systemd service...
	@sudo cp bin/$(BINARY_NAME) /usr/local/bin/
	@sudo cp scripts/telegram-store-bot.service /etc/systemd/system/
	@sudo systemctl daemon-reload
	@sudo systemctl enable telegram-store-bot
	@echo Service installed. Configure .env and start with: sudo systemctl start telegram-store-bot
endif

# Backup
backup: ## Backup database and configuration
	@echo Creating backup...
ifeq ($(OS),Windows_NT)
	@if not exist backups mkdir backups
	@if exist store.db (copy store.db backups\store-%DATE:~-4%%DATE:~-10,2%%DATE:~-7,2%_%TIME:~0,2%%TIME:~3,2%%TIME:~6,2%.db >NUL) else (echo No database found)
	@if exist .env (copy .env backups\env-%DATE:~-4%%DATE:~-10,2%%DATE:~-7,2%_%TIME:~0,2%%TIME:~3,2%%TIME:~6,2%.backup >NUL) else (echo No .env found)
else
	@mkdir -p backups
	@cp store.db backups/store-$(shell date +%Y%m%d_%H%M%S).db 2>/dev/null || echo "No database found"
	@cp .env backups/env-$(shell date +%Y%m%d_%H%M%S).backup 2>/dev/null || echo "No .env found"
endif
	@echo Backup created in ./backups/

# Windows-specific targets
windows-check: ## Check Windows build environment
ifeq ($(OS),Windows_NT)
	@echo Checking Windows build environment...
	@echo.
	@echo Checking Go installation...
	@where go >NUL 2>NUL && (go version && echo [OK] Go is installed) || (echo [ERROR] Go is not installed. Install from https://go.dev/dl/ && exit /b 1)
	@echo.
	@echo Checking GCC installation (required for SQLite)...
	@where gcc >NUL 2>NUL && (gcc --version && echo [OK] GCC is installed) || (echo [WARNING] GCC is not installed. SQLite requires CGO and GCC. && echo Install TDM-GCC from https://jmeubank.github.io/tdm-gcc/ && exit /b 1)
	@echo.
	@echo Checking Make installation...
	@where make >NUL 2>NUL && (make --version && echo [OK] Make is installed) || (echo [WARNING] Make not found but you are running it)
	@echo.
	@echo [OK] Windows build environment is ready!
else
	@echo This target is only for Windows systems
endif

windows-build: ## Build for Windows (using build scripts)
ifeq ($(OS),Windows_NT)
	@echo Building using Windows build script...
	@if exist build-windows.bat (call build-windows.bat) else (echo ERROR: build-windows.bat not found && exit /b 1)
else
	@echo This target is only for Windows systems
	@echo Use 'make build' instead
endif

windows-deps: ## Install common Windows development tools
ifeq ($(OS),Windows_NT)
	@echo Installing Windows development dependencies...
	@echo.
	@echo Installing Go tools...
	go install github.com/cosmtrek/air@latest
	@echo.
	@echo Done! To install GCC, visit: https://jmeubank.github.io/tdm-gcc/
else
	@echo This target is only for Windows systems
endif

.DEFAULT_GOAL := help