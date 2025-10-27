package bot

import (
	"fmt"
	"strings"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/database"
	"telegram-premium-store/internal/models"
	"telegram-premium-store/internal/payment"
	"telegram-premium-store/internal/qris"
	"telegram-premium-store/internal/scheduler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Bot represents the Telegram bot
type Bot struct {
	api              *tgbotapi.BotAPI
	config           *config.Config
	db               *database.DB
	paymentService   *payment.QRISService
	realQRISService  *qris.RealQRISService
	scheduler        *scheduler.Scheduler
	messages         *config.Messages
	updates          tgbotapi.UpdatesChannel
}

// New creates a new bot instance
func New(cfg *config.Config, db *database.DB, paymentService *payment.QRISService) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot API: %w", err)
	}

	api.Debug = cfg.LogLevel == "DEBUG"

	bot := &Bot{
		api:              api,
		config:           cfg,
		db:               db,
		paymentService:   paymentService,
		realQRISService:  qris.NewRealQRISService(cfg),
		messages:         config.GetMessages(),
	}

	// Initialize scheduler
	bot.scheduler = scheduler.NewScheduler(db, api, cfg)

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start() error {
	logrus.Info("Starting bot polling...")

	// Start background scheduler
	b.scheduler.Start()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)
	b.updates = updates

	for update := range updates {
		go b.handleUpdate(update)
	}

	return nil
}

// Stop stops the bot
func (b *Bot) Stop() {
	logrus.Info("Stopping bot...")
	
	// Stop scheduler
	if b.scheduler != nil {
		b.scheduler.Stop()
	}
	
	b.api.StopReceivingUpdates()
}

// handleUpdate processes incoming updates
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Panic in handleUpdate: %v", r)
		}
	}()

	// Handle callback queries
	if update.CallbackQuery != nil {
		b.handleCallbackQuery(update.CallbackQuery)
		return
	}

	// Handle messages
	if update.Message != nil {
		// Register user if not exists
		b.registerUser(update.Message.From)

		// Log user interaction for broadcast targeting
		b.db.LogUserInteraction(update.Message.From.ID, "message", "")

		if update.Message.IsCommand() {
			b.handleCommand(update.Message)
		} else {
			// Check if it's a QRIS image upload
			if update.Message.Photo != nil && b.isUserInState(update.Message.From.ID, "waiting_qris_upload") {
				b.handleQRISImageUpload(update.Message)
			} else if strings.HasPrefix(b.getUserState(update.Message.From.ID), "waiting_broadcast_") {
				// Handle broadcast message input
				targetType := strings.TrimPrefix(b.getUserState(update.Message.From.ID), "waiting_broadcast_")
				b.clearUserState(update.Message.From.ID)
				b.processBroadcastMessage(update.Message, targetType)
			} else {
				b.handleMessage(update.Message)
			}
		}
	}
}

// registerUser registers a new user or updates existing user info
func (b *Bot) registerUser(user *tgbotapi.User) {
	dbUser := &models.User{
		UserID:    user.ID,
		Username:  &user.UserName,
		FirstName: &user.FirstName,
		LastName:  &user.LastName,
		IsAdmin:   b.config.IsAdmin(user.ID),
	}

	if err := b.db.CreateUser(dbUser); err != nil {
		logrus.Errorf("Failed to register user %d: %v", user.ID, err)
	}
}

// handleCommand processes bot commands
func (b *Bot) handleCommand(message *tgbotapi.Message) {
	switch message.Command() {
	case "start":
		b.handleStart(message)
	case "help":
		b.handleHelp(message)
	case "catalog":
		b.handleCatalog(message, "", 0)
	case "cart":
		b.handleCart(message)
	case "history":
		b.handleHistory(message)
	case "payment":
		b.handlePaymentStatus(message)
	case "contact":
		b.handleContact(message)
	case "admin":
		b.handleAdmin(message)
	case "addproduct":
		b.handleAddProduct(message)
	case "users":
		b.handleUsers(message)
	case "orders":
		b.handleOrders(message)
	case "stats":
		b.handleStats(message)
	case "qrissetup":
		b.handleQRISSetup(message)
	case "qristest":
		b.handleQRISTestCommand(message)
	case "addstock":
		// Admin command to add product stock (supports all formats)
		b.processAddStockCommand(message)
	default:
		b.sendMessage(message.Chat.ID, "❌ Perintah tidak dikenal. Ketik /help untuk bantuan.")
	}
}

