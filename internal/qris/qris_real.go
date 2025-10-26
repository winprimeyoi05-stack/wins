package qris

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/models"

	"github.com/fyvri/go-qris"
	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	qrcodegen "github.com/skip2/go-qrcode"
)

// RealQRISService handles real QRIS implementation with static QR upload and dynamic generation
type RealQRISService struct {
	config           *config.Config
	staticQRPayload  string
	merchantInfo     *qris.MerchantInfo
	qrisUploadDir    string
	qrisGeneratedDir string
}

// NewRealQRISService creates a new real QRIS service
func NewRealQRISService(cfg *config.Config) *RealQRISService {
	service := &RealQRISService{
		config:           cfg,
		qrisUploadDir:    "uploads/qris",
		qrisGeneratedDir: "generated/qris",
	}

	// Create directories if not exist
	os.MkdirAll(service.qrisUploadDir, 0755)
	os.MkdirAll(service.qrisGeneratedDir, 0755)

	// Load existing static QR if available
	service.loadStaticQRPayload()

	return service
}

// UploadStaticQR uploads and processes static QR code image
func (q *RealQRISService) UploadStaticQR(imageData []byte, filename string) error {
	logrus.Info("🔄 Processing uploaded QRIS image...")

	// Save uploaded image
	uploadPath := filepath.Join(q.qrisUploadDir, filename)
	if err := os.WriteFile(uploadPath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to save uploaded image: %w", err)
	}

	// Decode QR code from image
	payload, err := q.decodeQRFromImage(imageData)
	if err != nil {
		return fmt.Errorf("failed to decode QR code: %w", err)
	}

	// Validate QRIS payload
	if !q.isValidQRISPayload(payload) {
		return fmt.Errorf("invalid QRIS payload format")
	}

	// Parse QRIS to extract merchant info
	merchantInfo, err := q.parseQRISPayload(payload)
	if err != nil {
		return fmt.Errorf("failed to parse QRIS payload: %w", err)
	}

	// Store the static payload and merchant info
	q.staticQRPayload = payload
	q.merchantInfo = merchantInfo

	// Save to config file for persistence
	if err := q.saveStaticQRConfig(payload, merchantInfo); err != nil {
		logrus.Warnf("Failed to save QRIS config: %v", err)
	}

	logrus.Info("✅ QRIS static payload successfully extracted and stored")
	logrus.Infof("📋 Merchant: %s", merchantInfo.MerchantName)
	logrus.Infof("🏪 City: %s", merchantInfo.MerchantCity)
	logrus.Infof("🆔 ID: %s", merchantInfo.MerchantID)

	return nil
}

// decodeQRFromImage decodes QR code from image bytes
func (q *RealQRISService) decodeQRFromImage(imageData []byte) (string, error) {
	// Decode image
	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image if too large (for better QR detection)
	if img.Bounds().Dx() > 1000 || img.Bounds().Dy() > 1000 {
		img = resize.Resize(800, 0, img, resize.Lanczos3)
	}

	// Convert to luminance source
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", fmt.Errorf("failed to create binary bitmap: %w", err)
	}

	// Decode QR code
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decode QR code: %w", err)
	}

	return result.GetText(), nil
}

// isValidQRISPayload validates if the payload is a valid QRIS format
func (q *RealQRISService) isValidQRISPayload(payload string) bool {
	// Basic QRIS validation
	if len(payload) < 50 {
		return false
	}

	// Check for QRIS indicators
	if !strings.Contains(payload, "00020") { // Payload Format Indicator
		return false
	}

	if !strings.Contains(payload, "010212") { // Point of Initiation Method (Dynamic)
		if !strings.Contains(payload, "010211") { // Point of Initiation Method (Static)
			return false
		}
	}

	// Check for Indonesian QRIS identifier
	if !strings.Contains(payload, "ID.CO.QRIS") {
		return false
	}

	return true
}

// parseQRISPayload parses QRIS payload to extract merchant information
func (q *RealQRISService) parseQRISPayload(payload string) (*qris.MerchantInfo, error) {
	// Use go-qris library to parse
	qrisData, err := qris.Decode(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QRIS: %w", err)
	}

	merchantInfo := &qris.MerchantInfo{
		MerchantID:   qrisData.MerchantAccountInformation.MerchantID,
		MerchantName: qrisData.MerchantName,
		MerchantCity: qrisData.MerchantCity,
		CountryCode:  qrisData.CountryCode,
		Currency:     qrisData.TransactionCurrency,
	}

	return merchantInfo, nil
}

