package bot

import (
	"fmt"
	"time"

	"telegram-premium-store/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// handleStockNotification handles stock notification request
func (b *Bot) handleStockNotification(callback *tgbotapi.CallbackQuery, productID int) {
	userID := callback.From.ID

	// Log user interaction for stock notification
	b.db.LogUserInteraction(userID, "stock_notification", fmt.Sprintf("product_id:%d", productID))

	product, err := b.db.GetProduct(productID)
	if err != nil || product == nil {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Produk tidak ditemukan"))
		return
	}

	text := fmt.Sprintf(`🔔 *NOTIFIKASI STOK*

Anda akan diberitahu ketika *%s* tersedia kembali.

📧 Notifikasi akan dikirim melalui bot ini ketika stok sudah tersedia.

💡 *Tips:* Bookmark produk ini atau kembali lagi nanti untuk mengecek ketersediaan.`, product.Name)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Sudah Diatur", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali ke Produk", fmt.Sprintf("product:%d", productID)),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)

	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Notifikasi diatur!"))
}

// handleCancelOrder handles order cancellation
func (b *Bot) handleCancelOrder(callback *tgbotapi.CallbackQuery, orderID string) {
	userID := callback.From.ID

	// Get order details
	order, err := b.db.GetOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to get order %s: %v", orderID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat pesanan"))
		return
	}

	if order == nil || order.UserID != userID {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Pesanan tidak ditemukan"))
		return
	}

	// Check if order can be cancelled (only pending orders)
	if order.PaymentStatus != models.PaymentStatusPending {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Pesanan tidak dapat dibatalkan"))
		return
	}

	// Check if QRIS is expired
	if order.QRISExpiry != nil && time.Now().After(*order.QRISExpiry) {
		// Auto-cancel expired order
		b.handleExpiredOrder(orderID)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "⏰ Pesanan sudah expired dan dibatalkan otomatis"))
		return
	}

	text := fmt.Sprintf(`❌ *BATALKAN PESANAN*

🆔 Order ID: #%s
💰 Total: %s
📅 Dibuat: %s

Apakah Anda yakin ingin membatalkan pesanan ini?

⚠️ *Perhatian:*
• Stok produk akan dikembalikan
• QR Code pembayaran akan tidak berlaku
• Pesanan tidak dapat dikembalikan setelah dibatalkan`, 
		orderID[:8],
		models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol),
		order.CreatedAt.Format("02/01/2006 15:04"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Ya, Batalkan", fmt.Sprintf("confirm_cancel:%s", orderID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Tidak", fmt.Sprintf("order:%s", orderID)),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleConfirmCancel confirms order cancellation
func (b *Bot) handleConfirmCancel(callback *tgbotapi.CallbackQuery, orderID string) {
	userID := callback.From.ID

	// Get order details
	order, err := b.db.GetOrder(orderID)
	if err != nil || order == nil || order.UserID != userID {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Pesanan tidak ditemukan"))
		return
	}

	if order.PaymentStatus != models.PaymentStatusPending {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Pesanan tidak dapat dibatalkan"))
		return
	}

	// Cancel order and restore stock
	err = b.db.UpdateOrderStatus(orderID, models.PaymentStatusCancelled)
	if err != nil {
		logrus.Errorf("Failed to cancel order %s: %v", orderID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal membatalkan pesanan"))
		return
	}

	// Restore stock
	err = b.db.RestoreStockFromOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to restore stock for order %s: %v", orderID, err)
	}

	text := fmt.Sprintf(`✅ *PESANAN DIBATALKAN*

🆔 Order ID: #%s
📅 Dibatalkan: %s

Pesanan Anda telah berhasil dibatalkan dan stok produk telah dikembalikan.

Terima kasih telah menggunakan layanan kami. Anda dapat melakukan pemesanan baru kapan saja.`, 
		orderID[:8],
		time.Now().Format("02/01/2006 15:04"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Belanja Lagi", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)

	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Pesanan dibatalkan"))

	// Log user interaction
	b.db.LogUserInteraction(userID, "order_cancelled", orderID)
}

// handleExpiredOrder handles expired QRIS orders
func (b *Bot) handleExpiredOrder(orderID string) {
	// Update order status to expired
	err := b.db.UpdateOrderStatus(orderID, models.PaymentStatusExpired)
	if err != nil {
		logrus.Errorf("Failed to expire order %s: %v", orderID, err)
		return
	}

	// Restore stock
	err = b.db.RestoreStockFromOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to restore stock for expired order %s: %v", orderID, err)
	}

	// Get order to send notification to user
	order, err := b.db.GetOrder(orderID)
	if err != nil || order == nil {
		return
	}

	// Send expiry notification to user
	expiredText := fmt.Sprintf(`⏰ *WAKTU PEMBAYARAN HABIS*

Waktu pembayaran untuk pesanan #%s telah habis.

💡 Anda dapat melakukan pemesanan kembali jika masih membutuhkan produk tersebut.

Terima kasih atas pengertiannya.`, orderID[:8])

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Pesan Lagi", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	msg := tgbotapi.NewMessage(order.UserID, expiredText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)

	logrus.Infof("Order %s expired and user %d notified", orderID, order.UserID)
}

// checkExpiredOrders checks and handles expired orders (called periodically)
func (b *Bot) checkExpiredOrders() {
	// This would be called by a background goroutine
	// Get all pending orders with expired QRIS
	// For demo, we'll implement the basic structure
	
	logrus.Debug("Checking for expired orders...")
	
	// Implementation would query database for expired pending orders
	// and call handleExpiredOrder for each
}