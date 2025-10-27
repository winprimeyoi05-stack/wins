# 🛠️ Panduan Instalasi Bot Telegram Premium Store (Go)

## 📋 Persyaratan Sistem

- **Go**: 1.21 atau lebih baru
- **SQLite3**: Biasanya sudah terinstall di sistem
- **RAM**: Minimal 512MB
- **Storage**: Minimal 100MB free space
- **OS**: Linux, Windows, atau macOS
- **Internet**: Koneksi stabil untuk Telegram API

## 🚀 Instalasi Cepat

### Metode 1: Instalasi Otomatis (Recommended)

```bash
# 1. Clone project
git clone <repository-url>
cd telegram-premium-store

# 2. Quick setup (install deps, create .env, build)
make quick-start

# 3. Edit konfigurasi
nano .env

# 4. Jalankan bot
make run
```

### Metode 2: Instalasi Manual

```bash
# 1. Install Go dependencies
go mod download
go mod tidy

# 2. Copy environment file
cp .env.example .env

# 3. Build aplikasi
go build -o bin/telegram-store-bot cmd/bot/main.go

# 4. Jalankan
./bin/telegram-store-bot
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
# Token dari BotFather
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz

# User ID Anda (bisa multiple, pisahkan dengan koma)
ADMIN_IDS=123456789,987654321

# Konfigurasi QRIS
QRIS_MERCHANT_NAME=Premium Apps Store
QRIS_MERCHANT_ID=ID1234567890123
QRIS_CITY=Jakarta
```

## 🏃‍♂️ Menjalankan Bot

### Development Mode

```bash
# Dengan auto-reload (install air dulu)
go install github.com/cosmtrek/air@latest
make dev

# Atau manual
make run
```

### Production Mode

```bash
# Build untuk production
make deploy-build

# Jalankan
./bin/telegram-store-bot
```

### Docker Mode

```bash
# Build dan jalankan dengan Docker
make docker-run

# Atau dengan docker-compose
make docker-compose-up
```

## 🔧 Konfigurasi Lanjutan

### Konfigurasi QRIS

Edit file `.env` untuk menyesuaikan informasi merchant QRIS:

```env
# Informasi Merchant QRIS
QRIS_MERCHANT_ID=ID1234567890123
QRIS_MERCHANT_NAME=Nama Toko Anda
QRIS_CITY=Kota Anda
QRIS_COUNTRY_CODE=ID
QRIS_CURRENCY_CODE=360
```

### Konfigurasi Database

```env
# Path database SQLite
DATABASE_PATH=store.db

# Atau untuk production dengan path absolut
DATABASE_PATH=/opt/telegram-store-bot/store.db
```

### Konfigurasi Logging

```env
# Level logging (DEBUG, INFO, WARN, ERROR)
LOG_LEVEL=INFO

# Untuk development
LOG_LEVEL=DEBUG
```

## 📊 Tools Admin

### CLI Admin Tools

```bash
# Build admin tools
go build -o bin/admin cmd/admin/main.go

# Jalankan admin CLI
./bin/admin

# Atau dengan Makefile
make admin
```

### Bot Admin Commands

Setelah bot berjalan, gunakan perintah admin:

```
/admin      - Panel admin
/addproduct - Tambah produk baru
/users      - Statistik pengguna  
/orders     - Kelola pesanan
/stats      - Statistik bot
```

## 🐳 Deployment dengan Docker

### 1. Build Image

```bash
# Build Docker image
make docker-build

# Atau manual
docker build -t telegram-store-bot .
```

### 2. Run Container

```bash
# Run dengan environment file
docker run -d \
  --name telegram-store-bot \
  --restart unless-stopped \
  --env-file .env \
  -v $(pwd)/data:/app/data \
  telegram-store-bot
```

### 3. Docker Compose (Recommended)

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Restart
docker-compose restart
```

## 🌐 Deployment ke VPS

### 1. Persiapan Server

```bash
# Update sistem
sudo apt update && sudo apt upgrade -y

# Install Go (jika belum ada)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Git
sudo apt install git -y
```

### 2. Deploy Aplikasi

```bash
# Clone repository
git clone <repository-url> /opt/telegram-store-bot
cd /opt/telegram-store-bot

