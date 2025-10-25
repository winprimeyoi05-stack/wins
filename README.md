# 🛒 Telegram Premium Store Bot - Go Edition

Bot Telegram yang dibangun dengan **Go (Golang)** untuk penjualan aplikasi premium dengan sistem pembayaran **QRIS dinamis**. Bot ini menyediakan pengalaman berbelanja yang lengkap dengan interface bahasa Indonesia yang user-friendly.

## ✨ Fitur Utama

### 🛍️ **Untuk Pelanggan**
- 📱 **Katalog Produk Lengkap** dengan sistem kategori
- 🛒 **Keranjang Belanja** dengan manajemen item
- 💳 **Pembayaran QRIS Dinamis** - QR Code otomatis ter-generate
- 📋 **Riwayat Pembelian** dengan detail lengkap
- 🔍 **Detail Produk** dengan informasi komprehensif
- 📞 **Customer Support** terintegrasi
- 🇮🇩 **Full Indonesian Language** support

### 👨‍💼 **Untuk Admin**
- 📊 **Dashboard Admin** untuk monitoring
- 📦 **Manajemen Produk** (CRUD operations)
- 👥 **Manajemen User** dan statistik
- 💰 **Kelola Pesanan** dan status pembayaran
- 📈 **Statistik Penjualan** real-time

### 🔧 **Fitur Teknis**
- ⚡ **High Performance** dengan Go
- 🗄️ **SQLite Database** dengan relasi yang proper
- 🔒 **Security First** - Admin access control, SQL injection protection
- 🐳 **Docker Ready** untuk deployment mudah
- 📊 **Structured Logging** dengan Logrus
- 🔄 **Auto-reload** development dengan Air
- 🛠️ **Makefile** untuk task automation

## 💳 Sistem Pembayaran QRIS

Bot ini menggunakan **QRIS (Quick Response Code Indonesian Standard)** yang mendukung semua aplikasi e-wallet dan mobile banking di Indonesia:

### 🏦 **Bank yang Didukung:**
- BCA Mobile, BNI Mobile Banking, BRI Mobile
- Mandiri Online, CIMB Niaga, Jenius

### 💰 **E-Wallet yang Didukung:**
- DANA, OVO, GoPay, LinkAja
- ShopeePay, Sakuku, i.saku, DOKU Wallet

### 🔄 **Fitur QRIS:**
- ✅ QR Code dinamis dengan nominal otomatis
- ⏰ Expiry time 15 menit per transaksi
- 🔐 Secure payment dengan EMV QR Code standard
- 📱 Compatible dengan semua aplikasi QRIS

## 🚀 Quick Start

### 1. **Clone & Setup**
```bash
git clone <repository-url>
cd telegram-premium-store
make quick-start
```

### 2. **Konfigurasi Bot**
```bash
# Edit file .env
nano .env

# Isi minimal konfigurasi ini:
BOT_TOKEN=your_bot_token_from_botfather
ADMIN_IDS=your_telegram_user_id
QRIS_MERCHANT_NAME=Nama Toko Anda
```

### 3. **Jalankan Bot**
```bash
make run
```

## 📋 Instalasi Lengkap

### **Persyaratan Sistem**
- **Go 1.21+** 
- **SQLite3** (biasanya sudah terinstall)
- **Git**

### **Instalasi Dependencies**
```bash
# Install Go dependencies
make deps

# Build aplikasi
make build

# Setup environment
make setup
```

### **Konfigurasi Bot Telegram**

#### 1. **Buat Bot Baru**
```
1. Buka Telegram, cari @BotFather
2. Ketik /newbot
3. Ikuti instruksi untuk nama dan username bot
4. Simpan token yang diberikan
```

#### 2. **Dapatkan User ID Admin**
```
1. Cari @userinfobot di Telegram  
2. Ketik /start
3. Simpan User ID yang ditampilkan
```

#### 3. **Edit Konfigurasi**
```bash
# Copy template
cp .env.example .env

# Edit dengan editor favorit
nano .env
```

**Konfigurasi Minimal:**
```env
BOT_TOKEN=1234567890:ABCdefGHIjklMNOpqrsTUVwxyz
ADMIN_IDS=123456789
QRIS_MERCHANT_NAME=Premium Apps Store
QRIS_MERCHANT_ID=ID1234567890123
```

