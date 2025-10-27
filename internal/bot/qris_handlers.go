package bot

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"telegram-premium-store/internal/qris"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// handleQRISSetup handles QRIS setup command for admins
func (b *Bot) handleQRISSetup(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
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

	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleQRISUpload handles QRIS image upload process
func (b *Bot) handleQRISUpload(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	text := `📤 *UPLOAD QR CODE STATIS*

Kirim gambar QR Code statis dari bank atau e-wallet Anda.

📋 *Langkah-langkah:*
1. Buka aplikasi bank/e-wallet Anda
2. Buat QR Code untuk menerima pembayaran
3. Screenshot atau download QR Code tersebut
4. Kirim gambar ke bot ini

⚠️ *Persyaratan:*
• Format: PNG atau JPEG
• Ukuran maksimal: 5MB
• QR Code harus jelas dan tidak blur
• Pastikan QR Code adalah milik Anda

Kirim gambar sekarang...`

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batal", "qris:setup"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)

	// Set user state to waiting for QRIS upload
	b.setUserState(callback.From.ID, "waiting_qris_upload")
}

// handleQRISTest handles QRIS test generation
func (b *Bot) handleQRISTest(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	if !b.realQRISService.IsConfigured() {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ QRIS belum dikonfigurasi"))
		return
	}

	// Generate test QRIS with 10000 amount
	testOrderID := "TEST-" + b.realQRISService.GenerateOrderID()
	testAmount := 10000

	qrisPayment, qrImage, err := b.realQRISService.GenerateDynamicQRIS(testOrderID, testAmount)
	if err != nil {
		logrus.Errorf("Failed to generate test QRIS: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal generate QRIS test"))
		return
	}

	// Send test QR code
	qrMsg := tgbotapi.NewPhoto(callback.Message.Chat.ID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("test_qris_%s.png", testOrderID),
		Bytes: qrImage,
	})

	qrMsg.Caption = fmt.Sprintf(`🧪 *TEST QRIS DINAMIS*

✅ QRIS berhasil digenerate!

🆔 Order ID: %s
💰 Nominal: Rp 10.000
⏰ Berlaku sampai: %s

🔍 *Informasi Teknis:*
• Payload Length: %d karakter
• Merchant: %s
• Expiry: 15 menit

💡 Coba scan dengan aplikasi e-wallet untuk memastikan QRIS berfungsi dengan baik.`,
		testOrderID,
		qrisPayment.ExpiryTime.Format("15:04:05"),
		len(qrisPayment.QRString),
		qrisPayment.MerchantName)

	qrMsg.ParseMode = tgbotapi.ModeMarkdown

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Setup QRIS", "qris:setup"),
		),
	)
	qrMsg.ReplyMarkup = keyboard

	b.api.Send(qrMsg)
	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Test QRIS generated"))
}

