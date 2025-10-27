package bot

import (
	"fmt"
	"strings"
	"time"

	"telegram-premium-store/internal/models"
	"telegram-premium-store/internal/payment"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// handlePaymentSuccess processes successful payment and delivers accounts to buyer
func (b *Bot) handlePaymentSuccess(orderID string, paidAmount int) error {
	// Get order details
	order, err := b.db.GetOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to get order %s: %v", orderID, err)
		return fmt.Errorf("failed to get order: %w", err)
	}

	if order == nil {
		return fmt.Errorf("order not found: %s", orderID)
	}

	// Verify payment amount against order amount
	verifier := payment.NewPaymentVerifier(b.config.PaymentSecretKey)
	verification, err := b.db.GetPaymentVerification(orderID)
	if err != nil {
		logrus.Errorf("Failed to get payment verification for order %s: %v", orderID, err)
		return fmt.Errorf("failed to get payment verification: %w", err)
	}

	if verification != nil {
		// Validate that paid amount matches expected amount
		if paidAmount != verification.ExpectedAmount {
			logrus.Errorf("Payment amount mismatch for order %s: expected %d, got %d", 
				orderID, verification.ExpectedAmount, paidAmount)
			
			// Send notification to admin about manipulation attempt
			b.notifyAdminManipulationAttempt(orderID, verification.ExpectedAmount, paidAmount, order.UserID)
			
			return fmt.Errorf("payment amount manipulation detected: expected %d, got %d", 
				verification.ExpectedAmount, paidAmount)
		}

		// Validate QRIS integrity if payload is provided
		if verification.QRISPayload != "" {
			if err := verifier.ValidateQRISIntegrity(verification.QRISPayload); err != nil {
				logrus.Errorf("QRIS integrity validation failed for order %s: %v", orderID, err)
				return fmt.Errorf("QRIS integrity validation failed: %w", err)
			}
		}

		// Mark payment as verified
		if err := b.db.MarkPaymentVerified(orderID); err != nil {
			logrus.Errorf("Failed to mark payment verified for order %s: %v", orderID, err)
		}
	}

	// Check if order is already paid
	if order.PaymentStatus == models.PaymentStatusPaid {
		logrus.Warnf("Order %s already marked as paid", orderID)
		return nil
	}

	// Update order status to paid
	if err := b.db.UpdateOrderStatus(orderID, models.PaymentStatusPaid); err != nil {
		logrus.Errorf("Failed to update order status for %s: %v", orderID, err)
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Get assigned accounts for this order
	soldAccounts, err := b.db.GetProductAccountsForOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to get accounts for order %s: %v", orderID, err)
		return fmt.Errorf("failed to get accounts: %w", err)
	}

	// Get buyer information
	buyer, err := b.db.GetUser(order.UserID)
	if err != nil {
		logrus.Errorf("Failed to get buyer info for user %d: %v", order.UserID, err)
	}

	// Send success message with accounts to buyer
	if err := b.sendAccountsToBuyer(order, soldAccounts); err != nil {
		logrus.Errorf("Failed to send accounts to buyer for order %s: %v", orderID, err)
		return fmt.Errorf("failed to send accounts to buyer: %w", err)
	}

	// Send comprehensive notification to admin
	b.sendAdminSaleNotification(order, soldAccounts, buyer, paidAmount)

	logrus.Infof("✅ Payment successful for order %s, accounts delivered to user %d", orderID, order.UserID)
	return nil
}

