# Payment Security & Account Delivery System

## 🔒 Fitur Keamanan Pembayaran QRIS

Sistem ini dilengkapi dengan verifikasi pembayaran yang komprehensif untuk mencegah manipulasi nominal QRIS:

### 1. **Verifikasi Nominal Pembayaran**

#### Mekanisme Verifikasi:
- Setiap order dibuat dengan `verification_hash` yang unik
- Hash dibuat berdasarkan: `OrderID + Amount + QRISPayload + SecretKey`
- Saat pembayaran diterima, sistem akan memvalidasi:
  - ✅ Nominal yang dibayarkan sesuai dengan expected amount
  - ✅ QRIS payload tidak dimanipulasi
  - ✅ Verification hash cocok dengan yang tersimpan

#### Pencegahan Manipulasi:
```
⚠️ DETEKSI MANIPULASI:
- Jika nominal tidak sesuai → Pembayaran DITOLAK
- Admin mendapat notifikasi real-time
- Order tetap dalam status pending
- Dapat dilakukan investigasi lebih lanjut
```

#### Flow Verifikasi:
```
1. Checkout → Generate QRIS dengan nominal X
2. Sistem simpan: order_id, expected_amount, qris_payload, verification_hash
3. Pembayaran diterima dengan nominal Y
4. Validasi: Y == X?
   - JA → Proses pembayaran ✅
   - TIDAK → Tolak & notifikasi admin 🚨
```

### 2. **QRIS Payload Integrity Check**

Sistem melakukan validasi integrity pada QRIS payload:
- ✅ Format QRIS sesuai standar EMV
- ✅ Field-field required ada semua
- ✅ CRC checksum valid
- ✅ Tidak ada manipulasi pada merchant ID atau amount field

### 3. **Database Schema untuk Verification**

```sql
CREATE TABLE payment_verifications (
    id INTEGER PRIMARY KEY,
    order_id TEXT NOT NULL,
    expected_amount INTEGER NOT NULL,
    qris_payload TEXT NOT NULL,
    verification_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    verified_at DATETIME,
    FOREIGN KEY (order_id) REFERENCES orders (id)
);
```

---

## 📦 Sistem Pengiriman Akun

### 1. **Format Akun (email | password)**

Saat pembayaran berhasil, pembeli akan menerima akun dalam format:
```
spotify.premium@gmail.com | SpotifyPass123!
```

#### Fitur Copy-to-Clipboard:
- Setiap akun dikirim dalam format `code` di Telegram
- Pembeli dapat tap/klik untuk menyalin
- Format `email | password` memudahkan parsing
- Instruksi lengkap disertakan dalam pesan

### 2. **Automatic Account Delivery**

Flow pengiriman akun otomatis:

```
1. Pembayaran Berhasil
   ↓
2. Validasi Pembayaran (amount verification)
   ↓
3. Ambil akun dari database (status: available)
   ↓
4. Kirim akun ke pembeli dengan format copyable
   ↓
5. Mark akun sebagai terjual (moved to sold_accounts)
   ↓
6. Notifikasi ke admin dengan detail lengkap
```

### 3. **Pesan ke Pembeli**

Format pesan yang diterima pembeli:

```
🎉 PEMBAYARAN BERHASIL!

✅ Pembayaran Anda untuk Order #ABC12345 telah dikonfirmasi.

💰 Total Pembayaran: Rp 50.000
📅 Tanggal: 26/10/2024 14:30

━━━━━━━━━━━━━━━━━━━━━

🔐 AKUN PREMIUM ANDA:

📦 Spotify Premium 1 Bulan
   Jumlah: 2 akun

   🔑 Akun #1:
   spotify1@gmail.com | Pass123!

   🔑 Akun #2:
   spotify2@gmail.com | Pass456!

━━━━━━━━━━━━━━━━━━━━━

📋 CARA MENGGUNAKAN:
1. Tap/klik pada kredensial akun untuk menyalin
2. Paste pada aplikasi terkait
3. Pisahkan email dan password dengan '|'

⚠️ PENTING:
• Simpan kredensial ini dengan aman
• Jangan share ke orang lain
• Segera ganti password setelah login
```