// handleQRISInfo handles QRIS information display
func (b *Bot) handleQRISInfo(callback *tgbotapi.CallbackQuery) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	var text strings.Builder
	text.WriteString("📋 *INFORMASI QRIS*\n\n")

	if b.realQRISService.IsConfigured() {
		merchantInfo := b.realQRISService.GetMerchantInfo()
		text.WriteString("✅ *Status: Dikonfigurasi*\n\n")
		text.WriteString(fmt.Sprintf("🏪 Merchant: %s\n", merchantInfo.MerchantName))
		text.WriteString(fmt.Sprintf("🏙️ Kota: %s\n", merchantInfo.MerchantCity))
		text.WriteString(fmt.Sprintf("🆔 ID: %s\n", merchantInfo.MerchantID))
		text.WriteString(fmt.Sprintf("🌍 Negara: %s\n", merchantInfo.CountryCode))
		text.WriteString(fmt.Sprintf("💱 Currency: %s\n\n", merchantInfo.Currency))

		supportedBanks := b.realQRISService.GetSupportedBanks()
		text.WriteString("📱 *Aplikasi yang Didukung:*\n")
		for i, bank := range supportedBanks {
			text.WriteString(fmt.Sprintf("• %s\n", bank))
			if i >= 9 { // Show first 10 only
				text.WriteString(fmt.Sprintf("• ... dan %d lainnya\n", len(supportedBanks)-10))
				break
			}
		}
	} else {
		text.WriteString("❌ *Status: Belum Dikonfigurasi*\n\n")
		text.WriteString("Upload QR Code statis terlebih dahulu untuk mengaktifkan QRIS dinamis.\n\n")
		text.WriteString("📋 *Cara Setup:*\n")
		text.WriteString("1. Klik 'Upload QR Statis'\n")
		text.WriteString("2. Kirim gambar QR Code dari bank/e-wallet\n")
		text.WriteString("3. Sistem akan mengekstrak payload otomatis\n")
		text.WriteString("4. QRIS dinamis siap digunakan\n")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Setup QRIS", "qris:setup"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleQRISImageUpload processes uploaded QRIS image
func (b *Bot) handleQRISImageUpload(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	// Check if user is in upload state
	if !b.isUserInState(message.From.ID, "waiting_qris_upload") {
		return
	}

	// Clear user state
	b.clearUserState(message.From.ID)

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	// Check if message has photo
	if message.Photo == nil || len(message.Photo) == 0 {
		b.sendMessage(message.Chat.ID, "❌ Kirim gambar QR Code yang valid!")
		return
	}

	// Get the largest photo
	photo := message.Photo[len(message.Photo)-1]

	// Download photo
	fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		logrus.Errorf("Failed to get file info: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal mengunduh gambar!")
		return
	}

	// Download file content
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", b.config.BotToken, file.FilePath)
	resp, err := http.Get(fileURL)
	if err != nil {
		logrus.Errorf("Failed to download file: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal mengunduh gambar!")
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("Failed to read image data: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal membaca data gambar!")
		return
	}

	// Validate image
	if err := b.realQRISService.ValidateQRISImage(imageData); err != nil {
		b.sendMessage(message.Chat.ID, fmt.Sprintf("❌ %s", err.Error()))
		return
	}

	// Send processing message
	msg := tgbotapi.NewMessage(message.Chat.ID, "🔄 Memproses QR Code... Mohon tunggu...")
	msg.ParseMode = tgbotapi.ModeMarkdown
	processingMsg, _ := b.api.Send(msg)

	// Process QRIS image
	filename := fmt.Sprintf("qris_static_%d.jpg", message.Date)
	err = b.realQRISService.UploadStaticQR(imageData, filename)
	if err != nil {
		logrus.Errorf("Failed to process QRIS image: %v", err)
		
		// Edit processing message with error
		editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, processingMsg.MessageID, 
			fmt.Sprintf("❌ Gagal memproses QR Code: %s\n\n💡 Pastikan gambar berisi QR Code QRIS yang valid.", err.Error()))
		b.api.Send(editMsg)
		return
	}

	// Success message
	merchantInfo := b.realQRISService.GetMerchantInfo()
	successText := fmt.Sprintf(`✅ *QR CODE BERHASIL DIPROSES!*

🏪 Merchant: %s
🏙️ Kota: %s  
🆔 ID: %s
💱 Currency: %s

🎉 QRIS dinamis sekarang sudah aktif! Sistem akan otomatis generate QR Code dengan nominal sesuai pesanan pelanggan.

💡 Gunakan /qristest untuk mencoba generate QRIS test.`,
		merchantInfo.MerchantName,
		merchantInfo.MerchantCity,
		merchantInfo.MerchantID,
		merchantInfo.Currency)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 Test Generate", "qris:test"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Setup QRIS", "qris:setup"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	editMsg := tgbotapi.NewEditMessageText(message.Chat.ID, processingMsg.MessageID, successText)
	editMsg.ParseMode = tgbotapi.ModeMarkdown
	editMsg.ReplyMarkup = &keyboard

	b.api.Send(editMsg)

	logrus.Info("✅ QRIS static QR successfully processed by admin")
}

// Add QRIS test command
func (b *Bot) handleQRISTestCommand(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	// Initialize real QRIS service if not already done
	if b.realQRISService == nil {
		b.realQRISService = qris.NewRealQRISService(b.config)
	}

	if !b.realQRISService.IsConfigured() {
		b.sendMessage(message.Chat.ID, "❌ QRIS belum dikonfigurasi. Gunakan /qrissetup untuk setup.")
		return
	}

	// Generate test QRIS
	testOrderID := "TEST-" + b.realQRISService.GenerateOrderID()
	testAmount := 10000

	qrisPayment, qrImage, err := b.realQRISService.GenerateDynamicQRIS(testOrderID, testAmount)
	if err != nil {
		logrus.Errorf("Failed to generate test QRIS: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal generate QRIS test!")
		return
	}

	// Send test QR code
	qrMsg := tgbotapi.NewPhoto(message.Chat.ID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("test_qris_%s.png", testOrderID),
		Bytes: qrImage,
	})

	qrMsg.Caption = fmt.Sprintf(`🧪 *TEST QRIS DINAMIS*

✅ QRIS berhasil digenerate!

🆔 Order ID: %s
💰 Nominal: Rp 10.000
⏰ Berlaku sampai: %s

💡 Coba scan dengan aplikasi e-wallet untuk memastikan QRIS berfungsi.`,
		testOrderID,
		qrisPayment.ExpiryTime.Format("15:04:05"))

	qrMsg.ParseMode = tgbotapi.ModeMarkdown

	b.api.Send(qrMsg)
}