package scheduler

import (
	"fmt"
	"time"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/database"
	"telegram-premium-store/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

// Scheduler handles background tasks
type Scheduler struct {
	db     *database.DB
	api    *tgbotapi.BotAPI
	config *config.Config
	stopCh chan bool
}

// NewScheduler creates a new scheduler
func NewScheduler(db *database.DB, api *tgbotapi.BotAPI, cfg *config.Config) *Scheduler {
	return &Scheduler{
		db:     db,
		api:    api,
		config: cfg,
		stopCh: make(chan bool),
	}
}

// Start starts the background scheduler
func (s *Scheduler) Start() {
	logrus.Info("🕐 Starting background scheduler...")

	// Start expired orders checker (every minute)
	go s.expiredOrdersChecker()

	// Start daily stock alert (check every hour, send at 8 PM)
	go s.dailyStockAlert()

	// Start admin notification checker (every 30 seconds)
	go s.adminNotificationChecker()

	logrus.Info("✅ Background scheduler started")
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	logrus.Info("🛑 Stopping background scheduler...")
	close(s.stopCh)
}

// expiredOrdersChecker checks for expired orders every minute
func (s *Scheduler) expiredOrdersChecker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkExpiredOrders()
		case <-s.stopCh:
			return
		}
	}
}

// checkExpiredOrders finds and handles expired orders
func (s *Scheduler) checkExpiredOrders() {
	// Get all pending orders that are expired
	rows, err := s.db.Query(`
		SELECT id, user_id, total_amount, qris_expiry, created_at
		FROM orders 
		WHERE payment_status = 'pending' 
		AND qris_expiry IS NOT NULL 
		AND qris_expiry < datetime('now')
	`)
	if err != nil {
		logrus.Errorf("Failed to query expired orders: %v", err)
		return
	}
	defer rows.Close()

	expiredCount := 0
	for rows.Next() {
		var orderID string
		var userID int64
		var totalAmount int
		var qrisExpiry time.Time
		var createdAt time.Time

		err := rows.Scan(&orderID, &userID, &totalAmount, &qrisExpiry, &createdAt)
		if err != nil {
			logrus.Errorf("Failed to scan expired order: %v", err)
			continue
		}

		// Handle expired order
		s.handleExpiredOrder(orderID, userID, totalAmount)
		expiredCount++
	}

	if expiredCount > 0 {
		logrus.Infof("Processed %d expired orders", expiredCount)
	}
}

// handleExpiredOrder processes a single expired order
func (s *Scheduler) handleExpiredOrder(orderID string, userID int64, totalAmount int) {
	// Update order status to expired
	err := s.db.UpdateOrderStatus(orderID, models.PaymentStatusExpired)
	if err != nil {
		logrus.Errorf("Failed to expire order %s: %v", orderID, err)
		return
	}

	// Restore stock
	err = s.db.RestoreStockFromOrder(orderID)
	if err != nil {
		logrus.Errorf("Failed to restore stock for expired order %s: %v", orderID, err)
	}

	// Send expiry notification to user
	expiredText := fmt.Sprintf(`⏰ *WAKTU PEMBAYARAN HABIS*

Waktu pembayaran untuk pesanan #%s telah habis.

💰 Nominal: %s
📅 Expired: %s

💡 Silakan dapat melakukan pemesanan kembali jika masih membutuhkan produk tersebut.

Terima kasih atas pengertiannya.`, 
		orderID[:8],
		models.FormatPrice(totalAmount, s.config.CurrencySymbol),
		time.Now().Format("02/01/2006 15:04"))

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📱 Pesan Lagi", "catalog:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏠 Menu Utama", "start"),
		),
	)

	msg := tgbotapi.NewMessage(userID, expiredText)
	msg.ParseMode = tgbotapi.ModeMarkdown
	msg.ReplyMarkup = keyboard

	_, err = s.api.Send(msg)
	if err != nil {
		logrus.Errorf("Failed to send expiry notification to user %d: %v", userID, err)
	}

	logrus.Infof("Order %s expired and user %d notified", orderID, userID)
}

// dailyStockAlert sends daily stock alerts at 8 PM
func (s *Scheduler) dailyStockAlert() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			// Check if it's 8 PM (20:00)
			if now.Hour() == 20 && now.Minute() < 5 {
				s.sendDailyStockAlert()
			}
		case <-s.stopCh:
			return
		}
	}
}