---

## 📊 Stock Management System

### 1. **Stock Movement: Available → Sold**

#### Database Structure:

**Table: `product_accounts`** (Available Stock)
```sql
CREATE TABLE product_accounts (
    id INTEGER PRIMARY KEY,
    product_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    is_sold BOOLEAN DEFAULT FALSE,
    sold_to_user_id INTEGER,
    sold_order_id TEXT,
    sold_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Table: `sold_accounts`** (Sold Stock Tracking)
```sql
CREATE TABLE sold_accounts (
    id INTEGER PRIMARY KEY,
    order_id TEXT NOT NULL,
    product_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    sold_price INTEGER NOT NULL,
    sold_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### 2. **Stock Movement Process**

Saat pembayaran berhasil:

```go
1. Get available accounts (is_sold = FALSE)
2. Assign to buyer:
   - Mark as sold in product_accounts
   - Copy to sold_accounts table
   - Link to order_id and user_id
3. Update stock counters
```

### 3. **Admin Monitoring**

Admin dapat melihat:
- ✅ **Available Stock**: Akun yang belum terjual
- ✅ **Sold Stock**: Akun yang sudah terjual
- ✅ **Total Stock**: Total semua akun
- ✅ **Sales History**: Riwayat penjualan per produk

Query untuk monitoring:
```sql
-- Get stock summary
SELECT 
    product_id,
    COUNT(CASE WHEN is_sold = FALSE THEN 1 END) as available,
    COUNT(CASE WHEN is_sold = TRUE THEN 1 END) as sold,
    COUNT(*) as total
FROM product_accounts
GROUP BY product_id;

-- Get sold accounts with buyer info
SELECT 
    sa.*,
    p.name as product_name,
    u.first_name, u.last_name, u.username
FROM sold_accounts sa
JOIN products p ON sa.product_id = p.id
JOIN users u ON sa.user_id = u.user_id
ORDER BY sa.sold_at DESC;
```

---

## 🔔 Admin Notification System

### 1. **Notifikasi Penjualan Lengkap**

Saat ada pembayaran berhasil, admin menerima notifikasi komprehensif:

```
💰 PEMBERITAHUAN PENJUALAN BARU!

━━━━━━━━━━━━━━━━━━━━━

📋 INFORMASI PESANAN
🆔 Order ID: ORD-abc12345-xyz
📅 Waktu: 26/10/2024 14:30:45
💰 Total: Rp 50.000
✅ Status: LUNAS

👤 DATA PEMBELI
📛 Nama: John Doe
👤 Username: @johndoe
🆔 User ID: 123456789

📦 DETAIL PEMBELIAN

1. Spotify Premium 1 Bulan
   • Jumlah: 2 akun
   • Harga satuan: Rp 25.000
   • Subtotal: Rp 50.000

━━━━━━━━━━━━━━━━━━━━━

🔐 AKUN YANG TERJUAL

📦 Spotify Premium 1 Bulan (2 akun):
   1. spotify1@gmail.com | Pass123!
   2. spotify2@gmail.com | Pass456!

━━━━━━━━━━━━━━━━━━━━━

📊 STATUS STOK TERKINI
• Spotify Premium 1 Bulan:
  ✅ Tersedia: 3 akun
  💰 Terjual: 7 akun
  📊 Total: 10 akun

━━━━━━━━━━━━━━━━━━━━━

✨ Transaksi berhasil diproses!
```

### 2. **Notifikasi Deteksi Manipulasi**

Jika terdeteksi manipulasi nominal:

```
🚨 DETEKSI MANIPULASI PEMBAYARAN!

⚠️ Terdeteksi upaya manipulasi nominal pembayaran:

🆔 Order ID: ORD-abc12345
👤 User ID: 987654321
💰 Nominal Expected: Rp 50.000
❌ Nominal Diterima: Rp 10.000
📊 Selisih: Rp 40.000

⏰ Waktu: 26/10/2024 14:30:45

🔒 TINDAKAN YANG DIAMBIL:
• Pembayaran ditolak
• Order tetap pending
• Perlu investigasi lebih lanjut

💡 Silakan cek detail order dan user untuk 
   tindakan lebih lanjut.
```

### 3. **Interactive Admin Controls**

Admin dapat:
- 🔍 Investigasi order yang bermasalah
- 📊 Lihat stok real-time
- 💰 Kelola pesanan
- 🧪 Simulasi pembayaran (untuk testing)

---

## 🧪 Testing Features

### Simulasi Pembayaran (Admin Only)

Admin dapat mensimulasi pembayaran berhasil untuk testing:

1. Buka detail order (status: pending)
2. Klik tombol "🧪 [Admin] Simulasi Pembayaran"
3. Sistem akan:
   - ✅ Validasi order
   - ✅ Proses pembayaran
   - ✅ Kirim akun ke pembeli
   - ✅ Notifikasi admin
   - ✅ Update stock

### Testing Manipulation Detection

Untuk test deteksi manipulasi:

```go
// Simulasi pembayaran dengan nominal berbeda
err := bot.handlePaymentSuccess(orderID, wrongAmount)
// Expected: error + admin notification
```

---

## 🔧 Configuration

### Required Environment Variables

```env
BOT_TOKEN=your_telegram_bot_token
ADMIN_IDS=123456789,987654321
QRIS_MERCHANT_ID=your_merchant_id
QRIS_MERCHANT_NAME=Your Store Name
```

### Secret Key Setup

Verification hash menggunakan bot token sebagai secret key:
```go
verifier := payment.NewPaymentVerifier(config.BotToken)
```

**⚠️ PENTING**: Jangan expose bot token di mana pun!

---

## 📈 Benefits

### Untuk Penjual (Admin):
✅ **Keamanan**: Pencegahan manipulasi pembayaran  
✅ **Monitoring**: Real-time stock tracking  
✅ **Automation**: Pengiriman akun otomatis  
✅ **Reporting**: Notifikasi lengkap setiap transaksi  
✅ **Audit Trail**: Complete transaction history  

### Untuk Pembeli:
✅ **Instant Delivery**: Akun diterima segera setelah bayar  
✅ **Easy to Use**: Format copyable yang mudah digunakan  
✅ **Clear Instructions**: Panduan lengkap cara pakai  
✅ **Security**: Reminder untuk ganti password  

### Untuk Sistem:
✅ **Scalable**: Database optimized untuk high volume  
✅ **Reliable**: Transaction atomicity & consistency  
✅ **Auditable**: Complete logging & verification trail  
✅ **Maintainable**: Clean code architecture  

---

## 🚀 Future Enhancements

Rencana pengembangan ke depan:

1. **Payment Gateway Integration**
   - Integrasi dengan real QRIS gateway (Midtrans/Xendit)
   - Webhook untuk payment notification
   - Auto-verification dari gateway

2. **Advanced Anti-Fraud**
   - IP tracking
   - Rate limiting per user
   - Behavioral analysis
   - Blacklist management

3. **Enhanced Reporting**
   - Daily/weekly/monthly sales reports
   - Revenue analytics
   - Top products tracking
   - Customer lifetime value

4. **Account Management**
   - Bulk import accounts (CSV)
   - Account validity checker
   - Auto-renewal reminders
   - Account replacement policy

---

## 📞 Support

Untuk pertanyaan atau issues:
- Telegram: Contact admin via bot `/contact`
- GitHub: Create an issue
- Email: Check bot configuration

---

**Version**: 1.0.0  
**Last Updated**: October 26, 2024  
**Status**: ✅ Production Ready
