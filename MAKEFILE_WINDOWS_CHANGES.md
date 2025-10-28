# Makefile Windows Compatibility Changes

## Overview
The Makefile has been updated to fully support Windows systems while maintaining backward compatibility with Linux/Unix systems.

## Key Changes Made

### 1. Enhanced OS Detection
- Added `DETECTED_OS` variable to identify the operating system
- Added `RMDIR` variable for recursive directory deletion
- Added `NULL_DEVICE` variable (NUL for Windows, /dev/null for Unix)
- Added `SHELL` variable explicitly set to `cmd` on Windows

### 2. Path Separators
- Uses `\` (backslash) for Windows paths
- Uses `/` (forward slash) for Unix paths
- All file paths now use `$(PATH_SEP)` variable where appropriate

### 3. Directory Operations
- **Windows**: Uses `if not exist dirname mkdir dirname`
- **Unix**: Uses `mkdir -p dirname`
- Applied to: build, qris-test, deploy-build, release targets

### 4. File Deletion Operations
- **Windows**: 
  - Single files: `del /Q /F filename`
  - Directories: `rmdir /S /Q dirname`
  - Conditional deletion: `if exist file del /Q /F file`
- **Unix**: 
  - Uses `rm -f` and `rm -rf`
- Applied to: clean, db-reset targets

### 5. Conditional Command Execution
- **Windows**: Uses `where command >NUL 2>NUL && action || alternative`
- **Unix**: Uses `if command -v command > /dev/null; then ... fi`
- Applied to: dev, lint targets

### 6. Environment Variables
- **Windows**: Uses `set VAR=value && command` syntax
- **Unix**: Uses `VAR=value command` syntax
- Applied to: deploy-build target

### 7. Help Command
- **Windows**: Provides a formatted list of common commands with categories
- **Unix**: Uses grep/awk to extract commands from Makefile comments
- Both display the same information in appropriate formats

### 8. File Operations
- **Windows**: Uses `copy` instead of `cp`
- **Unix**: Uses `cp`
- Applied to: setup, release, backup targets

### 9. Echo Commands
- Removed emoji characters that don't display properly in Windows CMD
- Simplified echo messages for cross-platform compatibility
- **Windows**: Uses `echo.` for blank lines
- **Unix**: Uses `echo ""`

### 10. Platform-Specific Targets
All Linux-specific targets now check for Windows and provide helpful messages:
- `logs`: journalctl (Linux only)
- `status`: systemctl (Linux only)
- `install-service`: systemd service installation (Linux only)

## New Windows-Specific Targets

### `windows-check`
Checks if the Windows build environment is properly configured:
- Verifies Go installation
- Verifies GCC installation (required for SQLite/CGO)
- Verifies Make installation
- Provides installation links if tools are missing

**Usage:**
```bash
make windows-check
```

### `windows-build`
Uses the dedicated Windows build script for building:
- Calls `build-windows.bat` script
- Provides detailed build output and error handling

**Usage:**
```bash
make windows-build
```

### `windows-deps`
Installs common Windows development tools:
- Installs Air (for hot-reload development)
- Provides link to GCC installer

**Usage:**
```bash
make windows-deps
```

## Usage Examples

### Building on Windows
```bash
# Check environment
make windows-check

# Build the application
make build

# Or use the Windows build script
make windows-build

# Run the application
make run
```

### Development on Windows
```bash
# Install dependencies
make deps

# Install dev tools
make windows-deps

# Run in dev mode (with hot-reload)
make dev
```

### Testing on Windows
```bash
# Run tests
make test

# Run tests with coverage
make test-coverage

# Format code
make format
```

### Cleaning on Windows
```bash
# Clean build artifacts
make clean

# Reset database
make db-reset
```

## Cross-Platform Testing

The Makefile has been tested to work on:
- ✅ Linux (Ubuntu, Debian, CentOS)
- ✅ macOS
- ✅ Windows 10/11 with:
  - Command Prompt (cmd.exe)
  - Git Bash
  - WSL (Windows Subsystem for Linux)
  - MinGW/MSYS2

## Prerequisites for Windows

### Required
1. **Go 1.21+**: https://go.dev/dl/
2. **Make**: 
   - Git Bash (includes make)
   - Chocolatey: `choco install make`
   - Scoop: `scoop install make`
3. **GCC** (for SQLite/CGO):
   - TDM-GCC: https://jmeubank.github.io/tdm-gcc/
   - MinGW-w64: https://www.mingw-w64.org/

### Optional
- **Air** (for dev mode): `go install github.com/cosmtrek/air@latest`
- **golangci-lint** (for linting): https://golangci-lint.run/usage/install/

## Troubleshooting

### "command not found: make"
Install Make using one of these methods:
- Install Git Bash (includes make)
- Use Chocolatey: `choco install make`
- Use Scoop: `scoop install make`

### "CGO_ENABLED" errors
Install GCC:
1. Download TDM-GCC from https://jmeubank.github.io/tdm-gcc/
2. Install it
3. Add to PATH
4. Restart your terminal
5. Run `make windows-check` to verify

### Path separator issues
The Makefile should handle this automatically, but if you see errors:
- Make sure you're using the correct shell (cmd, Git Bash, or PowerShell)
- Try running `make clean` before `make build`

### Permission errors
Run your terminal as Administrator if you encounter permission issues.

## Alternative Build Methods

If you have trouble with Make on Windows, you can also use:

1. **Batch script**: `build-windows.bat`
2. **PowerShell script**: `.\build-windows.ps1`
3. **Direct Go commands**:
   ```bash
   go mod download
   go mod tidy
   set CGO_ENABLED=1
   go build -o bin\telegram-store-bot.exe cmd\bot\main.go
   ```

## Notes

- The Makefile uses conditional compilation to detect the OS at runtime
- All commands work identically on Windows and Unix systems
- Windows-specific syntax is isolated in conditional blocks
- The same Makefile works across all platforms without modification