## 🏃‍♂️ Menjalankan Bot

### **Development Mode**
```bash
# Dengan auto-reload (install air dulu)
go install github.com/cosmtrek/air@latest
make dev

# Atau manual
make run
```

### **Production Mode**
```bash
# Build untuk production
make deploy-build

# Jalankan
./bin/telegram-store-bot
```

### **Dengan Docker**
```bash
# Build dan jalankan
make docker-run

# Atau dengan docker-compose
make docker-compose-up
```

## 📊 Manajemen & Monitoring

### **Admin Tools**
```bash
# Akses panel admin via bot
/admin

# Atau gunakan Makefile commands
make admin          # CLI admin tools
make logs           # View logs
make status         # Check status
make backup         # Backup database
```

### **Database Management**
```bash
# Reset database
make db-reset

# Backup database
make backup
```

### **Development Tools**
```bash
# Format code
make format

# Run tests
make test

# Test dengan coverage
make test-coverage

# Linting
make lint
```

## 🐳 Deployment

### **Docker Deployment**
```bash
# Build image
make docker-build

# Run dengan docker-compose
docker-compose up -d

# View logs
docker-compose logs -f
```

### **VPS Deployment**
```bash
# Build untuk production
make deploy-build

# Install sebagai systemd service (Linux)
make install-service

# Start service
sudo systemctl start telegram-store-bot
sudo systemctl enable telegram-store-bot
```

### **Heroku Deployment**
```bash
# Install Heroku CLI, lalu:
heroku create your-app-name
heroku config:set BOT_TOKEN=your_token
heroku config:set ADMIN_IDS=your_id
git push heroku main
```

## 📱 Cara Penggunaan Bot

### **Untuk Pelanggan:**

1. **Mulai Berbelanja**
   ```
   /start → Lihat menu utama
   /catalog → Browse produk
   ```

2. **Pilih Produk**
   ```
   📱 Pilih kategori atau lihat semua
   👁️ Klik "Detail" untuk info lengkap
   🛒 Klik "Beli" untuk tambah ke keranjang
   ```

3. **Checkout & Bayar**
   ```
   🛒 Buka keranjang → Checkout
   📱 Scan QR Code QRIS dengan aplikasi e-wallet
   ✅ Pembayaran otomatis terverifikasi
   ```

### **Untuk Admin:**

1. **Akses Panel Admin**
   ```
   /admin → Dashboard admin
   /addproduct → Tambah produk baru
   /users → Statistik pengguna
   /orders → Kelola pesanan
   ```

2. **Tambah Produk**
   ```
   Format: /addproduct Nama | Deskripsi | Harga | Kategori
   Contoh: /addproduct Spotify Premium | Musik unlimited | 25000 | music
   ```

## 🛠️ Struktur Project

```
telegram-premium-store/
├── cmd/bot/main.go              # Entry point aplikasi
├── internal/
│   ├── bot/                     # Bot handlers
│   │   ├── bot.go              # Main bot logic
│   │   └── callbacks.go        # Callback handlers
│   ├── config/                  # Konfigurasi
│   │   └── config.go           # Config & messages
│   ├── database/                # Database layer
│   │   └── database.go         # DB operations
│   ├── models/                  # Data models
│   │   └── models.go           # Struct definitions
│   └── payment/                 # Payment system
│       └── qris.go             # QRIS implementation
├── go.mod                       # Go modules
├── Makefile                     # Task automation
├── Dockerfile                   # Docker configuration
├── docker-compose.yml           # Docker Compose
├── .env.example                 # Environment template
└── README.md                    # Documentation
```

## 🔧 Kustomisasi

### **Menambah Produk Sample**
Edit `internal/database/database.go` pada fungsi `insertSampleData()`:

```go
sampleProducts := []models.Product{
    {
        Name:        "Produk Baru",
        Description: "Deskripsi produk",
        Price:       50000,
        Category:    "kategori",
        Stock:       100,
    },
    // tambah produk lainnya...
}
```

### **Mengubah Pesan Bot**
Edit `internal/config/config.go` pada struct `Messages`:

```go
Welcome: `🎉 *Selamat datang di Toko Anda!* 🎉
Pesan selamat datang kustom...`,
```

### **Menambah Kategori Produk**
Edit `internal/models/models.go` pada fungsi `GetDefaultCategories()`:

```go
{Name: "kategori_baru", DisplayName: "🆕 Kategori Baru", Icon: "🆕"},
```

### **Kustomisasi QRIS**
Edit `internal/payment/qris.go` untuk integrasi dengan payment gateway nyata:

```go
// Ganti dengan API payment gateway
func (q *QRISService) ValidatePayment(orderID string, amount int) (bool, error) {
    // Implementasi API call ke Midtrans/Xendit/dll
    return callPaymentGatewayAPI(orderID, amount)
}
```

## 🔒 Keamanan

### **Fitur Keamanan:**
- ✅ **Admin Access Control** dengan User ID verification
- ✅ **SQL Injection Protection** dengan prepared statements
- ✅ **Input Validation** untuk semua user input
- ✅ **Environment Variables** untuk data sensitif
- ✅ **Secure QRIS** dengan EMV standard

### **Best Practices:**
```bash
# Jangan commit .env file
echo ".env" >> .gitignore

# Gunakan strong admin IDs
ADMIN_IDS=123456789,987654321

# Set proper file permissions
chmod 600 .env
```

## 📈 Monitoring & Analytics

### **Logging**
```bash
# View real-time logs
make logs

# Check application status  
make status
```

### **Database Monitoring**
```bash
# Backup database secara berkala
make backup

# Reset database jika diperlukan
make db-reset
```

### **Performance Monitoring**
Bot menggunakan structured logging dengan Logrus untuk monitoring:
- Request/response times
- Error tracking
- User activity logs
- Payment transaction logs

## 🤝 Kontribusi

### **Development Workflow**
```bash
# Fork repository
git clone your-fork-url
cd telegram-premium-store

# Create feature branch
git checkout -b feature/amazing-feature

# Make changes and test
make test
make lint

# Commit and push
git commit -m "Add amazing feature"
git push origin feature/amazing-feature

# Create Pull Request
```

### **Code Style**
```bash
# Format code sebelum commit
make format

# Run linter
make lint

# Run tests
make test-coverage
```

## 🆘 Troubleshooting

### **Bot Tidak Merespon**
```bash
# Check token dan network
curl https://api.telegram.org/bot<TOKEN>/getMe

# Check logs
make logs

# Restart bot
make run
```

### **Database Error**
```bash
# Reset database
make db-reset

# Check permissions
ls -la store.db
```

### **QRIS Error**
```bash
# Check QRIS configuration
grep QRIS .env

# Test QR code generation
# (implementasi test di development)
```

### **Docker Issues**
```bash
# Rebuild image
make docker-build

# Check container logs
docker-compose logs -f

# Reset containers
docker-compose down && docker-compose up -d
```

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

## 🙏 Acknowledgments

- **Telegram Bot API** team untuk dokumentasi yang excellent
- **Go Community** untuk ecosystem yang luar biasa  
- **Bank Indonesia** untuk QRIS standard
- **Indonesian Developer Community** untuk support dan feedback

## 📞 Support

Jika membutuhkan bantuan atau customization:

- 📧 **Email:** support@example.com
- 💬 **Telegram:** @your_username  
- 🐛 **Issues:** [GitHub Issues](https://github.com/your-repo/issues)
- 📖 **Wiki:** [Documentation Wiki](https://github.com/your-repo/wiki)

## 🎯 Roadmap

### **v1.1.0 - Coming Soon**
- [ ] Real payment gateway integration (Midtrans, Xendit)
- [ ] Advanced analytics dashboard
- [ ] Multi-language support
- [ ] Product search functionality
- [ ] Discount codes & promotions

### **v1.2.0 - Future**
- [ ] Subscription management
- [ ] Affiliate program  
- [ ] API endpoints untuk external integration
- [ ] Mobile app companion
- [ ] AI-powered customer support

---

**Dibuat dengan ❤️ menggunakan Go untuk komunitas Indonesia**

🚀 **Ready untuk production** | 🔒 **Security first** | 📱 **Mobile optimized** | 🇮🇩 **Indonesian focused**