// sendAccountsToBuyer sends purchased accounts to the buyer with copy functionality
func (b *Bot) sendAccountsToBuyer(order *models.Order, accounts []models.SoldAccount) error {
	var message strings.Builder
	message.WriteString("🎉 *PEMBAYARAN BERHASIL!*\n\n")
	message.WriteString(fmt.Sprintf("✅ Pembayaran Anda untuk Order #%s telah dikonfirmasi.\n\n", order.ID[:8]))
	message.WriteString(fmt.Sprintf("💰 Total Pembayaran: %s\n", models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol)))
	message.WriteString(fmt.Sprintf("📅 Tanggal: %s\n\n", time.Now().Format("02/01/2006 15:04")))
	message.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")
	message.WriteString("🔐 *AKUN PREMIUM ANDA:*\n\n")

	// Group accounts by product
	productAccounts := make(map[string][]models.SoldAccount)
	for _, account := range accounts {
		productAccounts[account.ProductName] = append(productAccounts[account.ProductName], account)
	}

	// Build message with accounts (supports multiple formats)
	accountIndex := 1
	for productName, prodAccounts := range productAccounts {
		message.WriteString(fmt.Sprintf("📦 *%s*\n", productName))
		message.WriteString(fmt.Sprintf("   Jumlah: %d item\n\n", len(prodAccounts)))

		for _, account := range prodAccounts {
			contentLabel := account.GetContentLabel()
			contentData := account.FormatContent()
			
			message.WriteString(fmt.Sprintf("   %s #%d:\n", contentLabel, accountIndex))
			message.WriteString(fmt.Sprintf("   `%s`\n\n", contentData))
			accountIndex++
		}
	}

	message.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")
	message.WriteString("📋 *CARA MENGGUNAKAN:*\n")
	message.WriteString("1. Tap/klik pada data produk untuk menyalin\n")
	message.WriteString("2. 🔐 Akun: Login dengan email | password\n")
	message.WriteString("3. 🔗 Link: Klik atau salin link untuk redeem\n")
	message.WriteString("4. 🎫 Kode: Gunakan kode untuk aktivasi\n\n")
	message.WriteString("⚠️ *PENTING:*\n")
	message.WriteString("• Simpan data ini dengan aman\n")
	message.WriteString("• Jangan share ke orang lain\n")
	message.WriteString("• Segera gunakan sesuai petunjuk produk\n\n")
	message.WriteString("💬 Butuh bantuan? Hubungi /contact\n")
	message.WriteString("⭐️ Terima kasih telah berbelanja!")

	msg := tgbotapi.NewMessage(order.UserID, message.String())
	msg.ParseMode = tgbotapi.ModeMarkdown

	// Add keyboard with copy buttons for each account
	var keyboard [][]tgbotapi.InlineKeyboardButton
	
	accountButtonIndex := 1
	for productName, prodAccounts := range productAccounts {
		// Add product header button (non-clickable display)
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📦 %s", productName),
				fmt.Sprintf("product_header:%s", productName),
			),
		))

		// Add copy button for each account
		for _, account := range prodAccounts {
			contentLabel := account.GetContentLabel()
			contentData := account.FormatContent()
			
			keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("📋 Copy %s #%d", contentLabel, accountButtonIndex),
					fmt.Sprintf("copy_account:%d:%s", account.ID, orderID),
				),
			))
			accountButtonIndex++

			// Send individual copyable content message
			accountMsg := fmt.Sprintf("%s #%d - %s*\n\n`%s`\n\n_Tap untuk menyalin_", 
				contentLabel, accountButtonIndex-1, productName, contentData)
			copyMsg := tgbotapi.NewMessage(order.UserID, accountMsg)
			copyMsg.ParseMode = tgbotapi.ModeMarkdown
			b.api.Send(copyMsg)
		}
	}

	// Add help buttons
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📞 Hubungi Admin", "contact"),
		tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
	))

	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err := b.api.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send accounts message to user %d: %v", order.UserID, err)
		return err
	}

	logrus.Infof("Accounts delivered to user %d for order %s", order.UserID, order.ID)
	return nil
}

