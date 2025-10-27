# 🛠️ Panduan Instalasi Bot Telegram Penjualan Aplikasi Premium

## 📋 Persyaratan Sistem

- **Python**: 3.7 atau lebih baru
- **RAM**: Minimal 512MB
- **Storage**: Minimal 100MB free space
- **OS**: Linux, Windows, atau macOS
- **Internet**: Koneksi stabil untuk Telegram API

## 🚀 Instalasi Cepat

### Metode 1: Instalasi Otomatis (Recommended)

```bash
# 1. Clone atau download project
git clone <repository-url>
cd telegram-premium-app-bot

# 2. Jalankan setup otomatis
python3 setup.py
```

### Metode 2: Instalasi Manual

```bash
# 1. Install dependencies
pip3 install -r requirements.txt

# 2. Copy environment file
cp .env.example .env

# 3. Edit .env file
nano .env  # atau text editor lainnya

# 4. Initialize database
python3 -c "from database import Database; Database()"
```

## 🤖 Setup Bot Telegram

### 1. Buat Bot Baru

1. Buka Telegram dan cari **@BotFather**
2. Ketik `/start` untuk memulai
3. Ketik `/newbot` untuk membuat bot baru
4. Ikuti instruksi:
   - Masukkan nama bot (contoh: "Premium Apps Store")
   - Masukkan username bot (contoh: "premium_apps_store_bot")
5. **Simpan token** yang diberikan BotFather

### 2. Dapatkan User ID Admin

1. Cari **@userinfobot** di Telegram
2. Ketik `/start`
3. **Simpan User ID** yang ditampilkan

### 3. Konfigurasi Bot

Edit file `.env`:

```env
# Ganti dengan token dari BotFather
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# Ganti dengan User ID Anda (bisa multiple, pisahkan dengan koma)
ADMIN_IDS=123456789,987654321
```

## 🏃‍♂️ Menjalankan Bot

### Metode 1: Direct Python

```bash
python3 bot.py
```

### Metode 2: Menggunakan Runner

```bash
python3 run.py
```

### Metode 3: Docker (Production)

```bash
# Build dan jalankan dengan Docker Compose
docker-compose up -d
```

## 🔧 Konfigurasi Lanjutan

### Payment Methods

Edit `config.py` untuk mengubah metode pembayaran:

```python
PAYMENT_METHODS = {
    'dana': 'DANA: 081234567890',
    'gopay': 'GoPay: 081234567890', 
    'ovo': 'OVO: 081234567890',
    'bank': 'BCA: 1234567890 a.n. Nama Anda'
}
```

### Pesan Bot

Ubah pesan di `config.py` sesuai kebutuhan:

```python
MESSAGES = {
    'welcome': "Pesan selamat datang Anda...",
    'help': "Pesan bantuan Anda...",
    # dst...
}
```

## 📊 Tools Admin

Gunakan admin tools untuk mengelola bot:

```bash
python3 admin_tools.py
```

Menu yang tersedia:
- ➕ Tambah produk baru
- 📋 Lihat daftar produk
- 👥 Lihat pengguna
- 💰 Kelola pesanan
- 📊 Statistik bot

## 🐳 Deployment dengan Docker

### 1. Build Image

```bash
docker build -t premium-app-bot .
```

### 2. Run Container

```bash
docker run -d \
  --name premium-app-bot \
  --restart unless-stopped \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/.env:/app/.env:ro \
  premium-app-bot
```

### 3. Menggunakan Docker Compose (Recommended)

```bash
# Start
docker-compose up -d

# Stop
docker-compose down

# View logs
docker-compose logs -f

# Restart
docker-compose restart
```

## 🌐 Deployment ke VPS

### 1. Upload Files

```bash
# Menggunakan SCP
scp -r . user@your-server.com:/home/user/telegram-bot/

# Atau menggunakan Git
git clone <repository-url> /home/user/telegram-bot/
```

### 2. Setup Environment

```bash
ssh user@your-server.com
cd /home/user/telegram-bot/

# Install Python dan pip jika belum ada
sudo apt update
sudo apt install python3 python3-pip

# Setup bot
python3 setup.py
```

### 3. Setup Systemd Service (Auto-start)

Buat file `/etc/systemd/system/telegram-bot.service`:

```ini
[Unit]
Description=Telegram Premium App Bot
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/user/telegram-bot
ExecStart=/usr/bin/python3 /home/user/telegram-bot/bot.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable dan start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable telegram-bot
sudo systemctl start telegram-bot

# Check status
sudo systemctl status telegram-bot
```

## 🔍 Troubleshooting

### Bot Tidak Merespon

1. **Check Token**: Pastikan BOT_TOKEN benar
2. **Check Network**: Pastikan server bisa akses api.telegram.org
3. **Check Logs**: Lihat error di console atau logs

```bash
# Check logs jika menggunakan systemd
sudo journalctl -u telegram-bot -f

# Check logs jika menggunakan Docker
docker-compose logs -f
```

### Database Error

```bash
# Reset database
rm bot_database.db
python3 -c "from database import Database; Database()"
```

### Import Error

```bash
# Install ulang dependencies
pip3 install -r requirements.txt --force-reinstall
```

### Permission Error

```bash
# Fix permissions
chmod +x *.py
chown -R $USER:$USER .
```

## 📈 Monitoring & Maintenance

### 1. Check Bot Status

```bash
# Manual check
python3 -c "
import requests
token = 'YOUR_BOT_TOKEN'
response = requests.get(f'https://api.telegram.org/bot{token}/getMe')
print('✅ Bot OK' if response.json()['ok'] else '❌ Bot Error')
"
```

### 2. Database Backup

```bash
# Backup database
cp bot_database.db backup_$(date +%Y%m%d_%H%M%S).db

# Automated backup (add to crontab)
# 0 2 * * * cd /path/to/bot && cp bot_database.db backup_$(date +\%Y\%m\%d).db
```

### 3. Log Rotation

```bash
# Setup logrotate untuk bot logs
sudo nano /etc/logrotate.d/telegram-bot
```

```
/var/log/telegram-bot/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 ubuntu ubuntu
}
```

## 🆘 Support

Jika mengalami masalah:

1. 📖 Baca dokumentasi lengkap di README.md
2. 🔍 Check troubleshooting di atas
3. 📧 Hubungi developer
4. 🐛 Report bug di GitHub Issues

## 🔄 Update Bot

```bash
# Backup data
cp bot_database.db backup.db

# Pull update (jika menggunakan Git)
git pull origin main

# Install dependencies baru (jika ada)
pip3 install -r requirements.txt

# Restart bot
sudo systemctl restart telegram-bot
# atau
docker-compose restart
```