# 📚 Contoh Penggunaan Multi-Format Produk

## Contoh SQL untuk Menambahkan Berbagai Format Produk

### 1. Spotify Premium - Mixed Formats

```sql
-- Format Account
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (1, 'account', 'spotify.user1@gmail.com | SpotifyPass123!');

INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (1, 'account', 'premium.spotify2@gmail.com | MusicLover456@');

-- Format Code
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (1, 'code', 'SPOTIFY-PREMIUM-ABC123-XYZ789');

INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (1, 'code', 'SPOTIFY-GIFT-CODE-999888');
```

### 2. Netflix Premium - Link & Account

```sql
-- Format Link
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (2, 'link', 'https://netflix.com/redeem?code=NFLX-ABCD-1234-EFGH');

INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (2, 'link', 'https://netflix.com/redeem?code=NFLX-WXYZ-5678-IJKL');

-- Format Account
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (2, 'account', 'netflix.premium@gmail.com | NetflixHD2024!');
```

### 3. Game Items - Custom Format

```sql
-- Mobile Legends Diamond
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (10, 'custom', 'User ID: 123456789 (1234) | Zone: Asia | Diamonds: 5000');

-- Free Fire Voucher
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (11, 'code', 'FREEFIRE-VOUCHER-ABCD1234');

-- PUBG UC Top Up
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (12, 'custom', 'Player ID: 987654321 | Server: Asia | UC: 3000 + Bonus 300');
```

### 4. Microsoft Office 365 - License Key

```sql
-- Format Code untuk License Key
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (6, 'code', 'XXXXX-XXXXX-XXXXX-XXXXX-XXXXX');

INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (6, 'code', 'M365-PREMIUM-LICENSE-KEY-12345');
```

### 5. VPN Premium - Mixed Formats

```sql
-- Format Account
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (15, 'account', 'vpnuser@example.com | SecureVPN2024!');

-- Format Link (Activation Link)
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (15, 'link', 'https://vpnpremium.com/activate?key=ABC123DEF456GHI789');

-- Format Code (Subscription Code)
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (15, 'code', 'VPN-PREMIUM-1YEAR-XYZ999');
```

### 6. Adobe Creative Cloud - Link & Code

```sql
-- Format Link untuk Redeem
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (5, 'link', 'https://adobe.com/activate?code=ADOBE-CC-PREMIUM-ABC123');

-- Format Code untuk License
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (5, 'code', 'ADOBE-CC-LICENSE-2024-XYZ-12345');
```

### 7. E-Wallet Top Up - Custom Format

```sql
-- GoPay
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (20, 'custom', 'Nomor: 081234567890 | Nominal: Rp 100.000 | Metode: Transfer');

-- OVO
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (21, 'code', 'OVO-TOPUP-CODE-ABC123XYZ');

-- DANA
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (22, 'link', 'https://link.dana.id/topup?ref=DANA123456789&amount=100000');
```

### 8. Streaming Services - Various Formats

```sql
-- Disney+ Hotstar
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (30, 'account', 'disney.premium@gmail.com | DisneyPlus2024!');

-- Amazon Prime Video
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (31, 'code', 'PRIME-VIDEO-GIFT-CODE-ABCD1234EFGH5678');

-- Apple TV+
INSERT INTO product_accounts (product_id, content_type, content_data)
VALUES (32, 'link', 'https://tv.apple.com/redeem?code=APPLETV-PREMIUM-XYZ789');
```

## 🔧 Query Helper untuk Admin

### Melihat Stock Berdasarkan Format