// GenerateDynamicQRIS generates dynamic QRIS with specific amount
func (q *RealQRISService) GenerateDynamicQRIS(orderID string, amount int) (*models.QRISPayment, []byte, error) {
	if q.staticQRPayload == "" || q.merchantInfo == nil {
		return nil, nil, fmt.Errorf("static QRIS not configured. Please upload static QR first")
	}

	logrus.Infof("🔄 Generating dynamic QRIS for order %s with amount %d", orderID, amount)

	// Parse the static QRIS
	staticQRIS, err := qris.Decode(q.staticQRPayload)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode static QRIS: %w", err)
	}

	// Create dynamic QRIS with new amount
	dynamicQRIS := *staticQRIS

	// Set dynamic values
	dynamicQRIS.PointOfInitiationMethod = "12" // Dynamic QR
	dynamicQRIS.TransactionAmount = strconv.Itoa(amount)

	// Add additional data for order tracking
	if dynamicQRIS.AdditionalDataFieldTemplate == nil {
		dynamicQRIS.AdditionalDataFieldTemplate = &qris.AdditionalDataFieldTemplate{}
	}

	// Set bill number (order ID)
	dynamicQRIS.AdditionalDataFieldTemplate.BillNumber = orderID[:min(len(orderID), 25)] // Max 25 chars

	// Set reference label with timestamp
	timestamp := time.Now().Format("060102150405")
	refLabel := fmt.Sprintf("%s-%s", orderID[:min(len(orderID), 8)], timestamp)
	dynamicQRIS.AdditionalDataFieldTemplate.ReferenceLabel = refLabel

	// Generate new QRIS string
	dynamicPayload, err := qris.Encode(&dynamicQRIS)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encode dynamic QRIS: %w", err)
	}

	// Generate QR code image
	qrImage, err := qrcodegen.Encode(dynamicPayload, qrcodegen.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Save generated QR image for reference
	qrFilename := fmt.Sprintf("qris_%s_%d.png", orderID, time.Now().Unix())
	qrPath := filepath.Join(q.qrisGeneratedDir, qrFilename)
	if err := os.WriteFile(qrPath, qrImage, 0644); err != nil {
		logrus.Warnf("Failed to save generated QR image: %v", err)
	}

	// Create QRIS payment data
	payment := &models.QRISPayment{
		OrderID:      orderID,
		Amount:       amount,
		MerchantID:   q.merchantInfo.MerchantID,
		MerchantName: q.merchantInfo.MerchantName,
		City:         q.merchantInfo.MerchantCity,
		CountryCode:  q.merchantInfo.CountryCode,
		CurrencyCode: q.merchantInfo.Currency,
		QRString:     dynamicPayload,
		ExpiryTime:   time.Now().Add(5 * time.Minute), // 5 minutes expiry
	}

	logrus.Info("✅ Dynamic QRIS generated successfully")
	logrus.Infof("💰 Amount: %d", amount)
	logrus.Infof("🆔 Order ID: %s", orderID)
	logrus.Infof("⏰ Expires: %s", payment.ExpiryTime.Format("15:04:05"))

	return payment, qrImage, nil
}

// GetMerchantInfo returns current merchant information
func (q *RealQRISService) GetMerchantInfo() *qris.MerchantInfo {
	return q.merchantInfo
}

// IsConfigured checks if QRIS is properly configured
func (q *RealQRISService) IsConfigured() bool {
	return q.staticQRPayload != "" && q.merchantInfo != nil
}

// GetStaticQRStatus returns status of static QR configuration
func (q *RealQRISService) GetStaticQRStatus() string {
	if !q.IsConfigured() {
		return "❌ QRIS belum dikonfigurasi. Upload QR Code statis terlebih dahulu."
	}

	return fmt.Sprintf(`✅ QRIS sudah dikonfigurasi
🏪 Merchant: %s
🏙️ Kota: %s
🆔 ID: %s
💳 Currency: %s`,
		q.merchantInfo.MerchantName,
		q.merchantInfo.MerchantCity,
		q.merchantInfo.MerchantID,
		q.merchantInfo.Currency)
}

// saveStaticQRConfig saves static QR configuration to file
func (q *RealQRISService) saveStaticQRConfig(payload string, merchantInfo *qris.MerchantInfo) error {
	configPath := filepath.Join(q.qrisUploadDir, "qris_config.txt")
	
	config := fmt.Sprintf(`# QRIS Configuration
# Generated: %s

STATIC_PAYLOAD=%s
MERCHANT_ID=%s
MERCHANT_NAME=%s
MERCHANT_CITY=%s
COUNTRY_CODE=%s
CURRENCY=%s
`,
		time.Now().Format("2006-01-02 15:04:05"),
		payload,
		merchantInfo.MerchantID,
		merchantInfo.MerchantName,
		merchantInfo.MerchantCity,
		merchantInfo.CountryCode,
		merchantInfo.Currency)

	return os.WriteFile(configPath, []byte(config), 0644)
}