// handleStart handles /start command
func (b *Bot) handleStart(message *tgbotapi.Message) {
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

	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Welcome)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleHelp handles /help command
func (b *Bot) handleHelp(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.Help)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

// handleCatalog handles /catalog command and catalog display
func (b *Bot) handleCatalog(message *tgbotapi.Message, category string, page int) {
	const itemsPerPage = 5

	// Get categories for filter buttons
	categories, err := b.db.GetCategories()
	if err != nil {
		logrus.Errorf("Failed to get categories: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal memuat kategori.")
		return
	}

	// Get products
	products, err := b.db.GetProducts(category, itemsPerPage, page*itemsPerPage)
	if err != nil {
		logrus.Errorf("Failed to get products: %v", err)
		b.sendMessage(message.Chat.ID, "❌ Gagal memuat produk.")
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
		
		msg := tgbotapi.NewMessage(message.Chat.ID, text)
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
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

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	b.api.Send(msg)
}

// handleCart handles /cart command and cart display
func (b *Bot) handleCart(message *tgbotapi.Message) {
	userID := message.From.ID
	cartItems, err := b.db.GetCart(userID)
	if err != nil {
		logrus.Errorf("Failed to get cart for user %d: %v", userID, err)
		b.sendMessage(message.Chat.ID, "❌ Gagal memuat keranjang.")
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

		msg := tgbotapi.NewMessage(message.Chat.ID, b.messages.CartEmpty)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
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

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handleHistory handles /history command
func (b *Bot) handleHistory(message *tgbotapi.Message) {
	userID := message.From.ID
	orders, err := b.db.GetUserOrders(userID, 10, 0)
	if err != nil {
		logrus.Errorf("Failed to get orders for user %d: %v", userID, err)
		b.sendMessage(message.Chat.ID, "❌ Gagal memuat riwayat pembelian.")
		return
	}

	if len(orders) == 0 {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("📱 Mulai Belanja", "catalog:0"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
			),
		)

		msg := tgbotapi.NewMessage(message.Chat.ID, "📋 *RIWAYAT PEMBELIAN*\n\nAnda belum memiliki riwayat pembelian.")
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString("📋 *RIWAYAT PEMBELIAN*\n\n")

	for _, order := range orders {
		statusEmoji := b.getStatusEmoji(order.PaymentStatus)
		
		text.WriteString(fmt.Sprintf("🔸 Order #%s\n", order.ID[:8]))
		text.WriteString(fmt.Sprintf("💰 Total: %s\n", models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol)))
		text.WriteString(fmt.Sprintf("📅 Tanggal: %s\n", order.CreatedAt.Format("02/01/2006 15:04")))
		text.WriteString(fmt.Sprintf("📊 Status: %s %s\n", statusEmoji, cases.Title(language.Und).String(string(order.PaymentStatus))))
		
		if len(order.Items) > 0 {
			text.WriteString(fmt.Sprintf("📦 Item: %s", order.Items[0].ProductName))
			if len(order.Items) > 1 {
				text.WriteString(fmt.Sprintf(" (+%d lainnya)", len(order.Items)-1))
			}
			text.WriteString("\n")
		}
		text.WriteString("\n")
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Belanja Lagi", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// handlePaymentStatus handles /payment command
func (b *Bot) handlePaymentStatus(message *tgbotapi.Message) {
	// Get user's pending orders
	userID := message.From.ID
	orders, err := b.db.GetUserOrders(userID, 5, 0)
	if err != nil {
		logrus.Errorf("Failed to get orders for user %d: %v", userID, err)
		b.sendMessage(message.Chat.ID, "❌ Gagal memuat status pembayaran.")
		return
	}

	// Filter pending orders
	var pendingOrders []models.Order
	for _, order := range orders {
		if order.PaymentStatus == models.PaymentStatusPending {
			pendingOrders = append(pendingOrders, order)
		}
	}

	if len(pendingOrders) == 0 {
		msg := tgbotapi.NewMessage(message.Chat.ID, "💳 *STATUS PEMBAYARAN*\n\nTidak ada pembayaran yang sedang menunggu.")
		msg.ParseMode = tgbotapi.ModeMarkdown
		b.api.Send(msg)
		return
	}

	var text strings.Builder
	text.WriteString("💳 *STATUS PEMBAYARAN*\n\n")
	text.WriteString("Pesanan yang menunggu pembayaran:\n\n")

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, order := range pendingOrders {
		text.WriteString(fmt.Sprintf("🔸 Order #%s\n", order.ID[:8]))
		text.WriteString(fmt.Sprintf("💰 Total: %s\n", models.FormatPrice(order.TotalAmount, b.config.CurrencySymbol)))
		text.WriteString(fmt.Sprintf("📅 Dibuat: %s\n", order.CreatedAt.Format("02/01/2006 15:04")))
		
		if order.QRISExpiry != nil {
			if b.paymentService.IsExpired(order.QRISExpiry) {
				text.WriteString("⏰ Status: Expired\n")
			} else {
				text.WriteString(fmt.Sprintf("⏰ Berlaku sampai: %s\n", order.QRISExpiry.Format("15:04")))
			}
		}
		text.WriteString("\n")

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📄 Detail #%s", order.ID[:8]),
				fmt.Sprintf("order:%s", order.ID),
			),
		))
	}

	keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
	))

	msg := tgbotapi.NewMessage(message.Chat.ID, text.String())
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	b.api.Send(msg)
}

