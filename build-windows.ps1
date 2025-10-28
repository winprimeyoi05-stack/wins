#!/usr/bin/env pwsh
# PowerShell build script for Windows

Write-Host "Building Telegram Store Bot for Windows..." -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Host "ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from https://go.dev/dl/"
    exit 1
}

# Check if GCC is installed (required for SQLite)
if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
    Write-Host "WARNING: GCC is not installed" -ForegroundColor Yellow
    Write-Host "SQLite requires CGO and GCC. Please install one of:" -ForegroundColor Yellow
    Write-Host "  - TDM-GCC: https://jmeubank.github.io/tdm-gcc/" -ForegroundColor Yellow
    Write-Host "  - MinGW-w64: https://www.mingw-w64.org/" -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

# Clean old builds
Write-Host "Cleaning old builds..." -ForegroundColor Yellow
if (Test-Path "bin") {
    Remove-Item -Recurse -Force "bin"
}
if (Test-Path "telegram-store-bot.exe") {
    Remove-Item "telegram-store-bot.exe"
}

# Create bin directory
New-Item -ItemType Directory -Force -Path "bin" | Out-Null

# Download dependencies
Write-Host "Installing dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy

# Build the application
Write-Host "Building application..." -ForegroundColor Yellow
$env:CGO_ENABLED = "1"
go build -o bin\telegram-store-bot.exe cmd\bot\main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================"  -ForegroundColor Green
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "========================================"  -ForegroundColor Green
    Write-Host ""
    Write-Host "Binary created: bin\telegram-store-bot.exe" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "To run the bot:" -ForegroundColor Yellow
    Write-Host "  .\bin\telegram-store-bot.exe" -ForegroundColor White
    Write-Host ""
    Write-Host "Or use:" -ForegroundColor Yellow
    Write-Host "  make run" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "========================================"  -ForegroundColor Red
    Write-Host "Build FAILED!" -ForegroundColor Red
    Write-Host "========================================"  -ForegroundColor Red
    Write-Host ""
    Write-Host "Please check the error messages above." -ForegroundColor Yellow
    Write-Host ""
}

Read-Host "Press Enter to exit"
