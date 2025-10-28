# Quick Fix for Windows Error 193

## The Problem
You're trying to run a Linux binary on Windows, which causes **Error 193: "Not a valid Win32 application"**.

## The Solution (Choose ONE method)

### ✅ Method 1: Use Build Script (Easiest!)

**Open Command Prompt** in your project folder and run:
```cmd
build-windows.bat
```

**OR use PowerShell:**
```powershell
.\build-windows.ps1
```

This will:
- ✓ Check if Go and GCC are installed
- ✓ Clean old Linux binaries
- ✓ Build a proper Windows `.exe` file
- ✓ Tell you exactly how to run it

---

### ✅ Method 2: Use Make

**Open Git Bash or WSL** and run:
```bash
make clean
make run
```

This will rebuild everything and start the bot.

---

### ✅ Method 3: Manual Build

**If you prefer to do it manually:**

1. **Delete the old binary:**
   ```bash
   rm -rf bin
   ```

2. **Build for Windows:**
   ```bash
   mkdir bin
   go build -o bin/telegram-store-bot.exe cmd/bot/main.go
   ```

3. **Run the Windows executable:**
   ```cmd
   .\bin\telegram-store-bot.exe
   ```

---

### ✅ Method 4: Run Without Building

**Skip the build entirely:**
```bash
go run cmd/bot/main.go
```

This runs directly from source code (slower but works immediately).

---

## Prerequisites Check

Before building, make sure you have:

1. **Go** installed (version 1.21+)
   - Check: `go version`
   - Download: https://go.dev/dl/

2. **GCC** installed (required for SQLite)
   - Check: `gcc --version`
   - Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
   - Or MinGW-w64: https://www.mingw-w64.org/

3. **Environment file** (`.env`)
   - Copy from `.env.example` if needed
   - Add your `BOT_TOKEN` from @BotFather
   - Add your `ADMIN_IDS` from @userinfobot

---

## Common Mistakes

❌ **Running:** `./bin/telegram-store-bot` (Linux binary)  
✅ **Correct:** `.\bin\telegram-store-bot.exe` (Windows binary)

❌ **Having:** `bin/telegram-store-bot` (no .exe)  
✅ **Should be:** `bin/telegram-store-bot.exe` (with .exe)

❌ **Mixing:** Building on Linux, running on Windows  
✅ **Always:** Build on the same OS you'll run on

---

## Still Not Working?

### If GCC is missing:
```
Error: gcc: command not found
```
**Solution:** Install TDM-GCC from https://jmeubank.github.io/tdm-gcc/

### If Go is missing:
```
Error: go: command not found
```
**Solution:** Install Go from https://go.dev/dl/

### If .env is missing:
```
Error: BOT_TOKEN is required
```
**Solution:** 
```bash
cp .env.example .env
notepad .env  # Edit with your bot token
```

---

## After Building Successfully

1. **Run the bot:**
   ```cmd
   .\bin\telegram-store-bot.exe
   ```

2. **Open Telegram** and start a chat with your bot

3. **Setup QRIS** (if you want payment features):
   - Send `/qrissetup` to your bot
   - Follow the instructions

---

## Need More Help?

See the full documentation:
- [WINDOWS_BUILD.md](WINDOWS_BUILD.md) - Complete Windows build guide
- [README.md](README.md) - General documentation
- [INSTALLATION.md](INSTALLATION.md) - Installation guide
- [QRIS_SETUP_GUIDE.md](QRIS_SETUP_GUIDE.md) - Payment setup