// handleContact handles /contact command
func (b *Bot) handleContact(message *tgbotapi.Message) {
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

	msg := tgbotapi.NewMessage(message.Chat.ID, contactText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

// Admin command handlers
func (b *Bot) handleAdmin(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Statistik", "admin:stats"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👥 Pengguna", "admin:users"),
			tgbotapi.NewInlineKeyboardButtonData("📦 Produk", "admin:products"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("💰 Pesanan", "admin:orders"),
		),
	)

	msg := tgbotapi.NewMessage(message.Chat.ID, "👨‍💼 *PANEL ADMIN*\n\nPilih menu admin:")
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	b.api.Send(msg)
}

func (b *Bot) handleAddProduct(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	helpText := `📝 *TAMBAH PRODUK BARU*

Format: /addproduct <nama> | <deskripsi> | <harga> | <kategori>

*Contoh:*
/addproduct Discord Nitro | Premium Discord features | 50000 | gaming

*Kategori yang tersedia:*
• music - Musik & Audio
• entertainment - Hiburan  
• design - Design & Kreativitas
• productivity - Produktivitas
• education - Edukasi
• gaming - Gaming
• social - Sosial Media
• utility - Utilitas`

	b.sendMessage(message.Chat.ID, helpText)
}

func (b *Bot) handleUsers(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	b.sendMessage(message.Chat.ID, "👥 *STATISTIK PENGGUNA*\n\nFitur ini akan dikembangkan lebih lanjut.")
}

func (b *Bot) handleOrders(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	b.sendMessage(message.Chat.ID, "💰 *KELOLA PESANAN*\n\nFitur ini akan dikembangkan lebih lanjut.")
}

func (b *Bot) handleStats(message *tgbotapi.Message) {
	if !b.config.IsAdmin(message.From.ID) {
		b.sendMessage(message.Chat.ID, "❌ Anda tidak memiliki akses admin!")
		return
	}

	b.sendMessage(message.Chat.ID, "📊 *STATISTIK BOT*\n\nFitur ini akan dikembangkan lebih lanjut.")
}

// handleMessage processes non-command messages
func (b *Bot) handleMessage(message *tgbotapi.Message) {
	b.sendMessage(message.Chat.ID, "ℹ️ Gunakan menu atau ketik /help untuk melihat perintah yang tersedia.")
}

// Helper methods
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	b.api.Send(msg)
}

func (b *Bot) getStatusEmoji(status models.PaymentStatus) string {
	switch status {
	case models.PaymentStatusPaid:
		return "✅"
	case models.PaymentStatusPending:
		return "⏳"
	case models.PaymentStatusExpired:
		return "⏰"
	case models.PaymentStatusCancelled:
		return "❌"
	case models.PaymentStatusRefunded:
		return "🔄"
	default:
		return "❓"
	}
}