# Makefile Quick Start Guide for Windows

## Quick Setup

### 1. Check Your Environment
```bash
make windows-check
```
This will verify that Go, GCC, and Make are properly installed.

### 2. Install Dependencies
```bash
make deps
```

### 3. Build and Run
```bash
make build
make run
```

Or in one command:
```bash
make run
```

## Common Commands

### Building
```bash
make build              # Build the application
make windows-build      # Build using Windows batch script
make clean              # Clean build artifacts
make release            # Create release package
```

### Running
```bash
make run                # Build and run the bot
make dev                # Run with hot-reload (requires air)
```

### Testing
```bash
make test               # Run all tests
make test-coverage      # Run tests with coverage report
```

### QRIS Tools
```bash
make qris-test          # Build QRIS test tool
make qris-status        # Show QRIS configuration
make qris-upload IMAGE=path\to\qr.png  # Upload QRIS image
```

### Development
```bash
make format             # Format Go code
make lint               # Run linter
make windows-deps       # Install development tools
```

### Database
```bash
make db-reset           # Reset database
make backup             # Backup database and config
```

### Getting Help
```bash
make help               # Show all available commands
make                    # Same as 'make help'
```

## First Time Setup

1. **Install Prerequisites**
   ```bash
   # Download and install:
   # - Go from https://go.dev/dl/
   # - TDM-GCC from https://jmeubank.github.io/tdm-gcc/
   # - Git for Windows from https://git-scm.com/download/win (includes Git Bash with make)
   ```

2. **Verify Installation**
   ```bash
   make windows-check
   ```

3. **Setup Environment**
   ```bash
   make setup
   # Edit .env file with your BOT_TOKEN and ADMIN_IDS
   ```

4. **Build and Run**
   ```bash
   make quick-start
   # Follow the instructions displayed
   ```

## Alternative Methods

If you prefer not to use Make, you can also:

### Option 1: Use Build Scripts
```bash
# Command Prompt
build-windows.bat

# PowerShell
.\build-windows.ps1
```

### Option 2: Direct Go Commands
```bash
# Install dependencies
go mod download
go mod tidy

# Build
set CGO_ENABLED=1
go build -o bin\telegram-store-bot.exe cmd\bot\main.go

# Run
.\bin\telegram-store-bot.exe
```

## Troubleshooting

### Make not found
Install Git Bash which includes make, or install via:
- Chocolatey: `choco install make`
- Scoop: `scoop install make`

### GCC not found
1. Install TDM-GCC: https://jmeubank.github.io/tdm-gcc/
2. Add to PATH: `C:\TDM-GCC-64\bin`
3. Restart terminal
4. Verify: `gcc --version`

### CGO errors
Make sure:
1. GCC is installed
2. GCC is in your PATH
3. `CGO_ENABLED=1` is set (Makefile does this automatically)

### Path issues
Use Git Bash or adjust paths for your shell:
- Git Bash: Use Unix-style paths (`/c/Users/...`)
- CMD: Use Windows-style paths (`C:\Users\...`)

## Tips

1. **Use Git Bash**: Most compatible with Make on Windows
2. **Run as Administrator**: If you get permission errors
3. **Check Environment**: Run `make windows-check` if builds fail
4. **Clean First**: Run `make clean` before building after errors
5. **Use Windows Scripts**: If Make gives you trouble, use `build-windows.bat`

## Full Workflow Example

```bash
# 1. Clone and enter directory
git clone <repo>
cd telegram-store-bot

# 2. Check environment
make windows-check

# 3. Setup and configure
make setup
notepad .env  # Edit configuration

# 4. Install dev tools
make windows-deps

# 5. Build and test
make build
make test

# 6. Run the bot
make run
```

## Getting Help

- Run `make help` to see all commands
- Read `MAKEFILE_WINDOWS_CHANGES.md` for detailed changes
- Check `WINDOWS_BUILD.md` for Windows-specific build instructions
- See `README.md` for general project documentation
