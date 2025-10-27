package qris

import (
	"bytes"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/models"

	"github.com/google/uuid"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"github.com/nfnt/resize"
	"github.com/sirupsen/logrus"
	qrcodegen "github.com/skip2/go-qrcode"
)

// MerchantInfo holds merchant information extracted from QRIS
type MerchantInfo struct {
	MerchantID   string
	MerchantName string
	MerchantCity string
	CountryCode  string
	Currency     string
}

// RealQRISService handles real QRIS implementation with static QR upload and dynamic generation
type RealQRISService struct {
	config           *config.Config
	staticQRPayload  string
	merchantInfo     *MerchantInfo
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
func (q *RealQRISService) parseQRISPayload(payload string) (*MerchantInfo, error) {
	// Simple parser for QRIS payload using TLV format
	merchantInfo := &MerchantInfo{
		CountryCode:  "ID",
		Currency:     "360", // Indonesian Rupiah
	}

	// Extract merchant name (tag 59)
	if name := q.extractTLVValue(payload, "59"); name != "" {
		merchantInfo.MerchantName = name
	}

	// Extract merchant city (tag 60)
	if city := q.extractTLVValue(payload, "60"); city != "" {
		merchantInfo.MerchantCity = city
	}

	// Extract merchant ID from merchant account information (tag 26)
	if mai := q.extractTLVValue(payload, "26"); mai != "" {
		// Extract merchant ID from MAI (tag 01)
		if mid := q.extractTLVValue(mai, "01"); mid != "" {
			merchantInfo.MerchantID = mid
		}
	}

	// Validate that we extracted at least some info
	if merchantInfo.MerchantName == "" && merchantInfo.MerchantCity == "" {
		return nil, fmt.Errorf("failed to extract merchant information from QRIS")
	}

	return merchantInfo, nil
}

// extractTLVValue extracts value from TLV (Tag-Length-Value) format
func (q *RealQRISService) extractTLVValue(data, tag string) string {
	idx := strings.Index(data, tag)
	if idx == -1 || idx+4 > len(data) {
		return ""
	}

	// Get length (2 digits after tag)
	lengthStr := data[idx+2 : idx+4]
	length := 0
	fmt.Sscanf(lengthStr, "%d", &length)

	if length == 0 || idx+4+length > len(data) {
		return ""
	}

	return data[idx+4 : idx+4+length]
}

// modifyQRISPayload modifies static QRIS to dynamic with amount and order info
func (q *RealQRISService) modifyQRISPayload(staticPayload string, amount int, orderID string) string {
	// Start with static payload
	result := staticPayload

	// Replace Point of Initiation Method to dynamic (tag 01)
	result = q.replaceTLVValue(result, "01", "12")

	// Add/replace transaction amount (tag 54)
	amountStr := strconv.Itoa(amount)
	result = q.replaceTLVValue(result, "54", amountStr)

	// Build additional data field template (tag 62)
	additionalData := ""
	
	// Bill number (tag 01 within 62)
	billNumber := orderID[:min(len(orderID), 25)]
	additionalData += q.formatTLV("01", billNumber)
	
	// Reference label (tag 05 within 62)
	timestamp := time.Now().Format("060102150405")
	refLabel := fmt.Sprintf("%s-%s", orderID[:min(len(orderID), 8)], timestamp)
	additionalData += q.formatTLV("05", refLabel)
	
	// Replace additional data field
	result = q.replaceTLVValue(result, "62", additionalData)

	// Recalculate CRC (tag 63)
	// Remove old CRC
	if idx := strings.Index(result, "63"); idx != -1 {
		result = result[:idx]
	}
	
	// Add CRC placeholder and calculate
	result += "6304"
	crc := q.calculateCRC16(result)
	result += crc

	return result
}

// replaceTLVValue replaces or adds a TLV value
func (q *RealQRISService) replaceTLVValue(data, tag, value string) string {
	newTLV := q.formatTLV(tag, value)
	
	// Find and replace existing tag
	idx := strings.Index(data, tag)
	if idx != -1 && idx+2 < len(data) {
		// Get length of existing value
		lengthStr := data[idx+2 : idx+4]
		oldLength := 0
		fmt.Sscanf(lengthStr, "%d", &oldLength)
		
		if oldLength > 0 && idx+4+oldLength <= len(data) {
			// Replace existing TLV
			before := data[:idx]
			after := data[idx+4+oldLength:]
			return before + newTLV + after
		}
	}
	
	// Tag not found, append before CRC (tag 63)
	if crcIdx := strings.Index(data, "63"); crcIdx != -1 {
		return data[:crcIdx] + newTLV + data[crcIdx:]
	}
	
	// No CRC found, just append
	return data + newTLV
}

// formatTLV formats Tag-Length-Value according to EMV specification
func (q *RealQRISService) formatTLV(tag, value string) string {
	length := fmt.Sprintf("%02d", len(value))
	return tag + length + value
}

// calculateCRC16 calculates CRC-16 checksum for QRIS
func (q *RealQRISService) calculateCRC16(data string) string {
	// CRC-16-CCITT polynomial
	polynomial := uint16(0x1021)
	crc := uint16(0xFFFF)

	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc = crc << 1
			}
		}
	}

	return fmt.Sprintf("%04X", crc&0xFFFF)
}

// GenerateDynamicQRIS generates dynamic QRIS with specific amount
func (q *RealQRISService) GenerateDynamicQRIS(orderID string, amount int) (*models.QRISPayment, []byte, error) {
	if q.staticQRPayload == "" || q.merchantInfo == nil {
		return nil, nil, fmt.Errorf("static QRIS not configured. Please upload static QR first")
	}

	logrus.Infof("🔄 Generating dynamic QRIS for order %s with amount %d", orderID, amount)

	// Modify the static QRIS to make it dynamic
	dynamicPayload := q.modifyQRISPayload(q.staticQRPayload, amount, orderID)

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
func (q *RealQRISService) GetMerchantInfo() *MerchantInfo {
	if q.merchantInfo == nil {
		return &MerchantInfo{
			MerchantName: q.config.QRISMerchantName,
			MerchantCity: q.config.QRISCity,
			MerchantID:   q.config.QRISMerchantID,
			CountryCode:  q.config.QRISCountryCode,
			Currency:     q.config.QRISCurrencyCode,
		}
	}
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
func (q *RealQRISService) saveStaticQRConfig(payload string, merchantInfo *MerchantInfo) error {
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
		q.merchantInfo = &MerchantInfo{
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