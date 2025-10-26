# 🔧 Panduan Setup QRIS Dinamis - Bot Telegram Premium Store

## 📋 Tentang QRIS Dinamis

Bot ini menggunakan sistem **QRIS Dinamis** yang memungkinkan generate QR Code dengan nominal yang berbeda-beda sesuai dengan pesanan pelanggan. Sistem ini bekerja dengan cara:

1. **Upload QR Statis** - Admin upload QR Code statis dari bank/e-wallet
2. **Ekstraksi Payload** - Sistem mengekstrak informasi merchant dari QR statis
3. **Generate Dinamis** - Saat ada pesanan, sistem generate QR Code baru dengan nominal sesuai pesanan

## 🚀 Cara Setup QRIS

### 1. **Persiapan QR Code Statis**

#### **Dari Bank (Recommended):**
- Buka aplikasi mobile banking (BCA Mobile, BNI Mobile, BRI Mobile, dll)
- Pilih menu "QRIS" atau "Terima Pembayaran"
- Generate QR Code untuk merchant
- Screenshot atau download QR Code

#### **Dari E-Wallet:**
- Buka aplikasi e-wallet (DANA, OVO, GoPay, dll)
- Pilih "Terima Uang" atau "QRIS"
- Generate QR Code merchant
- Screenshot QR Code

### 2. **Setup di Bot Telegram**

#### **Langkah 1: Akses Menu Admin**
```
/admin → 🔧 Setup QRIS
```

#### **Langkah 2: Upload QR Statis**
```
1. Klik "📤 Upload QR Statis"
2. Kirim gambar QR Code ke bot
3. Tunggu proses ekstraksi payload
4. Verifikasi informasi merchant
```

#### **Langkah 3: Test QRIS**
```
1. Klik "🔍 Test Generate" 
2. Bot akan generate QR Code test dengan nominal Rp 10.000
3. Scan dengan aplikasi e-wallet untuk test
```

### 3. **Verifikasi Setup**

Setelah upload berhasil, bot akan menampilkan:
```
✅ QRIS sudah dikonfigurasi
🏪 Merchant: Nama Toko Anda
🏙️ Kota: Jakarta
🆔 ID: ID1234567890123
💳 Currency: 360
```

## 🔍 Cara Kerja Sistem

### **Flow Pembayaran:**

1. **Pelanggan Checkout** → Bot generate QR Code dinamis
2. **QR Code Berisi:**
   - Informasi merchant (dari QR statis)
   - Nominal sesuai pesanan
   - Order ID untuk tracking
   - Expiry time (15 menit)

3. **Pelanggan Scan & Bayar** → Pembayaran otomatis terverifikasi

### **Contoh QR Code yang Digenerate:**

```
🆔 Order ID: ORD-abc12345-xyz
💰 Nominal: Rp 65.000
⏰ Berlaku sampai: 15:30:45
📱 Merchant: Premium Apps Store
```

## 📱 Aplikasi yang Didukung

### **Mobile Banking:**
- 🏦 BCA Mobile
- 🏦 BNI Mobile Banking  
- 🏦 BRI Mobile
- 🏦 Mandiri Online
- 🏦 CIMB Niaga
- 🏦 Permata Mobile
- 🏦 Danamon D-Bank
- 🏦 OCBC OneB

### **E-Wallet:**
- 💳 DANA
- 💳 OVO
- 💳 GoPay
- 💳 LinkAja
- 💳 ShopeePay
- 💳 Jenius
- 💳 Sakuku
- 💳 i.saku
- 💳 DOKU Wallet
- 💳 Flip
- 💳 Bibit

## ⚙️ Konfigurasi Teknis

### **Format File yang Didukung:**
- **PNG** (Recommended)
- **JPEG/JPG**
- **Maksimal 5MB**

### **Persyaratan QR Code:**
- ✅ QR Code harus jelas dan tidak blur
- ✅ Berisi payload QRIS yang valid
- ✅ Menggunakan standard EMV QR Code
- ✅ Memiliki identifier "ID.CO.QRIS"

### **Validasi Otomatis:**
Bot akan otomatis validasi:
- Format gambar
- Ukuran file
- Keberadaan QR Code dalam gambar
- Validitas payload QRIS
- Informasi merchant

## 🔧 Perintah Admin QRIS