// sendAdminSaleNotification sends comprehensive sale notification to admin
func (b *Bot) sendAdminSaleNotification(order *models.Order, soldAccounts []models.SoldAccount, buyer *models.User, paidAmount int) {
	// Get all admin IDs
	adminIDs := b.config.AdminIDs

	var notification strings.Builder
	notification.WriteString("💰 *PEMBERITAHUAN PENJUALAN BARU!*\n\n")
	notification.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")
	
	// Order information
	notification.WriteString("📋 *INFORMASI PESANAN*\n")
	notification.WriteString(fmt.Sprintf("🆔 Order ID: `%s`\n", order.ID))
	notification.WriteString(fmt.Sprintf("📅 Waktu: %s\n", time.Now().Format("02/01/2006 15:04:05")))
	notification.WriteString(fmt.Sprintf("💰 Total: *%s*\n", models.FormatPrice(paidAmount, b.config.CurrencySymbol)))
	notification.WriteString(fmt.Sprintf("✅ Status: LUNAS\n\n"))

	// Buyer information
	notification.WriteString("👤 *DATA PEMBELI*\n")
	if buyer != nil {
		buyerName := "Unknown"
		if buyer.FirstName != nil && *buyer.FirstName != "" {
			buyerName = *buyer.FirstName
			if buyer.LastName != nil && *buyer.LastName != "" {
				buyerName += " " + *buyer.LastName
			}
		}
		notification.WriteString(fmt.Sprintf("📛 Nama: %s\n", buyerName))
		
		if buyer.Username != nil && *buyer.Username != "" {
			notification.WriteString(fmt.Sprintf("👤 Username: @%s\n", *buyer.Username))
		}
		notification.WriteString(fmt.Sprintf("🆔 User ID: `%d`\n", buyer.UserID))
	} else {
		notification.WriteString(fmt.Sprintf("🆔 User ID: `%d`\n", order.UserID))
	}
	notification.WriteString("\n")

	// Products and accounts sold
	notification.WriteString("📦 *DETAIL PEMBELIAN*\n\n")
	
	// Group accounts by product
	productAccountsMap := make(map[string][]models.SoldAccount)
	productPriceMap := make(map[string]int)
	
	for _, account := range soldAccounts {
		productAccountsMap[account.ProductName] = append(productAccountsMap[account.ProductName], account)
		productPriceMap[account.ProductName] = account.SoldPrice
	}

	itemNo := 1
	for productName, accounts := range productAccountsMap {
		quantity := len(accounts)
		price := productPriceMap[productName]
		subtotal := price * quantity

		notification.WriteString(fmt.Sprintf("%d. *%s*\n", itemNo, productName))
		notification.WriteString(fmt.Sprintf("   • Jumlah: %d akun\n", quantity))
		notification.WriteString(fmt.Sprintf("   • Harga satuan: %s\n", models.FormatPrice(price, b.config.CurrencySymbol)))
		notification.WriteString(fmt.Sprintf("   • Subtotal: %s\n\n", models.FormatPrice(subtotal, b.config.CurrencySymbol)))
		itemNo++
	}

	notification.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Sold accounts details (supports multiple formats)
	notification.WriteString("📦 *PRODUK YANG TERJUAL*\n\n")
	accountNo := 1
	for productName, accounts := range productAccountsMap {
		notification.WriteString(fmt.Sprintf("📦 *%s* (%d item):\n", productName, len(accounts)))
		for _, account := range accounts {
			contentLabel := account.GetContentLabel()
			contentData := account.FormatContent()
			notification.WriteString(fmt.Sprintf("   %d. %s: `%s`\n", accountNo, contentLabel, contentData))
			accountNo++
		}
		notification.WriteString("\n")
	}

	notification.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Stock status
	notification.WriteString("📊 *STATUS STOK TERKINI*\n")
	for productName, accounts := range productAccountsMap {
		// Get product ID from first account
		productID := accounts[0].ProductID
		
		// Get stock summary
		stockSummary, err := b.db.GetProductStockSummary(productID)
		if err == nil {
			notification.WriteString(fmt.Sprintf("• %s:\n", productName))
			notification.WriteString(fmt.Sprintf("  ✅ Tersedia: %d akun\n", stockSummary.AvailableStock))
			notification.WriteString(fmt.Sprintf("  💰 Terjual: %d akun\n", stockSummary.SoldStock))
			notification.WriteString(fmt.Sprintf("  📊 Total: %d akun\n\n", stockSummary.TotalStock))
		}
	}

	notification.WriteString("━━━━━━━━━━━━━━━━━━━━━\n\n")
	notification.WriteString("✨ *Transaksi berhasil diproses!*")

	// Send to all admins
	for _, adminID := range adminIDs {
		msg := tgbotapi.NewMessage(adminID, notification.String())
		msg.ParseMode = tgbotapi.ModeMarkdown

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📊 Lihat Stok", "admin:stock"),
				tgbotapi.NewInlineKeyboardButtonData("💰 Kelola Pesanan", "admin:orders"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Panel Admin", "admin:main"),
			),
		)
		msg.ReplyMarkup = keyboard

		_, err := b.api.Send(msg)
		if err != nil {
			logrus.Errorf("Failed to send sale notification to admin %d: %v", adminID, err)
		}
	}

	logrus.Infof("Sale notification sent to admins for order %s", order.ID)
}

