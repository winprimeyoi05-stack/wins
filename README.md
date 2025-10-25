# 🤖 Bot Telegram Penjualan Aplikasi Premium

Bot Telegram yang dirancang khusus untuk menjual aplikasi premium dengan fitur lengkap dan antarmuka yang user-friendly dalam bahasa Indonesia.

## ✨ Fitur Utama

### 👥 Untuk Pelanggan:
- 📱 **Katalog Produk Lengkap** - Browse aplikasi premium berdasarkan kategori
- 🛒 **Sistem Keranjang Belanja** - Tambah multiple produk sebelum checkout
- 💳 **Multiple Payment Methods** - DANA, GoPay, OVO, Transfer Bank
- 📋 **Riwayat Pembelian** - Track semua transaksi Anda
- 🔍 **Detail Produk** - Informasi lengkap setiap aplikasi
- 📞 **Customer Support** - Kontak langsung dengan admin

### 👨‍💼 Untuk Admin:
- 📊 **Dashboard Admin** - Kelola seluruh aspek toko
- 📦 **Manajemen Produk** - Tambah, edit, hapus produk
- 👥 **Manajemen User** - Lihat statistik pengguna
- 💰 **Kelola Pesanan** - Update status pembayaran
- 📈 **Statistik Penjualan** - Monitor performa toko

## 🚀 Instalasi & Setup

### 1. Clone Repository
```bash
git clone <repository-url>
cd telegram-premium-app-bot
```

### 2. Install Dependencies
```bash
pip install -r requirements.txt
```

### 3. Setup Environment
```bash
cp .env.example .env
```

Edit file `.env` dan isi dengan data Anda:
```env
BOT_TOKEN=your_telegram_bot_token_here
ADMIN_IDS=your_telegram_user_id,another_admin_id
```

### 4. Dapatkan Bot Token
1. Chat dengan [@BotFather](https://t.me/botfather) di Telegram
2. Ketik `/newbot` dan ikuti instruksi
3. Copy token yang diberikan ke file `.env`

### 5. Dapatkan User ID Anda
1. Chat dengan [@userinfobot](https://t.me/userinfobot)
2. Copy User ID Anda ke file `.env` sebagai ADMIN_IDS

### 6. Jalankan Bot
```bash
python bot.py
```

## 📱 Cara Penggunaan

### Untuk Pelanggan:

1. **Mulai Chat** - Ketik `/start` untuk memulai
2. **Browse Katalog** - Gunakan `/catalog` atau tombol "📱 Lihat Katalog"
3. **Filter Kategori** - Pilih kategori untuk melihat produk spesifik
4. **Detail Produk** - Klik "👁️ Detail" untuk info lengkap
5. **Tambah ke Keranjang** - Klik "🛒 Beli" atau "🛒 Tambah ke Keranjang"
6. **Checkout** - Buka keranjang dan pilih "💳 Checkout"
7. **Pembayaran** - Pilih metode pembayaran dan ikuti instruksi
8. **Konfirmasi** - Kirim bukti pembayaran ke admin

### Untuk Admin:

1. **Panel Admin** - Ketik `/admin` untuk mengakses dashboard
2. **Tambah Produk** - Gunakan `/addproduct` dengan format yang ditentukan
3. **Lihat Users** - Ketik `/users` untuk statistik pengguna
4. **Kelola Pesanan** - Akses melalui panel admin untuk update status

## 🛠️ Struktur Project

```
telegram-premium-app-bot/
├── bot.py              # Main bot application
├── database.py         # Database operations
├── config.py          # Configuration and messages
├── requirements.txt   # Python dependencies
├── .env.example      # Environment variables template
├── README.md         # Documentation
└── bot_database.db   # SQLite database (auto-created)
```

## 💾 Database Schema

Bot menggunakan SQLite dengan 4 tabel utama:

- **users** - Data pengguna
- **products** - Katalog produk
- **orders** - Riwayat pesanan
- **cart** - Keranjang belanja

## 🔧 Kustomisasi

### Menambah Produk Sample
Edit `database.py` pada fungsi `insert_sample_products()` untuk menambah produk default.

### Mengubah Pesan
Edit `config.py` pada dictionary `MESSAGES` untuk mengubah teks bot.

### Menambah Payment Method
Edit `config.py` pada dictionary `PAYMENT_METHODS` untuk menambah metode pembayaran.

### Styling Pesan
Bot menggunakan Markdown formatting. Anda bisa mengubah style di setiap fungsi handler.

## 📋 Perintah Bot

### Perintah Umum:
- `/start` - Mulai menggunakan bot
- `/help` - Bantuan dan daftar perintah
- `/catalog` - Lihat katalog produk
- `/cart` - Buka keranjang belanja
- `/history` - Riwayat pembelian
- `/contact` - Informasi kontak

### Perintah Admin:
- `/admin` - Panel admin
- `/addproduct` - Tambah produk baru
- `/users` - Statistik pengguna

## 🔒 Keamanan

- ✅ Admin access control dengan User ID verification
- ✅ SQL injection protection dengan parameterized queries
- ✅ Input validation untuk semua user input
- ✅ Environment variables untuk data sensitif

## 🚀 Deployment

### Heroku:
1. Create Heroku app
2. Set environment variables di Heroku dashboard
3. Deploy dengan Git atau GitHub integration

### VPS/Server:
1. Upload files ke server
2. Install Python dan dependencies
3. Setup systemd service untuk auto-restart
4. Gunakan reverse proxy (nginx) jika diperlukan

## 🤝 Kontribusi

1. Fork repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

## 📝 License

Distributed under the MIT License. See `LICENSE` for more information.

## 📞 Support

Jika Anda membutuhkan bantuan atau customization:

- 📧 Email: support@example.com
- 💬 Telegram: @your_username
- 🐛 Issues: [GitHub Issues](https://github.com/your-repo/issues)

## 🎯 Roadmap

- [ ] Integration dengan payment gateway (Midtrans, Xendit)
- [ ] Multi-language support
- [ ] Advanced analytics dashboard
- [ ] Automated delivery system
- [ ] Subscription management
- [ ] Affiliate program
- [ ] Mobile app companion

---

**Dibuat dengan ❤️ untuk komunitas Indonesia**