# 📦 Multi-Format Product Support - Quick Start

## Apa yang Berubah?

Sistem Telegram Premium Store sekarang mendukung **berbagai format produk**, bukan hanya `email | password`!

### Format yang Didukung:

| Format | Icon | Contoh | Use Case |
|--------|------|--------|----------|
| **Account** | 🔐 | `user@gmail.com \| pass123` | Login credentials |
| **Link** | 🔗 | `https://netflix.com/redeem?code=ABC` | Redeem URLs |
| **Code** | 🎫 | `SPOTIFY-PREMIUM-XYZ789` | Voucher/License keys |
| **Custom** | 📝 | `UserID: 123 \| Level: 100` | Game accounts, etc |

## 🚀 Quick Start untuk Admin

### Menambahkan Stock Baru

Gunakan command `/addstock` dengan format:

```
/addstock [product_id] [type] [data]
```

### Contoh Penggunaan:

**1. Akun Spotify:**
```
/addstock 1 account premium.spotify@gmail.com | Spotify2024!
```

**2. Link Netflix:**
```
/addstock 2 link https://netflix.com/redeem?code=NFLX-ABC-1234
```

**3. Kode YouTube Premium:**
```
/addstock 3 code YOUTUBE-PREMIUM-XYZ789
```

**4. Custom (Game Account):**
```
/addstock 10 custom Player ID: 987654321 | Server: Asia | Level: 100
```

## 📖 Dokumentasi Lengkap

- **[MULTI_FORMAT_GUIDE.md](MULTI_FORMAT_GUIDE.md)** - Panduan lengkap dengan penjelasan detail
- **[MULTI_FORMAT_EXAMPLES.md](MULTI_FORMAT_EXAMPLES.md)** - Banyak contoh SQL dan use cases
- **[CHANGELOG_MULTIFORMAT.md](CHANGELOG_MULTIFORMAT.md)** - Daftar lengkap perubahan

## 🎯 Keuntungan

✅ **Fleksibel** - Tidak terbatas pada format email|password  
✅ **User-Friendly** - Instruksi spesifik untuk setiap format  
✅ **Backward Compatible** - Data lama tetap berfungsi  
✅ **Easy to Use** - Command sederhana untuk admin  
✅ **Scalable** - Mudah menambahkan format baru

## 💻 Untuk Developer

### Models Update:
```go
type ProductAccount struct {
    ContentType ProductContentType // account/link/code/custom
    ContentData string              // The actual content
    // ... other fields
}
```

### Menambah Stock via Code:
```go
// New method - recommended
db.AddProductContent(productID, "link", "https://example.com/redeem/ABC")

// Old method - still works but deprecated
db.AddProductAccount(productID, "user@email.com", "password")
```

### Display Content:
```go
account := getAccount()
label := account.GetContentLabel()    // Returns: 🔐 Akun
content := account.FormatContent()     // Returns: formatted content
```

## 🔄 Migration

Database akan **otomatis di-migrate** saat aplikasi dijalankan:
- Data lama (email|password) digabung ke `content_data`
- Field lama tetap ada untuk compatibility
- Tidak perlu action manual

## 🧪 Testing Checklist

- [ ] Test menambah stock format account
- [ ] Test menambah stock format link
- [ ] Test menambah stock format code
- [ ] Test menambah stock format custom
- [ ] Test pembelian produk dengan berbagai format
- [ ] Verifikasi tampilan di user side
- [ ] Verifikasi notifikasi admin
- [ ] Test backward compatibility dengan data lama

## 📞 Butuh Bantuan?

1. Cek [MULTI_FORMAT_GUIDE.md](MULTI_FORMAT_GUIDE.md) untuk panduan lengkap
2. Lihat [MULTI_FORMAT_EXAMPLES.md](MULTI_FORMAT_EXAMPLES.md) untuk contoh-contoh
3. Review [CHANGELOG_MULTIFORMAT.md](CHANGELOG_MULTIFORMAT.md) untuk detail perubahan

## ⚡ Tips & Tricks

**💡 Tip 1:** Gunakan format yang paling sesuai dengan jenis produk
- Akun streaming → `account`
- Gift card → `code`
- Redeem URL → `link`
- Game account dengan detail → `custom`

**💡 Tip 2:** Pastikan data yang diinput sudah benar sebelum menambahkan

**💡 Tip 3:** Gunakan format yang konsisten untuk produk sejenis

**💡 Tip 4:** Test dulu dengan 1-2 stock sebelum bulk insert

## 🎉 Selamat!

Multi-format support sekarang aktif! Nikmati fleksibilitas mengelola berbagai jenis produk digital. 🚀

---

**Version:** 2.0.0  
**Last Update:** 2025-10-27  
**Status:** ✅ Production Ready
