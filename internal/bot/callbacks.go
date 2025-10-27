package bot

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"telegram-premium-store/internal/models"
	"telegram-premium-store/internal/payment"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// handleCallbackQuery processes callback queries from inline keyboards
func (b *Bot) handleCallbackQuery(callback *tgbotapi.CallbackQuery) {
	defer func() {
		// Always answer callback query to remove loading state
		b.api.Request(tgbotapi.NewCallback(callback.ID, ""))
	}()

	data := callback.Data
	parts := strings.Split(data, ":")

	switch parts[0] {
	case "start":
		b.handleStartCallback(callback)
	case "help":
		b.handleHelpCallback(callback)
	case "catalog":
		page := 0
		if len(parts) > 1 {
			if p, err := strconv.Atoi(parts[1]); err == nil {
				page = p
			}
		}
		b.handleCatalogCallback(callback, "", page)
	case "category":
		category := ""
		page := 0
		if len(parts) > 1 {
			category = parts[1]
		}
		if len(parts) > 2 {
			if p, err := strconv.Atoi(parts[2]); err == nil {
				page = p
			}
		}
		b.handleCatalogCallback(callback, category, page)
	case "product":
		if len(parts) > 1 {
			if productID, err := strconv.Atoi(parts[1]); err == nil {
				b.handleProductDetail(callback, productID)
			}
		}
	case "buy":
		if len(parts) > 1 {
			productID, _ := strconv.Atoi(parts[1])
			quantity := 1
			if len(parts) > 2 {
				quantity, _ = strconv.Atoi(parts[2])
			}
			b.handleAddToCart(callback, productID, quantity)
		}
	case "addcart":
		if len(parts) > 1 {
			productID, _ := strconv.Atoi(parts[1])
			quantity := 1
			if len(parts) > 2 {
				quantity, _ = strconv.Atoi(parts[2])
			}
			b.handleAddToCart(callback, productID, quantity)
		}
	case "notify":
		if len(parts) > 1 {
			if productID, err := strconv.Atoi(parts[1]); err == nil {
				b.handleStockNotification(callback, productID)
			}
		}
	case "cart":
		b.handleCartCallback(callback)
	case "clearcart":
		b.handleClearCart(callback)
	case "checkout":
		b.handleCheckout(callback)
	case "order":
		if len(parts) > 1 {
			b.handleOrderDetail(callback, parts[1])
		}
	case "contact":
		b.handleContactCallback(callback)
	case "cancel":
		if len(parts) > 1 {
			b.handleCancelOrder(callback, parts[1])
		}
	case "confirm_cancel":
		if len(parts) > 1 {
			b.handleConfirmCancel(callback, parts[1])
		}
	case "admin":
		if len(parts) > 1 {
			b.handleAdminCallback(callback, parts[1])
		}
	case "qris":
		if len(parts) > 1 {
			b.handleQRISCallback(callback, parts[1])
		}
	case "simulate_payment":
		if len(parts) > 1 {
			b.handleSimulatePayment(callback, parts[1])
		}
	case "copy_account":
		// Handle copy account callback
		if len(parts) >= 3 {
			accountID := parts[1]
			orderID := parts[2]
			b.handleCopyAccount(callback, accountID, orderID)
		}
	case "product_header":
		// Just acknowledge, this is a display button
		b.api.Request(tgbotapi.NewCallback(callback.ID, ""))
	case "investigate":
		// Handle investigation request
		if len(parts) > 1 {
			b.handleInvestigateOrder(callback, parts[1])
		}
	default:
		logrus.Warnf("Unknown callback data: %s", data)
	}
}

