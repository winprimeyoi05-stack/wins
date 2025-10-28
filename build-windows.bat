@echo off
REM Build script for Windows
echo Building Telegram Store Bot for Windows...
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go from https://go.dev/dl/
    exit /b 1
)

REM Check if GCC is installed (required for SQLite)
where gcc >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo WARNING: GCC is not installed
    echo SQLite requires CGO and GCC. Please install one of:
    echo   - TDM-GCC: https://jmeubank.github.io/tdm-gcc/
    echo   - MinGW-w64: https://www.mingw-w64.org/
    echo.
    pause
    exit /b 1
)

REM Clean old builds
echo Cleaning old builds...
if exist bin rmdir /s /q bin
if exist telegram-store-bot.exe del telegram-store-bot.exe

REM Create bin directory
if not exist bin mkdir bin

REM Download dependencies
echo Installing dependencies...
go mod download
go mod tidy

REM Build the application
echo Building application...
set CGO_ENABLED=1
go build -o bin\telegram-store-bot.exe cmd\bot\main.go

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ========================================
    echo Build successful!
    echo ========================================
    echo.
    echo Binary created: bin\telegram-store-bot.exe
    echo.
    echo To run the bot:
    echo   .\bin\telegram-store-bot.exe
    echo.
    echo Or use:
    echo   make run
    echo.
) else (
    echo.
    echo ========================================
    echo Build FAILED!
    echo ========================================
    echo.
    echo Please check the error messages above.
    echo.
)

pause
