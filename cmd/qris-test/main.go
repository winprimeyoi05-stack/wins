package main

import (
	"fmt"
	"log"
	"os"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/qris"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize QRIS service
	qrisService := qris.NewRealQRISService(cfg)

	if len(os.Args) < 2 {
		showUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "upload":
		if len(os.Args) < 3 {
			fmt.Println("❌ Usage: qris-test upload <image_path>")
			return
		}
		uploadQRIS(qrisService, os.Args[2])

	case "generate":
		if len(os.Args) < 4 {
			fmt.Println("❌ Usage: qris-test generate <order_id> <amount>")
			return
		}
		generateQRIS(qrisService, os.Args[2], os.Args[3])

	case "status":
		showStatus(qrisService)

	case "test":
		testQRIS(qrisService)

	default:
		showUsage()
	}
}

func showUsage() {
	fmt.Println(`🔧 QRIS Test Tool

Usage:
  qris-test upload <image_path>     Upload and process static QR image
  qris-test generate <order_id> <amount>  Generate dynamic QRIS
  qris-test status                  Show QRIS configuration status
  qris-test test                    Generate test QRIS

Examples:
  qris-test upload qr_static.png
  qris-test generate ORD-123 50000
  qris-test status
  qris-test test`)
}

func uploadQRIS(service *qris.RealQRISService, imagePath string) {
	fmt.Printf("🔄 Processing QRIS image: %s\n", imagePath)

	err := service.ProcessQRISImageFromFile(imagePath)
	if err != nil {
		fmt.Printf("❌ Failed to process QRIS: %v\n", err)
		return
	}

	fmt.Println("✅ QRIS static payload successfully processed!")
	showStatus(service)
}

func generateQRIS(service *qris.RealQRISService, orderID, amountStr string) {
	if !service.IsConfigured() {
		fmt.Println("❌ QRIS not configured. Upload static QR first.")
		return
	}

	var amount int
	if _, err := fmt.Sscanf(amountStr, "%d", &amount); err != nil {
		fmt.Printf("❌ Invalid amount: %s\n", amountStr)
		return
	}

	fmt.Printf("🔄 Generating dynamic QRIS for order %s with amount %d\n", orderID, amount)

	qrisPayment, qrImage, err := service.GenerateDynamicQRIS(orderID, amount)
	if err != nil {
		fmt.Printf("❌ Failed to generate QRIS: %v\n", err)
		return
	}

	// Save QR image
	filename := fmt.Sprintf("qris_%s.png", orderID)
	if err := os.WriteFile(filename, qrImage, 0644); err != nil {
		fmt.Printf("⚠️ Failed to save QR image: %v\n", err)
	} else {
		fmt.Printf("💾 QR image saved: %s\n", filename)
	}

	fmt.Println("✅ Dynamic QRIS generated successfully!")
	fmt.Printf("🆔 Order ID: %s\n", qrisPayment.OrderID)
	fmt.Printf("💰 Amount: Rp %d\n", qrisPayment.Amount)
	fmt.Printf("🏪 Merchant: %s\n", qrisPayment.MerchantName)
	fmt.Printf("⏰ Expires: %s\n", qrisPayment.ExpiryTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("📄 Payload Length: %d characters\n", len(qrisPayment.QRString))
}

func showStatus(service *qris.RealQRISService) {
	fmt.Println("📋 QRIS Configuration Status:")
	fmt.Println(service.GetStaticQRStatus())

	if service.IsConfigured() {
		merchantInfo := service.GetMerchantInfo()
		fmt.Printf("\n🔍 Technical Details:\n")
		fmt.Printf("   Merchant ID: %s\n", merchantInfo.MerchantID)
		fmt.Printf("   Country Code: %s\n", merchantInfo.CountryCode)
		fmt.Printf("   Currency: %s\n", merchantInfo.Currency)
	}
}

func testQRIS(service *qris.RealQRISService) {
	if !service.IsConfigured() {
		fmt.Println("❌ QRIS not configured. Upload static QR first.")
		return
	}

	testOrderID := "TEST-" + service.GenerateOrderID()
	testAmount := 10000

	fmt.Printf("🧪 Generating test QRIS (Order: %s, Amount: Rp %d)\n", testOrderID, testAmount)

	qrisPayment, qrImage, err := service.GenerateDynamicQRIS(testOrderID, testAmount)
	if err != nil {
		fmt.Printf("❌ Failed to generate test QRIS: %v\n", err)
		return
	}

	// Save test QR image
	filename := fmt.Sprintf("test_qris_%s.png", testOrderID)
	if err := os.WriteFile(filename, qrImage, 0644); err != nil {
		fmt.Printf("⚠️ Failed to save test QR image: %v\n", err)
	} else {
		fmt.Printf("💾 Test QR image saved: %s\n", filename)
	}

	fmt.Println("✅ Test QRIS generated successfully!")
	fmt.Printf("🆔 Order ID: %s\n", qrisPayment.OrderID)
	fmt.Printf("💰 Amount: Rp %d\n", qrisPayment.Amount)
	fmt.Printf("⏰ Expires: %s\n", qrisPayment.ExpiryTime.Format("15:04:05"))
	fmt.Println("💡 Try scanning with your e-wallet app to test!")
}