// handleStartCallback handles start button callback
func (b *Bot) handleStartCallback(callback *tgbotapi.CallbackQuery) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Lihat Katalog", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 Keranjang", "cart"),
			tgbotapi.NewInlineKeyboardButtonData("📞 Kontak", "contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("ℹ️ Bantuan", "help"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, b.messages.Welcome)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleHelpCallback handles help button callback
func (b *Bot) handleHelpCallback(callback *tgbotapi.CallbackQuery) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, b.messages.Help)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleCatalogCallback handles catalog display via callback
func (b *Bot) handleCatalogCallback(callback *tgbotapi.CallbackQuery, category string, page int) {
	const itemsPerPage = 5

	// Get categories for filter buttons
	categories, err := b.db.GetCategories()
	if err != nil {
		logrus.Errorf("Failed to get categories: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat kategori"))
		return
	}

	// Get products
	products, err := b.db.GetProducts(category, itemsPerPage, page*itemsPerPage)
	if err != nil {
		logrus.Errorf("Failed to get products: %v", err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat produk"))
		return
	}

	if len(products) == 0 {
		text := "🚫 Tidak ada produk yang tersedia"
		if category != "" {
			text += " untuk kategori ini"
		}
		text += "."
		
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
			),
		)
		
		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
		edit.ReplyMarkup = &keyboard
		b.api.Send(edit)
		return
	}

	// Build message text
	var text strings.Builder
	text.WriteString("📱 *KATALOG APLIKASI PREMIUM*\n\n")

	if category != "" {
		for _, cat := range categories {
			if cat.Name == category {
				text.WriteString(fmt.Sprintf("📂 Kategori: %s\n\n", cat.DisplayName))
				break
			}
		}
	}

	// Category filter buttons (only show if no category selected)
	var keyboard [][]tgbotapi.InlineKeyboardButton
	if category == "" {
		text.WriteString("🏷️ *Filter berdasarkan kategori:*\n\n")
		
		var catRow []tgbotapi.InlineKeyboardButton
		for _, cat := range categories {
			if cat.Count > 0 {
				catRow = append(catRow, tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s (%d)", cat.Icon, cat.Count),
					fmt.Sprintf("category:%s:0", cat.Name),
				))
				
				if len(catRow) == 2 {
					keyboard = append(keyboard, catRow)
					catRow = nil
				}
			}
		}
		if len(catRow) > 0 {
			keyboard = append(keyboard, catRow)
		}
		
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Semua Produk", "catalog:0"),
		))
	}

	// Product list
	for _, product := range products {
		text.WriteString(fmt.Sprintf("🔸 *%s*\n", product.Name))
		text.WriteString(fmt.Sprintf("💰 %s\n", models.FormatPrice(product.Price, b.config.CurrencySymbol)))
		
		desc := product.Description
		if len(desc) > 80 {
			desc = desc[:80] + "..."
		}
		text.WriteString(fmt.Sprintf("📝 %s\n\n", desc))

		// Product buttons
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👁️ Detail", fmt.Sprintf("product:%d", product.ID)),
			tgbotapi.NewInlineKeyboardButtonData("🛒 Beli", fmt.Sprintf("buy:%d", product.ID)),
		))
	}

	// Navigation buttons
	var navRow []tgbotapi.InlineKeyboardButton
	
	// Previous page
	if page > 0 {
		prevCallback := fmt.Sprintf("catalog:%d", page-1)
		if category != "" {
			prevCallback = fmt.Sprintf("category:%s:%d", category, page-1)
		}
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("⬅️ Sebelumnya", prevCallback))
	}

	// Next page (check if there are more products)
	if len(products) == itemsPerPage {
		nextCallback := fmt.Sprintf("catalog:%d", page+1)
		if category != "" {
			nextCallback = fmt.Sprintf("category:%s:%d", category, page+1)
		}
		navRow = append(navRow, tgbotapi.NewInlineKeyboardButtonData("Selanjutnya ➡️", nextCallback))
	}

	if len(navRow) > 0 {
		keyboard = append(keyboard, navRow)
	}

	// Back buttons
	if category != "" {
		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Semua Kategori", "catalog:0"),
		))
	}
	
	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
	))

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: keyboard}

	b.api.Send(edit)
}