# Build aplikasi
make deploy-build

# Setup environment
cp .env.example .env
nano .env  # Edit dengan konfigurasi Anda
```

### 3. Setup Systemd Service

```bash
# Install sebagai service
make install-service

# Start dan enable service
sudo systemctl start telegram-store-bot
sudo systemctl enable telegram-store-bot

# Check status
sudo systemctl status telegram-store-bot

# View logs
sudo journalctl -u telegram-store-bot -f
```

### 4. Setup Nginx (Opsional)

Jika ingin menggunakan webhook mode:

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location /webhook {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🔍 Troubleshooting

### Bot Tidak Merespon

1. **Check Token**:
```bash
curl "https://api.telegram.org/bot<TOKEN>/getMe"
```

2. **Check Logs**:
```bash
# Jika menggunakan systemd
sudo journalctl -u telegram-store-bot -f

# Jika menggunakan Docker
docker-compose logs -f

# Jika running manual
./bin/telegram-store-bot
```

3. **Check Network**:
```bash
# Test koneksi ke Telegram API
ping api.telegram.org
```

### Database Error

```bash
# Check database file
ls -la store.db

# Reset database
make db-reset

# Backup database
make backup
```

### Import/Build Error

```bash
# Clean dan rebuild
make clean
make deps
make build

# Check Go version
go version

# Update dependencies
go mod tidy
```

### Permission Error

```bash
# Fix file permissions
chmod +x bin/telegram-store-bot
chown -R $USER:$USER .

# Fix service permissions (jika menggunakan systemd)
sudo chown -R ubuntu:ubuntu /opt/telegram-store-bot
```

### Docker Issues

```bash
# Rebuild image
make docker-build

# Check container status
docker ps -a

# View container logs
docker logs telegram-store-bot

# Reset containers
docker-compose down -v
docker-compose up -d
```

## 📈 Monitoring & Maintenance

### 1. Health Check

```bash
# Check bot status
make status

# Check application health
curl http://localhost:8080/health  # jika webhook mode
```

### 2. Database Backup

```bash
# Manual backup
make backup

# Automated backup (add to crontab)
0 2 * * * cd /opt/telegram-store-bot && make backup
```

### 3. Log Rotation

Setup logrotate untuk mengelola logs:

```bash
# Create logrotate config
sudo nano /etc/logrotate.d/telegram-store-bot
```

```
/var/log/telegram-store-bot/*.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 644 ubuntu ubuntu
}
```

### 4. Update Aplikasi

```bash
# Backup data
make backup

# Pull updates
git pull origin main

# Rebuild
make build

# Restart service
sudo systemctl restart telegram-store-bot

# Check status
sudo systemctl status telegram-store-bot
```

## 🔒 Security Best Practices

### 1. Environment Security

```bash
# Set proper file permissions
chmod 600 .env

# Don't commit .env to git
echo ".env" >> .gitignore
```

### 2. Database Security

```bash
# Set database permissions
chmod 600 store.db

# Regular backups
make backup
```

### 3. Service Security

```bash
# Run as non-root user
sudo useradd -r -s /bin/false telegram-bot

# Set service permissions in systemd
User=telegram-bot
Group=telegram-bot
```

### 4. Network Security

```bash
# Firewall setup (jika diperlukan)
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP (jika webhook)
sudo ufw allow 443   # HTTPS (jika webhook)
sudo ufw enable
```

## 🆘 Support

Jika mengalami masalah:

1. 📖 Baca dokumentasi lengkap di README.md
2. 🔍 Check troubleshooting guide di atas
3. 📧 Hubungi developer
4. 🐛 Report bug di GitHub Issues

## 🔄 Update & Maintenance

### Regular Updates

```bash
# Weekly maintenance
make backup
git pull origin main
make build
sudo systemctl restart telegram-store-bot
```

### Version Upgrades

```bash
# Check current version
git log --oneline -n 5

# Update to latest
git fetch --tags
git checkout v1.x.x  # replace with latest version

# Rebuild and restart
make build
sudo systemctl restart telegram-store-bot
```