// notifyAdminManipulationAttempt notifies admin about payment manipulation attempt
func (b *Bot) notifyAdminManipulationAttempt(orderID string, expectedAmount, paidAmount int, userID int64) {
	adminIDs := b.config.AdminIDs

	notification := fmt.Sprintf(`🚨 *DETEKSI MANIPULASI PEMBAYARAN!*

⚠️ Terdeteksi upaya manipulasi nominal pembayaran:

🆔 Order ID: %s
👤 User ID: %d
💰 Nominal Expected: %s
❌ Nominal Diterima: %s
📊 Selisih: %s

⏰ Waktu: %s

🔒 *Tindakan yang diambil:*
• Pembayaran ditolak
• Order tetap pending
• Perlu investigasi lebih lanjut

💡 Silakan cek detail order dan user untuk tindakan lebih lanjut.`,
		orderID,
		userID,
		models.FormatPrice(expectedAmount, b.config.CurrencySymbol),
		models.FormatPrice(paidAmount, b.config.CurrencySymbol),
		models.FormatPrice(abs(expectedAmount-paidAmount), b.config.CurrencySymbol),
		time.Now().Format("02/01/2006 15:04:05"))

	for _, adminID := range adminIDs {
		msg := tgbotapi.NewMessage(adminID, notification)
		msg.ParseMode = tgbotapi.ModeMarkdown

		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🔍 Investigasi", fmt.Sprintf("investigate:%s", orderID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Panel Admin", "admin:main"),
			),
		)
		msg.ReplyMarkup = keyboard

		b.api.Send(msg)
	}

	logrus.Warnf("⚠️ Manipulation attempt detected for order %s: expected %d, got %d", 
		orderID, expectedAmount, paidAmount)
}

// handleSimulatePayment simulates payment success for testing (admin only)
func (b *Bot) handleSimulatePayment(callback *tgbotapi.CallbackQuery, orderID string) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Get order
	order, err := b.db.GetOrder(orderID)
	if err != nil || order == nil {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Order tidak ditemukan"))
		return
	}

	// Simulate payment success
	if err := b.handlePaymentSuccess(orderID, order.TotalAmount); err != nil {
		logrus.Errorf("Failed to simulate payment for order %s: %v", orderID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal mensimulasi pembayaran"))
		return
	}

	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Pembayaran disimulasi berhasil!"))

	text := fmt.Sprintf("✅ *PEMBAYARAN DISIMULASI*\n\nOrder #%s berhasil diproses.\nAkun telah dikirim ke pembeli.", orderID[:8])
	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(edit)
}

// abs returns absolute value of integer
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
