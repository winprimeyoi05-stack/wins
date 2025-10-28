# Building and Running on Windows

## Prerequisites
- Go 1.21 or higher installed
- Make for Windows (via MinGW, Cygwin, or Git Bash)
- Git Bash or similar Unix-like shell

## Quick Start

### Option 1: Using Build Scripts (Easiest)

**For Command Prompt:**
```cmd
build-windows.bat
```

**For PowerShell:**
```powershell
.\build-windows.ps1
```

**For Git Bash / WSL:**
```bash
make clean && make build
```

### Option 2: Using Make
```bash
# Clean old builds first (important!)
make clean

# Build and run the bot
make run

# Or build only
make build

# Then run manually
.\bin\telegram-store-bot.exe
```

### Option 3: Direct Go Commands
If you don't have Make installed:

```bash
# Install dependencies
go mod download
go mod tidy

# Clean old builds
rm -rf bin
mkdir bin

# Build the application (ensure CGO is enabled)
set CGO_ENABLED=1
go build -o bin/telegram-store-bot.exe cmd/bot/main.go

# Run the application
.\bin\telegram-store-bot.exe
```

### Option 4: Run without Building
```bash
# Run directly from source
go run cmd/bot/main.go
```

## Common Issues

### Error 193: Not a valid Win32 application
This error occurs when trying to run a Linux binary on Windows. **This is the most common issue!**

**Solutions:**

**Method 1 - Use Build Scripts (Recommended):**
```cmd
REM For Command Prompt
build-windows.bat

REM For PowerShell
.\build-windows.ps1
```

**Method 2 - Manual Fix:**
```bash
# 1. Clean old builds
make clean
# OR manually: rm -rf bin

# 2. Rebuild for Windows
make build

# 3. Run the Windows executable
.\bin\telegram-store-bot.exe
```

**Important Notes:**
- The Windows executable MUST be named `telegram-store-bot.exe` (with .exe extension)
- If you see `./bin/telegram-store-bot` without `.exe`, it's a Linux binary
- Always run `make clean` before rebuilding if you previously built on Linux
- On Windows, use `.\bin\telegram-store-bot.exe` not `./bin/telegram-store-bot`

### CGO_ENABLED Error
If you get SQLite-related errors, ensure CGO is enabled:
```bash
set CGO_ENABLED=1
go build -o bin/telegram-store-bot.exe cmd/bot/main.go
```

### Missing GCC
SQLite requires CGO and GCC. Install one of:
- **TDM-GCC**: https://jmeubank.github.io/tdm-gcc/
- **MinGW-w64**: https://www.mingw-w64.org/
- **Cygwin**: https://www.cygwin.com/

## Environment Setup

1. Copy `.env.example` to `.env`:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` with your configuration:
   - `BOT_TOKEN`: Get from @BotFather on Telegram
   - `ADMIN_IDS`: Your Telegram user ID (get from @userinfobot)

## Running the Bot

```bash
# Using Make
make run

# Or directly
.\bin\telegram-store-bot.exe
```

## Development

```bash
# Run tests
make test

# Format code
make format

# Clean build artifacts
make clean
```

## QRIS Setup

After starting the bot:
1. Open Telegram and start a chat with your bot
2. Use the `/qrissetup` command
3. Follow the instructions to upload your static QR code
4. The bot will automatically generate dynamic QR codes for payments

## Troubleshooting

### Bot Won't Start
- Check `.env` file has valid `BOT_TOKEN`
- Verify firewall isn't blocking the connection
- Check logs for detailed error messages

### Database Errors
- Ensure write permissions in the directory
- Try deleting `store.db` and restarting

### Build Errors
- Update Go: `go version` should be 1.21+
- Clean and rebuild: `make clean && make build`
- Check GCC is installed: `gcc --version`

## Additional Help

For more information, see:
- [README.md](README.md) - General documentation
- [INSTALLATION.md](INSTALLATION.md) - Installation guide
- [QRIS_SETUP_GUIDE.md](QRIS_SETUP_GUIDE.md) - QRIS configuration
