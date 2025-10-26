package bot

import (
	"fmt"

	"telegram-premium-store/internal/qris"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleQRISCallback handles QRIS-related callback queries
func (b *Bot) handleQRISCallback(callback *tgbotapi.CallbackQuery, action string) {
	switch action {
	case "setup":
		b.handleQRISSetupCallback(callback)
	case "upload":
		b.handleQRISUpload(callback)
	case "test":
		b.handleQRISTest(callback)
	case "info":
		b.handleQRISInfo(callback)
	default:
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Aksi tidak dikenal"))
	}
}

// handleQRISSetupCallback shows QRIS setup menu
func (b *Bot) handleQRISSetupCallback(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	status := b.realQRISService.GetStaticQRStatus()
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📤 Upload QR Statis", "qris:upload"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔍 Test Generate", "qris:test"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Info QRIS", "qris:info"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	text := fmt.Sprintf(`🔧 *SETUP QRIS DINAMIS*

%s

📋 *Cara Setup:*
1. Upload QR Code statis dari bank/e-wallet Anda
2. Sistem akan mengekstrak payload QRIS
3. QR dinamis akan digenerate otomatis saat pembayaran

💡 *Tips:*
• Gunakan QR Code dari aplikasi bank/e-wallet
• Pastikan QR Code jelas dan tidak blur
• Format yang didukung: PNG, JPEG
• Maksimal ukuran file: 5MB`, status)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}