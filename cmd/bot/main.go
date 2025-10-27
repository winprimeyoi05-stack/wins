package main

import (
	"os"
	"os/signal"
	"syscall"

	"telegram-premium-store/internal/bot"
	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/database"
	"telegram-premium-store/internal/payment"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Setup logging
	setupLogging(cfg.LogLevel)

	logrus.Info("🚀 Starting Telegram Premium Store Bot...")

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		logrus.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize payment service
	paymentService := payment.NewQRISService(cfg)

	// Initialize and start bot
	telegramBot, err := bot.New(cfg, db, paymentService)
	if err != nil {
		logrus.Fatalf("Failed to create bot: %v", err)
	}

	// Start bot in goroutine
	go func() {
		if err := telegramBot.Start(); err != nil {
			logrus.Fatalf("Bot error: %v", err)
		}
	}()

	logrus.Info("✅ Bot started successfully!")
	logrus.Info("📋 Features:")
	logrus.Info("   • Product catalog with categories")
	logrus.Info("   • Shopping cart system")
	logrus.Info("   • Dynamic QRIS payment")
	logrus.Info("   • Admin panel")
	logrus.Info("   • Indonesian language support")

	// Wait for interrupt signal to gracefully shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logrus.Info("🛑 Shutting down bot...")
	telegramBot.Stop()
	logrus.Info("👋 Bot stopped gracefully")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		ForceColors:   true,
	})

	switch level {
	case "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}