// loadStaticQRPayload loads existing static QR configuration
func (q *RealQRISService) loadStaticQRPayload() {
	configPath := filepath.Join(q.qrisUploadDir, "qris_config.txt")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return // No existing config
	}

	lines := strings.Split(string(data), "\n")
	var payload, merchantID, merchantName, merchantCity, countryCode, currency string

	for _, line := range lines {
		if strings.HasPrefix(line, "STATIC_PAYLOAD=") {
			payload = strings.TrimPrefix(line, "STATIC_PAYLOAD=")
		} else if strings.HasPrefix(line, "MERCHANT_ID=") {
			merchantID = strings.TrimPrefix(line, "MERCHANT_ID=")
		} else if strings.HasPrefix(line, "MERCHANT_NAME=") {
			merchantName = strings.TrimPrefix(line, "MERCHANT_NAME=")
		} else if strings.HasPrefix(line, "MERCHANT_CITY=") {
			merchantCity = strings.TrimPrefix(line, "MERCHANT_CITY=")
		} else if strings.HasPrefix(line, "COUNTRY_CODE=") {
			countryCode = strings.TrimPrefix(line, "COUNTRY_CODE=")
		} else if strings.HasPrefix(line, "CURRENCY=") {
			currency = strings.TrimPrefix(line, "CURRENCY=")
		}
	}

	if payload != "" && merchantID != "" {
		q.staticQRPayload = payload
		q.merchantInfo = &qris.MerchantInfo{
			MerchantID:   merchantID,
			MerchantName: merchantName,
			MerchantCity: merchantCity,
			CountryCode:  countryCode,
			Currency:     currency,
		}

		logrus.Info("✅ Loaded existing QRIS configuration")
		logrus.Infof("🏪 Merchant: %s", merchantName)
	}
}

// ValidateQRISImage validates uploaded image for QRIS processing
func (q *RealQRISService) ValidateQRISImage(imageData []byte) error {
	// Check file size (max 5MB)
	if len(imageData) > 5*1024*1024 {
		return fmt.Errorf("file terlalu besar. Maksimal 5MB")
	}

	// Check if it's a valid image
	_, format, err := image.DecodeConfig(bytes.NewReader(imageData))
	if err != nil {
		return fmt.Errorf("file bukan gambar yang valid")
	}

	// Check supported formats
	if format != "jpeg" && format != "jpg" && format != "png" {
		return fmt.Errorf("format tidak didukung. Gunakan JPEG atau PNG")
	}

	return nil
}

// ProcessQRISImageFromFile processes QRIS image from file path
func (q *RealQRISService) ProcessQRISImageFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	filename := filepath.Base(filePath)
	return q.UploadStaticQR(data, filename)
}

// GetSupportedBanks returns list of banks/e-wallets that support QRIS
func (q *RealQRISService) GetSupportedBanks() []string {
	return []string{
		"🏦 BCA Mobile",
		"🏦 BNI Mobile Banking",
		"🏦 BRI Mobile",
		"🏦 Mandiri Online",
		"🏦 CIMB Niaga",
		"🏦 Permata Mobile",
		"🏦 Danamon D-Bank",
		"🏦 OCBC OneB",
		"💳 DANA",
		"💳 OVO",
		"💳 GoPay",
		"💳 LinkAja",
		"💳 ShopeePay",
		"💳 Jenius",
		"💳 Sakuku",
		"💳 i.saku",
		"💳 DOKU Wallet",
		"💳 Flip",
		"💳 Bibit",
		"💳 Akulaku PayLater",
	}
}

// GetPaymentInstructions returns localized payment instructions
func (q *RealQRISService) GetPaymentInstructions(orderID string, amount int) string {
	return fmt.Sprintf(`💳 *INSTRUKSI PEMBAYARAN QRIS*

1️⃣ Buka aplikasi e-wallet atau mobile banking Anda
2️⃣ Pilih menu "Scan QR" atau "QRIS"
3️⃣ Scan QR Code di atas
4️⃣ Pastikan nominal: *Rp %s*
5️⃣ Pastikan merchant: *%s*
6️⃣ Konfirmasi pembayaran
7️⃣ Pembayaran akan otomatis terverifikasi

⚠️ *PENTING:*
• QR Code berlaku selama 5 menit
• Jangan ubah nominal pembayaran
• Order ID: *%s*
• Simpan screenshot untuk referensi

🔄 Status pembayaran akan diupdate otomatis setelah transaksi berhasil.`,
		models.FormatPrice(amount, "Rp"),
		q.merchantInfo.MerchantName,
		orderID)
}

// IsExpired checks if a QRIS payment has expired
func (q *RealQRISService) IsExpired(expiryTime *time.Time) bool {
	if expiryTime == nil {
		return false
	}
	return time.Now().After(*expiryTime)
}

// GenerateOrderID generates a unique order ID
func (q *RealQRISService) GenerateOrderID() string {
	// Generate UUID and take first 8 characters + timestamp
	id := uuid.New().String()
	timestamp := strconv.FormatInt(time.Now().Unix(), 36)
	return fmt.Sprintf("ORD-%s-%s", id[:8], timestamp)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}