// sendDailyStockAlert sends stock alert to all admins
func (s *Scheduler) sendDailyStockAlert() {
	// Get products with low stock (0 or <= 5)
	lowStockProducts, err := s.db.GetLowStockProducts(5)
	if err != nil {
		logrus.Errorf("Failed to get low stock products: %v", err)
		return
	}

	outOfStockProducts := make([]models.Product, 0)
	lowStockList := make([]models.Product, 0)

	for _, product := range lowStockProducts {
		if product.Stock == 0 {
			outOfStockProducts = append(outOfStockProducts, product)
		} else {
			lowStockList = append(lowStockList, product)
		}
	}

	// Only send alert if there are products with issues
	if len(outOfStockProducts) == 0 && len(lowStockList) == 0 {
		return
	}

	var alertText string
	alertText = "🚨 *LAPORAN STOK HARIAN*\n"
	alertText += fmt.Sprintf("📅 %s - 20:00 WIB\n\n", time.Now().Format("02/01/2006"))

	if len(outOfStockProducts) > 0 {
		alertText += "❌ *STOK HABIS:*\n"
		for _, product := range outOfStockProducts {
			alertText += fmt.Sprintf("• %s\n", product.Name)
		}
		alertText += "\n"
	}

	if len(lowStockList) > 0 {
		alertText += "⚠️ *STOK RENDAH:*\n"
		for _, product := range lowStockList {
			alertText += fmt.Sprintf("• %s (sisa: %d)\n", product.Name, product.Stock)
		}
		alertText += "\n"
	}

	alertText += "💡 *Rekomendasi:* Segera lakukan restock untuk produk yang stoknya habis atau rendah.\n\n"
	alertText += "Gunakan /admin untuk mengelola stok produk."

	// Send to all admins
	for _, adminID := range s.config.AdminIDs {
		msg := tgbotapi.NewMessage(adminID, alertText)
		msg.ParseMode = tgbotapi.ModeMarkdown

		_, err := s.api.Send(msg)
		if err != nil {
			logrus.Errorf("Failed to send stock alert to admin %d: %v", adminID, err)
		}
	}

	logrus.Infof("Daily stock alert sent to %d admins", len(s.config.AdminIDs))
}

// adminNotificationChecker checks for new paid orders to notify admins
func (s *Scheduler) adminNotificationChecker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkNewPaidOrders()
		case <-s.stopCh:
			return
		}
	}
}

// checkNewPaidOrders checks for newly paid orders and notifies admins
func (s *Scheduler) checkNewPaidOrders() {
	// Get orders that were paid in the last minute and haven't been notified
	rows, err := s.db.Query(`
		SELECT id, user_id, total_amount, completed_at
		FROM orders 
		WHERE payment_status = 'paid' 
		AND completed_at > datetime('now', '-1 minute')
		AND id NOT IN (
			SELECT DISTINCT interaction_data 
			FROM user_interactions 
			WHERE interaction_type = 'admin_notified' 
			AND created_at > datetime('now', '-1 hour')
		)
	`)
	if err != nil {
		logrus.Errorf("Failed to query new paid orders: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var orderID string
		var userID int64
		var totalAmount int
		var completedAt time.Time

		err := rows.Scan(&orderID, &userID, &totalAmount, &completedAt)
		if err != nil {
			logrus.Errorf("Failed to scan paid order: %v", err)
			continue
		}

		// Get user info
		user, _ := s.db.GetUser(userID)
		userName := "Unknown"
		if user != nil && user.FirstName != nil {
			userName = *user.FirstName
			if user.LastName != nil {
				userName += " " + *user.LastName
			}
		}

		// Send notification to all admins
		notificationText := fmt.Sprintf(`💰 *PEMBAYARAN BERHASIL*

🆔 Order: #%s
👤 Customer: %s (ID: %d)
💰 Total: %s
📅 Dibayar: %s

✅ Stok produk telah otomatis dikurangi.

Gunakan /admin untuk mengelola pesanan.`,
			orderID[:8],
			userName,
			userID,
			models.FormatPrice(totalAmount, s.config.CurrencySymbol),
			completedAt.Format("02/01/2006 15:04"))

		for _, adminID := range s.config.AdminIDs {
			msg := tgbotapi.NewMessage(adminID, notificationText)
			msg.ParseMode = tgbotapi.ModeMarkdown

			_, err := s.api.Send(msg)
			if err != nil {
				logrus.Errorf("Failed to send payment notification to admin %d: %v", adminID, err)
			}
		}

		// Mark as notified
		s.db.LogUserInteraction(userID, "admin_notified", orderID)

		logrus.Infof("Payment notification sent for order %s", orderID)
	}
}