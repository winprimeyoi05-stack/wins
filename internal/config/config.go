package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	// Bot Configuration
	BotToken    string
	AdminIDs    []int64
	BotUsername string

	// Database Configuration
	DatabasePath string

	// QRIS Payment Configuration
	QRISMerchantID     string
	QRISMerchantName   string
	QRISCity           string
	QRISCountryCode    string
	QRISCurrencyCode   string
	QRISTransactionFee int

	// Server Configuration
	ServerPort string
	WebhookURL string

	// Business Configuration
	StoreName        string
	StoreDescription string
	AdminUsername    string
	AdminEmail       string
	SupportPhone     string
	BusinessHours    string
	Timezone         string

	// Application Settings
	LogLevel        string
	MaxCartItems    int
	CurrencySymbol  string
	DefaultImageURL string

	// Payment Security
	PaymentSecretKey string
}

// Messages contains all Indonesian messages for the bot
type Messages struct {
	Welcome     string
	Help        string
	Contact     string
	CartEmpty   string
	OrderSuccess string
	PaymentInstructions string
}

// Load creates a new Config instance from environment variables
func Load() *Config {
	return &Config{
		// Bot Configuration
		BotToken:    getEnv("BOT_TOKEN", ""),
		AdminIDs:    parseAdminIDs(getEnv("ADMIN_IDS", "")),
		BotUsername: getEnv("BOT_USERNAME", "premium_store_bot"),

		// Database Configuration
		DatabasePath: getEnv("DATABASE_PATH", "store.db"),

		// QRIS Configuration
		QRISMerchantID:     getEnv("QRIS_MERCHANT_ID", "ID1234567890123"),
		QRISMerchantName:   getEnv("QRIS_MERCHANT_NAME", "Premium Store"),
		QRISCity:           getEnv("QRIS_CITY", "Jakarta"),
		QRISCountryCode:    getEnv("QRIS_COUNTRY_CODE", "ID"),
		QRISCurrencyCode:   getEnv("QRIS_CURRENCY_CODE", "360"),
		QRISTransactionFee: getEnvAsInt("QRIS_TRANSACTION_FEE", 0),

		// Server Configuration
		ServerPort: getEnv("SERVER_PORT", "8080"),
		WebhookURL: getEnv("WEBHOOK_URL", ""),

		// Business Configuration
		StoreName:        getEnv("STORE_NAME", "Premium Apps Store"),
		StoreDescription: getEnv("STORE_DESCRIPTION", "Toko aplikasi premium terpercaya"),
		AdminUsername:    getEnv("ADMIN_USERNAME", "admin"),
		AdminEmail:       getEnv("ADMIN_EMAIL", "admin@premiumstore.com"),
		SupportPhone:     getEnv("SUPPORT_PHONE", "+6281234567890"),
		BusinessHours:    getEnv("BUSINESS_HOURS", "08:00 - 22:00 WIB"),
		Timezone:         getEnv("TIMEZONE", "Asia/Jakarta"),

		// Application Settings
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		MaxCartItems:    getEnvAsInt("MAX_CART_ITEMS", 10),
		CurrencySymbol:  getEnv("CURRENCY_SYMBOL", "Rp"),
		DefaultImageURL: getEnv("DEFAULT_PRODUCT_IMAGE", "https://via.placeholder.com/300x200?text=Premium+App"),

		// Payment Security
		PaymentSecretKey: getEnv("PAYMENT_SECRET_KEY", ""),
	}
}

// GetMessages returns all Indonesian messages
func GetMessages() *Messages {
	return &Messages{
		Welcome: `🎉 *Selamat datang di Premium Apps Store!* 🎉

Kami menyediakan berbagai aplikasi premium berkualitas tinggi dengan harga terjangkau dan pembayaran mudah melalui QRIS.

📱 *Fitur Unggulan:*
• Katalog aplikasi premium lengkap
• Pembayaran QRIS dinamis & aman
• Sistem keranjang belanja
• Support 24/7
• Garansi aplikasi

Ketik /help untuk melihat semua perintah yang tersedia.`,

		Help: `📋 *DAFTAR PERINTAH:*

🏠 /start - Mulai menggunakan bot
📱 /catalog - Lihat katalog aplikasi
🛒 /cart - Lihat keranjang belanja
💰 /history - Riwayat pembelian
💳 /payment - Status pembayaran
📞 /contact - Hubungi admin
ℹ️ /help - Bantuan

👨‍💼 *PERINTAH ADMIN:*
/admin - Panel admin
/addproduct - Tambah produk baru
/users - Lihat daftar pengguna
/orders - Kelola pesanan
/stats - Statistik penjualan`,

		Contact: `📞 *HUBUNGI KAMI:*

👨‍💼 Admin: @%s
📧 Email: %s
📱 WhatsApp: %s
⏰ Jam Operasional: %s

💬 Untuk pertanyaan lebih lanjut, silakan hubungi admin di atas.
🔄 Atau gunakan menu bantuan dengan mengetik /help`,

		CartEmpty: `🛒 *KERANJANG BELANJA*

Keranjang Anda kosong.
Silakan pilih produk dari katalog terlebih dahulu.

📱 Gunakan /catalog untuk melihat produk yang tersedia.`,

		OrderSuccess: `✅ *PESANAN BERHASIL DIBUAT!*

🆔 Order ID: #%s
💰 Total: %s %s
📅 Tanggal: %s

Silakan lakukan pembayaran melalui QRIS yang telah digenerate.
Pembayaran akan otomatis terverifikasi setelah berhasil.`,

		PaymentInstructions: `💳 *INSTRUKSI PEMBAYARAN QRIS*

1️⃣ Scan QR Code di bawah ini dengan aplikasi e-wallet atau mobile banking Anda
2️⃣ Pastikan nominal pembayaran sesuai: *%s %s*
3️⃣ Konfirmasi pembayaran di aplikasi Anda
4️⃣ Pembayaran akan otomatis terverifikasi dalam 1-2 menit
5️⃣ Anda akan menerima notifikasi setelah pembayaran berhasil

⚠️ *PENTING:*
• QR Code ini hanya berlaku untuk 15 menit
• Jangan ubah nominal pembayaran
• Simpan Order ID untuk referensi: *#%s*`,
	}
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func parseAdminIDs(adminIDsStr string) []int64 {
	var adminIDs []int64
	if adminIDsStr == "" {
		return adminIDs
	}

	ids := strings.Split(adminIDsStr, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			adminIDs = append(adminIDs, id)
		}
	}
	return adminIDs
}

// IsAdmin checks if the given user ID is an admin
func (c *Config) IsAdmin(userID int64) bool {
	for _, adminID := range c.AdminIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}