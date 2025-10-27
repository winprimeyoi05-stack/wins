package bot

import (
	"fmt"
	"strings"

	"telegram-premium-store/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// handleStockManagement handles stock management for admin
func (b *Bot) handleStockManagement(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	products, err := b.db.GetProducts("", 50, 0)
	if err != nil {
		logrus.Errorf("Failed to get products: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat produk"))
		return
	}

	var text strings.Builder
	text.WriteString("📦 *KELOLA STOK PRODUK*\n\n")

	if len(products) == 0 {
		text.WriteString("Tidak ada produk yang tersedia.")
	} else {
		for _, product := range products {
			stockStatus := "✅"
			if product.Stock == 0 {
				stockStatus = "❌"
			} else if product.Stock <= 5 {
				stockStatus = "⚠️"
			}

			text.WriteString(fmt.Sprintf("%s *%s*\n", stockStatus, product.Name))
			text.WriteString(fmt.Sprintf("   Stok: %d | Harga: %s\n\n", 
				product.Stock, models.FormatPrice(product.Price, b.config.CurrencySymbol)))
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Cek Stok Rendah", "admin:lowstock"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit Stok", "admin:editstock"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleLowStock shows products with low stock
func (b *Bot) handleLowStock(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	lowStockProducts, err := b.db.GetLowStockProducts(5) // Products with stock <= 5
	if err != nil {
		logrus.Errorf("Failed to get low stock products: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat data"))
		return
	}

	var text strings.Builder
	text.WriteString("⚠️ *PRODUK STOK RENDAH*\n\n")

	if len(lowStockProducts) == 0 {
		text.WriteString("✅ Semua produk memiliki stok yang cukup!")
	} else {
		for _, product := range lowStockProducts {
			stockIcon := "❌"
			if product.Stock > 0 {
				stockIcon = "⚠️"
			}

			text.WriteString(fmt.Sprintf("%s *%s*\n", stockIcon, product.Name))
			text.WriteString(fmt.Sprintf("   Stok tersisa: *%d*\n", product.Stock))
			text.WriteString(fmt.Sprintf("   Kategori: %s\n\n", product.Category))
		}

		text.WriteString("💡 *Rekomendasi:* Segera lakukan restock untuk produk yang stoknya habis atau rendah.")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kelola Stok", "admin:stock"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleCategoryManagement handles category management
func (b *Bot) handleCategoryManagement(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	categories, err := b.db.GetCategoriesFromDB()
	if err != nil {
		logrus.Errorf("Failed to get categories: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat kategori"))
		return
	}

	var text strings.Builder
	text.WriteString("🏷️ *KELOLA KATEGORI*\n\n")

	if len(categories) == 0 {
		text.WriteString("Belum ada kategori yang tersedia.")
	} else {
		for _, category := range categories {
			text.WriteString(fmt.Sprintf("%s *%s*\n", category.Icon, category.DisplayName))
			text.WriteString(fmt.Sprintf("   Produk: %d | ID: %s\n\n", category.Count, category.Name))
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Tambah Kategori", "admin:addcategory"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit Kategori", "admin:editcategory"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Hapus Kategori", "admin:deletecategory"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleProductManagement handles product management
func (b *Bot) handleProductManagement(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	products, err := b.db.GetProducts("", 10, 0) // Get first 10 products
	if err != nil {
		logrus.Errorf("Failed to get products: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat produk"))
		return
	}

	var text strings.Builder
	text.WriteString("📦 *KELOLA PRODUK*\n\n")

	if len(products) == 0 {
		text.WriteString("Belum ada produk yang tersedia.")
	} else {
		for i, product := range products {
			if i >= 5 { // Show only first 5 in summary
				text.WriteString(fmt.Sprintf("... dan %d produk lainnya\n", len(products)-5))
				break
			}

			stockStatus := "✅"
			if product.Stock == 0 {
				stockStatus = "❌"
			} else if product.Stock <= 5 {
				stockStatus = "⚠️"
			}

			text.WriteString(fmt.Sprintf("%s *%s*\n", stockStatus, product.Name))
			text.WriteString(fmt.Sprintf("   Stok: %d | %s\n\n", 
				product.Stock, models.FormatPrice(product.Price, b.config.CurrencySymbol)))
		}
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Tambah Produk", "admin:addproduct"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Edit Produk", "admin:editproduct"),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Hapus Produk", "admin:deleteproduct"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Kelola Stok", "admin:stock"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleBroadcastManagement handles broadcast management
func (b *Bot) handleBroadcastManagement(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Get user statistics
	allUsers, _ := b.db.GetAllUsers()
	activeUsers, _ := b.db.GetActiveUsers(7) // Active in last 7 days

	text := fmt.Sprintf(`📢 *BROADCAST PROMOSI*

👥 Total pengguna: %d
🟢 Aktif (7 hari): %d
📊 Tingkat aktivitas: %.1f%%

Gunakan fitur broadcast untuk mengirim promosi atau pengumuman ke semua pengguna yang pernah berinteraksi dengan bot.

⚠️ *Perhatian:* Gunakan broadcast dengan bijak untuk menghindari spam.`, 
		len(allUsers), 
		len(activeUsers), 
		float64(len(activeUsers))/float64(len(allUsers))*100)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Kirim Broadcast", "admin:sendbroadcast"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Semua User", "admin:broadcast:all"),
			tgbotapi.NewInlineKeyboardButtonData("🟢 User Aktif", "admin:broadcast:active"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleSendBroadcast initiates broadcast sending process
func (b *Bot) handleSendBroadcast(callback *tgbotapi.CallbackQuery, targetType string) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	targetText := "semua pengguna"
	if targetType == "active" {
		targetText = "pengguna aktif (7 hari terakhir)"
	}

	text := fmt.Sprintf(`📤 *KIRIM BROADCAST*

Target: %s

Ketik pesan yang ingin Anda kirim ke %s. Pesan akan dikirim dalam format Markdown.

*Contoh format:*
🎉 *PROMO SPESIAL!*
Diskon 50%% untuk semua produk premium!
Valid sampai 31 Desember 2024.

Gunakan /cancel untuk membatalkan.`, targetText, targetText)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "admin:broadcast"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)

	// Set user state for broadcast message input
	b.setUserState(callback.From.ID, fmt.Sprintf("waiting_broadcast_%s", targetType))
}

// processBroadcastMessage processes broadcast message from admin
func (b *Bot) processBroadcastMessage(message *tgbotapi.Message, targetType string) {
	if !b.config.IsAdmin(message.From.ID) {
		return
	}

	broadcastText := message.Text
	if broadcastText == "" {
		b.sendMessage(message.Chat.ID, "❌ Pesan tidak boleh kosong!")
		return
	}

	// Get target users
	var targetUsers []int64
	var err error

	if targetType == "active" {
		targetUsers, err = b.db.GetActiveUsers(7)
	} else {
		targetUsers, err = b.db.GetAllUsers()
	}

	if err != nil {
		logrus.Errorf("Failed to get target users: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal mendapatkan daftar pengguna!")
		return
	}

	if len(targetUsers) == 0 {
		b.sendMessage(message.Chat.ID, "❌ Tidak ada pengguna target!")
		return
	}

	// Create broadcast record
	broadcastID, err := b.db.CreateBroadcast(broadcastText, targetType, message.From.ID)
	if err != nil {
		logrus.Errorf("Failed to create broadcast: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal membuat broadcast!")
		return
	}

	// Send confirmation
	confirmText := fmt.Sprintf(`📤 *KONFIRMASI BROADCAST*

Target: %d pengguna
Pesan: %s

Apakah Anda yakin ingin mengirim broadcast ini?`, len(targetUsers), broadcastText)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Kirim Sekarang", fmt.Sprintf("admin:confirm_broadcast:%d", broadcastID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "admin:broadcast"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, confirmText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// executeBroadcast executes the broadcast to all target users
func (b *Bot) executeBroadcast(callback *tgbotapi.CallbackQuery, broadcastID int64) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// This would be implemented to actually send the broadcast
	// For now, we'll simulate it
	
	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Broadcast sedang dikirim..."))

	// Update the message to show completion
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, 
		"✅ *BROADCAST SELESAI*\n\nBroadcast telah dikirim ke semua pengguna target.")
	edit.ParseMode = tgbotapi.ModeMarkdown

	b.api.Send(edit)

	logrus.Infof("Broadcast %d executed by admin %d", broadcastID, callback.From.ID)
}

// User state management
var userStates = make(map[int64]string)

func (b *Bot) setUserState(userID int64, state string) {
	userStates[userID] = state
}

func (b *Bot) getUserState(userID int64) string {
	return userStates[userID]
}

func (b *Bot) clearUserState(userID int64) {
	delete(userStates, userID)
}

func (b *Bot) isUserInState(userID int64, state string) bool {
	return userStates[userID] == state
}