// handleProductDetail shows detailed product information
func (b *Bot) handleProductDetail(callback *tgbotapi.CallbackQuery, productID int) {
	product, err := b.db.GetProduct(productID)
	if err != nil {
		logrus.Errorf("Failed to get product %d: %v", productID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat produk"))
		return
	}

	if product == nil {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Produk tidak ditemukan"))
		return
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("📱 *%s*\n\n", product.Name))
	text.WriteString(fmt.Sprintf("📝 *Deskripsi:*\n%s\n\n", product.Description))
	text.WriteString(fmt.Sprintf("💰 *Harga:* %s\n", models.FormatPrice(product.Price, b.config.CurrencySymbol)))
	
	// Find category display name
	categories, _ := b.db.GetCategories()
	for _, cat := range categories {
		if cat.Name == product.Category {
			text.WriteString(fmt.Sprintf("🏷️ *Kategori:* %s\n", cat.DisplayName))
			break
		}
	}
	
	text.WriteString(fmt.Sprintf("📦 *Stok:* %d tersedia\n\n", product.Stock))
	text.WriteString("✅ *Status:* Tersedia")

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🛒 Tambah ke Keranjang", fmt.Sprintf("addcart:%d", product.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Beli Sekarang", fmt.Sprintf("buy:%d", product.ID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Kembali ke Katalog", "catalog:0"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleAddToCart adds product to user's cart
func (b *Bot) handleAddToCart(callback *tgbotapi.CallbackQuery, productID, quantity int) {
	userID := callback.From.ID

	// Check if product exists and is available
	product, err := b.db.GetProduct(productID)
	if err != nil {
		logrus.Errorf("Failed to get product %d: %v", productID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat produk"))
		return
	}

	if product == nil {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Produk tidak ditemukan"))
		return
	}

	if product.Stock < quantity {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Stok tidak mencukupi"))
		return
	}

	// Add to cart
	err = b.db.AddToCart(userID, productID, quantity)
	if err != nil {
		logrus.Errorf("Failed to add product %d to cart for user %d: %v", productID, userID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal menambahkan ke keranjang"))
		return
	}

	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Produk ditambahkan ke keranjang!"))
	
	// Show cart after adding
	b.handleCartCallback(callback)
}

// handleCartCallback shows user's shopping cart
func (b *Bot) handleCartCallback(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	cartItems, err := b.db.GetCart(userID)
	if err != nil {
		logrus.Errorf("Failed to get cart for user %d: %v", userID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat keranjang"))
		return
	}

	if len(cartItems) == 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📱 Lihat Katalog", "catalog:0"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
			),
		)

		edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, b.messages.CartEmpty)
		edit.ParseMode = tgbotapi.ModeMarkdown
		edit.ReplyMarkup = &keyboard
		b.api.Send(edit)
		return
	}

	// Build cart display
	var text strings.Builder
	text.WriteString("🛒 *KERANJANG BELANJA*\n\n")

	totalPrice := 0
	for _, item := range cartItems {
		subtotal := item.ProductPrice * item.Quantity
		totalPrice += subtotal

		text.WriteString(fmt.Sprintf("🔸 *%s*\n", item.ProductName))
		text.WriteString(fmt.Sprintf("   Jumlah: %d x %s = %s\n\n",
			item.Quantity,
			models.FormatPrice(item.ProductPrice, b.config.CurrencySymbol),
			models.FormatPrice(subtotal, b.config.CurrencySymbol)))
	}

	text.WriteString(fmt.Sprintf("💰 *Total: %s*\n", models.FormatPrice(totalPrice, b.config.CurrencySymbol)))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💳 Checkout", "checkout"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Kosongkan Keranjang", "clearcart"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Lanjut Belanja", "catalog:0"),
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleClearCart clears user's shopping cart
func (b *Bot) handleClearCart(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	
	err := b.db.ClearCart(userID)
	if err != nil {
		logrus.Errorf("Failed to clear cart for user %d: %v", userID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal mengosongkan keranjang"))
		return
	}

	b.api.Request(tgbotapi.NewCallback(callback.ID, "🗑️ Keranjang dikosongkan!"))
	b.handleCartCallback(callback)
}

// handleCheckout processes checkout and creates order with QRIS payment
func (b *Bot) handleCheckout(callback *tgbotapi.CallbackQuery) {
	userID := callback.From.ID
	
	// Get cart items
	cartItems, err := b.db.GetCart(userID)
	if err != nil {
		logrus.Errorf("Failed to get cart for user %d: %v", userID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat keranjang"))
		return
	}

	if len(cartItems) == 0 {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Keranjang kosong"))
		return
	}

	// Check if real QRIS is configured
	if !b.realQRISService.IsConfigured() {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Sistem pembayaran belum dikonfigurasi"))
		return
	}

	// Validate account availability for all items before checkout
	for _, item := range cartItems {
		product, err := b.db.GetProduct(item.ProductID)
		if err != nil || product == nil {
			b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Produk tidak ditemukan"))
			return
		}

		// Check available accounts instead of stock
		availableAccounts, err := b.db.GetAvailableAccountCount(item.ProductID)
		if err != nil {
			logrus.Errorf("Failed to get available accounts for product %d: %v", item.ProductID, err)
			b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal validasi stok"))
			return
		}

		if availableAccounts < item.Quantity {
			b.api.Request(tgbotapi.NewCallback(callback.ID, 
				fmt.Sprintf("❌ Stok %s tidak mencukupi! Tersedia: %d akun, diminta: %d", 
					product.Name, availableAccounts, item.Quantity)))
			return
		}

		if availableAccounts == 0 {
			b.api.Request(tgbotapi.NewCallback(callback.ID, 
				fmt.Sprintf("❌ %s sedang tidak tersedia (stok habis)", product.Name)))
			return
		}
	}

	// Calculate total
	totalAmount := 0
	var orderItems []models.OrderItem
	for _, item := range cartItems {
		subtotal := item.ProductPrice * item.Quantity
		totalAmount += subtotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     item.ProductPrice,
		})
	}

	// Generate order ID using real QRIS service
	orderID := b.realQRISService.GenerateOrderID()

	// Generate dynamic QRIS payment
	qrisPayment, qrImage, err := b.realQRISService.GenerateDynamicQRIS(orderID, totalAmount)
	if err != nil {
		logrus.Errorf("Failed to generate QRIS for order %s: %v", orderID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal membuat pembayaran"))
		return
	}

	// Create order
	order := &models.Order{
		ID:            orderID,
		UserID:        userID,
		TotalAmount:   totalAmount,
		PaymentMethod: "qris",
		PaymentStatus: models.PaymentStatusPending,
		QRISCode:      &qrisPayment.QRString,
		QRISExpiry:    &qrisPayment.ExpiryTime,
		Items:         orderItems,
	}

	// Create order with account assignment
	_, err = b.db.CreateOrderWithAccounts(order)
	if err != nil {
		logrus.Errorf("Failed to create order %s: %v", orderID, err)
		if strings.Contains(err.Error(), "insufficient accounts") {
			b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Stok akun tidak mencukupi"))
		} else {
			b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal membuat pesanan"))
		}
		return
	}

	// Create payment verification record
	verifier := payment.NewPaymentVerifier(b.config.BotToken) // Use bot token as secret key
	verificationHash := verifier.GenerateVerificationHash(orderID, totalAmount, qrisPayment.QRString)
	
	err = b.db.CreatePaymentVerification(orderID, totalAmount, qrisPayment.QRString, verificationHash)
	if err != nil {
		logrus.Errorf("Failed to create payment verification for order %s: %v", orderID, err)
	}

	// Clear cart after successful order creation
	b.db.ClearCart(userID)

	// Send order success message
	orderText := fmt.Sprintf(b.messages.OrderSuccess,
		orderID,
		b.config.CurrencySymbol,
		models.FormatPrice(totalAmount, b.config.CurrencySymbol),
		time.Now().Format("02/01/2006 15:04"))

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, orderText)
	edit.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(edit)

	// Send QRIS QR Code
	qrMsg := tgbotapi.NewPhoto(callback.Message.Chat.ID, tgbotapi.FileBytes{
		Name:  fmt.Sprintf("qris_%s.png", orderID),
		Bytes: qrImage,
	})

	// Payment instructions using real QRIS service
	instructions := b.realQRISService.GetPaymentInstructions(orderID, totalAmount)
	qrMsg.Caption = instructions
	qrMsg.ParseMode = tgbotapi.ModeMarkdown

	// Add supported banks info
	supportedBanks := b.realQRISService.GetSupportedBanks()
	qrMsg.Caption += "\n\n📱 *Aplikasi yang mendukung QRIS:*\n"
	for i, bank := range supportedBanks {
		qrMsg.Caption += fmt.Sprintf("• %s\n", bank)
		if i >= 9 { // Show first 10 only to avoid message length limit
			qrMsg.Caption += fmt.Sprintf("• ... dan %d lainnya\n", len(supportedBanks)-10)
			break
		}
	}

	// Add keyboard for order management
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Detail Pesanan", fmt.Sprintf("order:%s", orderID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("❌ Batalkan Pesanan", fmt.Sprintf("cancel:%s", orderID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📞 Hubungi Admin", "contact"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)
	qrMsg.ReplyMarkup = keyboard

	b.api.Send(qrMsg)

	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Pesanan berhasil dibuat!"))
}

// handleOrderDetail shows order details
func (b *Bot) handleOrderDetail(callback *tgbotapi.CallbackQuery, orderID string) {
	order, err := b.db.GetOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to get order %s: %v", orderID, err)
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Gagal memuat pesanan"))
		return
	}

	if order == nil || order.UserID != callback.From.ID {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Pesanan tidak ditemukan"))
		return
	}

	var text strings.Builder
	text.WriteString("📄 *DETAIL PESANAN*\n\n")
	text.WriteString(fmt.Sprintf("🆔 Order ID: #%s\n", order.ID))
	text.WriteString(fmt.Sprintf("📅 Tanggal: %s\n", order.CreatedAt.Format("02/01/2006 15:04")))
	text.WriteString(fmt.Sprintf("💰 Total: %s\n", models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol)))
	text.WriteString(fmt.Sprintf("💳 Metode: QRIS\n"))
	text.WriteString(fmt.Sprintf("📊 Status: %s %s\n\n", b.getStatusEmoji(order.PaymentStatus), cases.Title(language.Und).String(string(order.PaymentStatus))))

	if order.QRISExpiry != nil {
		if b.paymentService.IsExpired(order.QRISExpiry) {
			text.WriteString("⏰ QR Code: Expired\n\n")
		} else {
			text.WriteString(fmt.Sprintf("⏰ QR Code berlaku sampai: %s\n\n", order.QRISExpiry.Format("15:04")))
		}
	}

	text.WriteString("📦 *Item Pesanan:*\n")
	for _, item := range order.Items {
		subtotal := item.Price * item.Quantity
		text.WriteString(fmt.Sprintf("• %s\n", item.ProductName))
		text.WriteString(fmt.Sprintf("  %d x %s = %s\n",
			item.Quantity,
			models.FormatPrice(item.Price, b.config.CurrencySymbol),
			models.FormatPrice(subtotal, b.config.CurrencySymbol)))
	}

	var keyboardRows [][]tgbotapi.InlineKeyboardButton
	
	// Add simulate payment button for admins if order is pending
	if b.config.IsAdmin(callback.From.ID) && order.PaymentStatus == models.PaymentStatusPending {
		keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🧪 [Admin] Simulasi Pembayaran", fmt.Sprintf("simulate_payment:%s", orderID)),
		))
	}
	
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("📞 Hubungi Admin", "contact"),
	))
	keyboardRows = append(keyboardRows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
	))
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(keyboardRows...)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleContactCallback handles contact button callback
func (b *Bot) handleContactCallback(callback *tgbotapi.CallbackQuery) {
	contactText := fmt.Sprintf(b.messages.Contact,
		b.config.AdminUsername,
		b.config.AdminEmail,
		b.config.SupportPhone,
		b.config.BusinessHours)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, contactText)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleAdminCallback handles admin panel callbacks