```sql
-- Lihat semua stock account format
SELECT p.name, pa.content_type, pa.content_data, pa.is_sold
FROM product_accounts pa
JOIN products p ON pa.product_id = p.id
WHERE pa.content_type = 'account'
ORDER BY p.name, pa.created_at;

-- Lihat stock link format yang belum terjual
SELECT p.name, pa.content_data
FROM product_accounts pa
JOIN products p ON pa.product_id = p.id
WHERE pa.content_type = 'link' AND pa.is_sold = FALSE
ORDER BY p.name;

-- Hitung jumlah stock per format untuk setiap produk
SELECT 
    p.name AS product_name,
    pa.content_type,
    COUNT(*) AS total_stock,
    SUM(CASE WHEN pa.is_sold = FALSE THEN 1 ELSE 0 END) AS available_stock,
    SUM(CASE WHEN pa.is_sold = TRUE THEN 1 ELSE 0 END) AS sold_stock
FROM product_accounts pa
JOIN products p ON pa.product_id = p.id
GROUP BY p.name, pa.content_type
ORDER BY p.name, pa.content_type;
```

### Bulk Insert dari CSV/Excel

```sql
-- Template untuk bulk insert
-- Format: product_id, content_type, content_data

INSERT INTO product_accounts (product_id, content_type, content_data) VALUES
(1, 'account', 'user1@gmail.com | pass123'),
(1, 'account', 'user2@gmail.com | pass456'),
(1, 'code', 'CODE-ABC-123'),
(2, 'link', 'https://example.com/redeem?code=XYZ'),
(2, 'link', 'https://example.com/redeem?code=ABC');
```

### Update Format Lama ke Format Baru

```sql
-- Migrasi data lama yang masih menggunakan email & password terpisah
UPDATE product_accounts
SET 
    content_type = 'account',
    content_data = email || ' | ' || password
WHERE content_data IS NULL OR content_data = ''
AND email IS NOT NULL 
AND password IS NOT NULL;

-- Verifikasi hasil migrasi
SELECT 
    id,
    product_id,
    content_type,
    content_data,
    email,
    password
FROM product_accounts
WHERE content_type = 'account'
LIMIT 10;
```

## 🎯 Use Cases per Industri

### E-Learning Platform
```sql
-- Coursera, Udemy, Skillshare
INSERT INTO product_accounts (product_id, content_type, content_data) VALUES
(40, 'account', 'coursera.student@gmail.com | LearnPassword123!'),
(41, 'code', 'UDEMY-COURSE-ACCESS-ABC123'),
(42, 'link', 'https://skillshare.com/redeem/premium-year-xyz789');
```

### Gaming Platform
```sql
-- Steam, Epic Games, Xbox
INSERT INTO product_accounts (product_id, content_type, content_data) VALUES
(50, 'code', 'STEAM-WALLET-CODE-XXXXX-XXXXX-XXXXX'),
(51, 'link', 'https://epicgames.com/redeem?code=EPIC-PREMIUM-ABC'),
(52, 'custom', 'Xbox Gamertag: Player123 | Game Pass Ultimate: 12 Months');
```

### Music Streaming
```sql
-- Spotify, Apple Music, YouTube Music
INSERT INTO product_accounts (product_id, content_type, content_data) VALUES
(60, 'account', 'spotify.premium@gmail.com | MusicLover123!'),
(61, 'code', 'APPLEMUSIC-GIFT-CARD-XYZ789ABC'),
(62, 'link', 'https://music.youtube.com/redeem?code=YTMUSIC-PREMIUM');
```

## 📊 Dashboard Admin Query

```sql
-- Stock Overview per Format
SELECT 
    content_type,
    COUNT(*) AS total,
    SUM(CASE WHEN is_sold = FALSE THEN 1 ELSE 0 END) AS available,
    SUM(CASE WHEN is_sold = TRUE THEN 1 ELSE 0 END) AS sold,
    ROUND(SUM(CASE WHEN is_sold = TRUE THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) AS sold_percentage
FROM product_accounts
GROUP BY content_type;

-- Top Selling Products by Format
SELECT 
    p.name,
    pa.content_type,
    COUNT(*) AS sold_count
FROM sold_accounts sa
JOIN products p ON sa.product_id = p.id
JOIN product_accounts pa ON sa.account_id = pa.id
WHERE sa.sold_at >= datetime('now', '-30 days')
GROUP BY p.name, pa.content_type
ORDER BY sold_count DESC
LIMIT 10;
```

---

**Tips:** Simpan query-query ini untuk mempermudah management produk harian!
