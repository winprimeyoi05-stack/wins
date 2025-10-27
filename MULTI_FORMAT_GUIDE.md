# 📦 Panduan Multi-Format Produk

Sistem ini sekarang mendukung berbagai format produk, tidak hanya terbatas pada format `email | password`.

## 🎯 Format yang Didukung

### 1. 🔐 Account (email | password)
Format tradisional untuk akun dengan email dan password.

**Contoh:**
```
user@example.com | password123
premium.account@gmail.com | SecurePass456!
```

**Cara menambahkan:**
```go
db.AddProductContent(productID, "account", "user@example.com | password123")
```

### 2. 🔗 Link (URL/Redeem Link)
Format untuk produk berbasis link redeem atau aktivasi.

**Contoh:**
```
https://netflix.com/redeem?code=NFLX-ABCD-1234-EFGH
https://spotify.com/premium/activate/ABC123XYZ
https://canva.com/redeem/CANVA-PRO-LINK-123
```

**Cara menambahkan:**
```go
db.AddProductContent(productID, "link", "https://netflix.com/redeem?code=NFLX-ABCD-1234")
```

### 3. 🎫 Code (Kode Redeem/Voucher)
Format untuk kode redeem, voucher, atau license key.

**Contoh:**
```
SPOTIFY-PREMIUM-ABC123
NETFLIX-VOUCHER-XYZ789
ADOBE-LICENSE-KEY-12345
STEAM-WALLET-CODE-999
```

**Cara menambahkan:**
```go
db.AddProductContent(productID, "code", "SPOTIFY-PREMIUM-ABC123")
```

### 4. 📝 Custom (Format Kustom)
Format bebas untuk data produk lainnya.

**Contoh:**
```
Username: player123 | Server: Asia | Level: 100
PIN: 123456 | Serial: ABCD-EFGH-1234
Akses VIP - Token: xyz789abc
```

**Cara menambahkan:**
```go
db.AddProductContent(productID, "custom", "Username: player123 | Server: Asia")
```

## 📊 Struktur Database

### Tabel `product_accounts`
```sql
CREATE TABLE product_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    product_id INTEGER NOT NULL,
    content_type TEXT DEFAULT 'account',  -- 'account', 'link', 'code', 'custom'
    content_data TEXT NOT NULL,           -- Data produk dalam format apapun
    email TEXT,                           -- Legacy (untuk backward compatibility)
    password TEXT,                        -- Legacy (untuk backward compatibility)
    is_sold BOOLEAN DEFAULT FALSE,
    sold_to_user_id INTEGER,
    sold_order_id TEXT,
    sold_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Tabel `sold_accounts`
```sql
CREATE TABLE sold_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id TEXT NOT NULL,
    product_id INTEGER NOT NULL,
    account_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    content_type TEXT DEFAULT 'account',
    content_data TEXT NOT NULL,
    email TEXT,                           -- Legacy
    password TEXT,                        -- Legacy
    sold_price INTEGER NOT NULL,
    sold_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## 💻 Cara Menggunakan di Kode

### Menambahkan Produk dengan Format Berbeda

```go
import "telegram-premium-store/internal/database"

// Account format
db.AddProductContent(1, "account", "user@example.com | password123")

// Link format
db.AddProductContent(2, "link", "https://netflix.com/redeem?code=ABC123")

// Code format
db.AddProductContent(3, "code", "SPOTIFY-PREMIUM-XYZ789")

// Custom format
db.AddProductContent(4, "custom", "Username: player123 | PIN: 9999")
```

### Mengambil dan Menampilkan Produk

```go
accounts, err := db.GetAvailableAccounts(productID)
if err != nil {
    return err
}

for _, account := range accounts {
    // Mendapatkan label berdasarkan tipe
    label := account.GetContentLabel() // Returns: 🔐 Akun, 🔗 Link, 🎫 Kode, atau 📝 Data
    
    // Mendapatkan konten yang sudah diformat
    content := account.FormatContent()
    
    fmt.Printf("%s: %s\n", label, content)
}
```

## 🎨 Tampilan untuk User

Ketika user membeli produk, mereka akan menerima format yang sesuai:

### Format Account (🔐)
```
🔐 Akun #1:
user@example.com | password123

Cara menggunakan:
1. Login dengan email dan password yang diberikan
2. Segera ganti password setelah login pertama
```

### Format Link (🔗)
```
🔗 Link #1:
https://netflix.com/redeem?code=ABC123

Cara menggunakan:
1. Klik link di atas
2. Login ke akun Anda
3. Ikuti instruksi untuk redeem
```

### Format Code (🎫)
```
🎫 Kode #1:
SPOTIFY-PREMIUM-XYZ789

Cara menggunakan:
1. Buka aplikasi/website
2. Masuk ke menu Redeem/Tukar Kode
3. Masukkan kode di atas
```

### Format Custom (📝)
```
📝 Data #1:
Username: player123 | Server: Asia | Level: 100

Cara menggunakan:
Gunakan data di atas sesuai instruksi produk
```

## 🔄 Migration dari Format Lama

Sistem secara otomatis melakukan migrasi data lama (email | password) ke format baru saat database diinisialisasi:

```sql
-- Data lama akan digabungkan menjadi content_data
UPDATE product_accounts 
SET content_data = email || ' | ' || password 
WHERE content_data IS NULL AND email IS NOT NULL AND password IS NOT NULL;
```

Field `email` dan `password` tetap ada untuk backward compatibility tapi **deprecated**.

## 📝 Best Practices

1. **Gunakan tipe yang sesuai:**
   - Akun dengan login? → `account`
   - URL redeem? → `link`
   - Kode voucher? → `code`
   - Format khusus? → `custom`

2. **Format data dengan jelas:**
   - Account: `email | password`
   - Link: URL lengkap dengan protocol (https://)
   - Code: Gunakan format yang mudah disalin
   - Custom: Buat format yang mudah dipahami user

3. **Validasi input:**
   - Pastikan format sesuai sebelum disimpan
   - Link harus valid dan bisa diakses
   - Code tidak boleh kosong atau duplikat

4. **Testing:**
   - Test setiap format sebelum dijual
   - Pastikan user bisa menggunakan produk dengan mudah
   - Verifikasi copyable text berfungsi dengan baik

## 🛠️ Troubleshooting

### Q: Data lama tidak muncul?
**A:** Jalankan migration query untuk update field `content_data`.

### Q: Bagaimana menambahkan format baru?
**A:** Tambahkan konstanta baru di `models.ProductContentType` dan update fungsi `GetContentLabel()`.

### Q: Apakah bisa mixing format dalam satu produk?
**A:** Ya! Satu produk bisa memiliki berbagai format stock (account, link, code).

## 📞 Support

Untuk pertanyaan atau bantuan lebih lanjut, hubungi tim development atau buka issue di repository.

---

**Update Terakhir:** 2024-10-27
**Versi:** 2.0.0