func (b *Bot) handleAdminCallback(callback *tgbotapi.CallbackQuery, action string) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	parts := strings.Split(action, ":")
	mainAction := parts[0]

	switch mainAction {
	case "main":
		b.handleAdminMain(callback)
	case "stats":
		b.handleAdminStats(callback)
	case "users":
		b.handleAdminUsers(callback)
	case "products":
		b.handleProductManagement(callback)
	case "orders":
		b.handleAdminOrders(callback)
	case "stock":
		b.handleStockManagement(callback)
	case "lowstock":
		b.handleLowStock(callback)
	case "categories":
		b.handleCategoryManagement(callback)
	case "broadcast":
		if len(parts) > 1 {
			if parts[1] == "all" || parts[1] == "active" {
				b.handleSendBroadcast(callback, parts[1])
			} else {
				b.handleBroadcastManagement(callback)
			}
		} else {
			b.handleBroadcastManagement(callback)
		}
	case "confirm_broadcast":
		if len(parts) > 1 {
			if broadcastID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				b.executeBroadcast(callback, broadcastID)
			}
		}
	default:
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Aksi tidak dikenal"))
	}
}

// Admin callback handlers (simplified for demo)
func (b *Bot) handleAdminStats(callback *tgbotapi.CallbackQuery) {
	text := "📊 *STATISTIK BOT*\n\nFitur ini akan dikembangkan lebih lanjut."
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

func (b *Bot) handleAdminUsers(callback *tgbotapi.CallbackQuery) {
	text := "👥 *KELOLA PENGGUNA*\n\nFitur ini akan dikembangkan lebih lanjut."
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

func (b *Bot) handleAdminProducts(callback *tgbotapi.CallbackQuery) {
	text := "📦 *KELOLA PRODUK*\n\nFitur ini akan dikembangkan lebih lanjut."
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

func (b *Bot) handleAdminMain(callback *tgbotapi.CallbackQuery) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Statistik", "admin:stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📦 Kelola Produk", "admin:products"),
			tgbotapi.NewInlineKeyboardButtonData("🏷️ Kelola Kategori", "admin:categories"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Kelola Stok", "admin:stock"),
			tgbotapi.NewInlineKeyboardButtonData("💰 Kelola Pesanan", "admin:orders"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📢 Broadcast", "admin:broadcast"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔧 Setup QRIS", "qris:setup"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, 
		"👨‍💼 *PANEL ADMIN*\n\nPilih menu admin:")
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

func (b *Bot) handleAdminOrders(callback *tgbotapi.CallbackQuery) {
	text := "💰 *KELOLA PESANAN*\n\nFitur ini akan dikembangkan lebih lanjut."
	
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text)
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
}

// handleCopyAccount handles copy account callback (displays copyable text)
func (b *Bot) handleCopyAccount(callback *tgbotapi.CallbackQuery, accountID, orderID string) {
	// This callback is mainly for UI interaction feedback
	b.api.Request(tgbotapi.NewCallback(callback.ID, "✅ Tap pada kredensial untuk menyalin"))
}

// handleInvestigateOrder handles investigation request from admin
func (b *Bot) handleInvestigateOrder(callback *tgbotapi.CallbackQuery, orderID string) {
	if !b.config.IsAdmin(callback.From.ID) {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Akses ditolak"))
		return
	}

	// Get order details
	order, err := b.db.GetOrder(orderID)
	if err != nil || order == nil {
		b.api.Request(tgbotapi.NewCallback(callback.ID, "❌ Order tidak ditemukan"))
		return
	}

	// Get payment verification
	verification, err := b.db.GetPaymentVerification(orderID)
	
	var text strings.Builder
	text.WriteString("🔍 *INVESTIGASI ORDER*\n\n")
	text.WriteString(fmt.Sprintf("🆔 Order ID: `%s`\n", orderID))
	text.WriteString(fmt.Sprintf("👤 User ID: `%d`\n", order.UserID))
	text.WriteString(fmt.Sprintf("💰 Total: %s\n", models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol)))
	text.WriteString(fmt.Sprintf("📊 Status: %s\n", order.PaymentStatus))
	text.WriteString(fmt.Sprintf("📅 Dibuat: %s\n\n", order.CreatedAt.Format("02/01/2006 15:04")))

	if verification != nil {
		text.WriteString("🔐 *Verification Data:*\n")
		text.WriteString(fmt.Sprintf("• Expected Amount: %s\n", models.FormatPrice(verification.ExpectedAmount, b.config.CurrencySymbol)))
		text.WriteString(fmt.Sprintf("• Hash: `%s`\n", verification.VerificationHash[:16]))
		if verification.VerifiedAt != nil {
			text.WriteString(fmt.Sprintf("• Verified: %s\n", verification.VerifiedAt.Format("02/01/2006 15:04")))
		} else {
			text.WriteString("• Verified: ❌ Not yet\n")
		}
	} else {
		text.WriteString("⚠️ No verification data found\n")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📄 Detail Order", fmt.Sprintf("order:%s", orderID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔙 Panel Admin", "admin:main"),
		),
	)

	edit := tgbotapi.NewEditMessageText(callback.Message.Chat.ID, callback.Message.MessageID, text.String())
	edit.ParseMode = tgbotapi.ModeMarkdown
	edit.ReplyMarkup = &keyboard

	b.api.Send(edit)
	b.api.Request(tgbotapi.NewCallback(callback.ID, ""))
}