### **Setup & Konfigurasi:**
```bash
/qrissetup    # Menu setup QRIS lengkap
/qristest     # Generate QR Code test
```

### **Via Menu Admin:**
```
/admin → Setup QRIS → Upload QR Statis
```

## 🛠️ Troubleshooting

### **❌ "Gagal memproses QR Code"**

**Penyebab & Solusi:**
1. **Gambar tidak jelas** → Upload gambar dengan resolusi lebih tinggi
2. **QR Code tidak valid** → Pastikan menggunakan QR QRIS dari bank/e-wallet
3. **Format tidak didukung** → Gunakan PNG atau JPEG
4. **File terlalu besar** → Kompres gambar di bawah 5MB

### **❌ "QR Code tidak berisi QRIS"**

**Solusi:**
- Pastikan QR Code dari aplikasi bank/e-wallet resmi
- Jangan gunakan QR Code untuk transfer antar individu
- Gunakan QR Code merchant/bisnis

### **❌ "Sistem pembayaran belum dikonfigurasi"**

**Solusi:**
1. Setup QRIS terlebih dahulu dengan `/qrissetup`
2. Upload QR Code statis yang valid
3. Verifikasi konfigurasi dengan test generate

### **❌ QR Code tidak bisa di-scan**

**Penyebab & Solusi:**
1. **QR Code expired** → Generate ulang (max 15 menit)
2. **Nominal tidak sesuai** → Jangan ubah nominal saat bayar
3. **Aplikasi tidak support** → Gunakan aplikasi yang didukung QRIS

## 📊 Monitoring & Maintenance

### **Cek Status QRIS:**
```
/admin → Setup QRIS → Info QRIS
```

### **File yang Disimpan:**
```
uploads/qris/
├── qris_static_*.jpg      # QR Code statis yang diupload
├── qris_config.txt        # Konfigurasi merchant
└── ...

generated/qris/
├── qris_ORD-*_*.png      # QR Code dinamis yang digenerate
└── ...
```

### **Backup Konfigurasi:**
```bash
# Backup file konfigurasi QRIS
cp uploads/qris/qris_config.txt backup/qris_config_$(date +%Y%m%d).txt
```

## 🔒 Keamanan & Best Practices

### **Keamanan:**
- ✅ QR Code statis hanya disimpan lokal
- ✅ Payload QRIS di-encrypt dalam database
- ✅ QR Code dinamis auto-expire dalam 15 menit
- ✅ Validasi merchant info setiap generate

### **Best Practices:**
1. **Gunakan QR Code dari akun bisnis** (bukan personal)
2. **Update QR statis secara berkala** (3-6 bulan)
3. **Monitor transaksi** melalui aplikasi bank/e-wallet
4. **Backup konfigurasi** secara rutin
5. **Test generate** setelah update QR statis

## 🆘 Support & Bantuan

### **Jika Mengalami Masalah:**

1. **Cek log aplikasi:**
```bash
# Lihat log bot
make logs

# Atau jika menggunakan systemd
sudo journalctl -u telegram-store-bot -f
```

2. **Reset konfigurasi QRIS:**
```bash
# Hapus konfigurasi lama
rm -rf uploads/qris/qris_config.txt

# Setup ulang
/qrissetup
```

3. **Hubungi developer** jika masalah persisten

## 📈 Upgrade & Update

### **Update Library QRIS:**
```bash
# Update go-qris library
go get -u github.com/fyvri/go-qris

# Rebuild aplikasi
make build
```

### **Migrasi dari Mock ke Real QRIS:**
Bot otomatis akan menggunakan Real QRIS jika sudah dikonfigurasi, dan fallback ke mock jika belum.

---

## 🎯 Tips Sukses

1. **Pilih Bank/E-Wallet Utama** - Gunakan yang paling sering dipakai pelanggan
2. **Test Berkala** - Generate test QR Code setiap minggu
3. **Monitor Transaksi** - Cek aplikasi bank/e-wallet untuk konfirmasi pembayaran
4. **Educate Customer** - Berikan panduan pembayaran QRIS ke pelanggan
5. **Keep Backup** - Simpan screenshot QR statis sebagai backup

**🎉 Selamat! QRIS Dinamis siap digunakan untuk menerima pembayaran dari pelanggan!**