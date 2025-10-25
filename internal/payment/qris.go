package payment

import (
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"

	"telegram-premium-store/internal/config"
	"telegram-premium-store/internal/models"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// QRISService handles QRIS payment generation and processing
type QRISService struct {
	config *config.Config
}

// NewQRISService creates a new QRIS service
func NewQRISService(cfg *config.Config) *QRISService {
	return &QRISService{
		config: cfg,
	}
}

// GenerateQRIS generates a QRIS payment QR code for the given order
func (q *QRISService) GenerateQRIS(orderID string, amount int) (*models.QRISPayment, []byte, error) {
	// Create QRIS payment data
	payment := &models.QRISPayment{
		OrderID:      orderID,
		Amount:       amount,
		MerchantID:   q.config.QRISMerchantID,
		MerchantName: q.config.QRISMerchantName,
		City:         q.config.QRISCity,
		CountryCode:  q.config.QRISCountryCode,
		CurrencyCode: q.config.QRISCurrencyCode,
		ExpiryTime:   time.Now().Add(15 * time.Minute), // 15 minutes expiry
	}

	// Generate QRIS string
	qrisString, err := q.generateQRISString(payment)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate QRIS string: %w", err)
	}
	payment.QRString = qrisString

	// Generate QR code image
	qrImage, err := qrcode.Encode(qrisString, qrcode.Medium, 256)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate QR code: %w", err)
	}

	return payment, qrImage, nil
}

// generateQRISString creates the QRIS payment string according to EMV QR Code specification
func (q *QRISService) generateQRISString(payment *models.QRISPayment) (string, error) {
	var qris strings.Builder

	// Payload Format Indicator (ID "00")
	qris.WriteString(q.formatTLV("00", "01"))

	// Point of Initiation Method (ID "01") - Static QR = "11", Dynamic QR = "12"
	qris.WriteString(q.formatTLV("01", "12"))

	// Merchant Account Information (ID "26" for QRIS Indonesia)
	merchantInfo := q.buildMerchantAccountInfo(payment)
	qris.WriteString(q.formatTLV("26", merchantInfo))

	// Merchant Category Code (ID "52")
	qris.WriteString(q.formatTLV("52", "0000"))

	// Transaction Currency (ID "53") - 360 for Indonesian Rupiah
	qris.WriteString(q.formatTLV("53", payment.CurrencyCode))

	// Transaction Amount (ID "54")
	amountStr := fmt.Sprintf("%.0f", float64(payment.Amount))
	qris.WriteString(q.formatTLV("54", amountStr))

	// Country Code (ID "58")
	qris.WriteString(q.formatTLV("58", payment.CountryCode))

	// Merchant Name (ID "59")
	qris.WriteString(q.formatTLV("59", payment.MerchantName))

	// Merchant City (ID "60")
	qris.WriteString(q.formatTLV("60", payment.City))

	// Additional Data Field Template (ID "62")
	additionalData := q.buildAdditionalData(payment)
	qris.WriteString(q.formatTLV("62", additionalData))

	// CRC (ID "63") - will be calculated and appended
	qrisWithoutCRC := qris.String() + "6304"
	crc := q.calculateCRC16(qrisWithoutCRC)
	qris.WriteString(fmt.Sprintf("63%02X%s", len(crc), crc))

	return qris.String(), nil
}

// buildMerchantAccountInfo builds the merchant account information field
func (q *QRISService) buildMerchantAccountInfo(payment *models.QRISPayment) string {
	var info strings.Builder

	// Global Unique Identifier (ID "00")
	info.WriteString(q.formatTLV("00", "ID.CO.QRIS.WWW"))

	// Merchant PAN (ID "01")
	info.WriteString(q.formatTLV("01", payment.MerchantID))

	// Merchant ID (ID "02")
	info.WriteString(q.formatTLV("02", "UMI"))

	// Merchant Criteria (ID "03")
	info.WriteString(q.formatTLV("03", "UMI"))

	return info.String()
}

// buildAdditionalData builds the additional data field
func (q *QRISService) buildAdditionalData(payment *models.QRISPayment) string {
	var data strings.Builder

	// Bill Number (ID "01") - using order ID
	data.WriteString(q.formatTLV("01", payment.OrderID))

	// Reference Label (ID "05") - using order ID with timestamp
	refLabel := fmt.Sprintf("%s-%d", payment.OrderID[:8], time.Now().Unix())
	data.WriteString(q.formatTLV("05", refLabel))

	// Terminal Label (ID "07")
	data.WriteString(q.formatTLV("07", "STORE01"))

	return data.String()
}

// formatTLV formats Tag-Length-Value according to EMV specification
func (q *QRISService) formatTLV(tag, value string) string {
	length := fmt.Sprintf("%02d", len(value))
	return tag + length + value
}

// calculateCRC16 calculates CRC-16 checksum for QRIS
func (q *QRISService) calculateCRC16(data string) string {
	// Simple CRC-16 implementation for demonstration
	// In production, use proper CRC-16-CCITT implementation
	hash := md5.Sum([]byte(data))
	crc := fmt.Sprintf("%04X", int(hash[0])<<8|int(hash[1]))
	return crc
}

// ValidatePayment validates a QRIS payment (mock implementation)
func (q *QRISService) ValidatePayment(orderID string, amount int) (bool, error) {
	// In a real implementation, this would:
	// 1. Check with the payment processor API
	// 2. Verify the transaction status
	// 3. Validate the amount and order ID
	// 4. Return the actual payment status

	// For demo purposes, we'll simulate payment validation
	// In production, integrate with actual QRIS payment gateway like:
	// - Bank Indonesia QRIS
	// - Midtrans QRIS
	// - Xendit QRIS
	// - DANA Business
	// - OVO Business
	// - GoPay Business

	return true, nil // Mock: always return success
}

// CheckPaymentStatus checks the status of a QRIS payment
func (q *QRISService) CheckPaymentStatus(orderID string) (models.PaymentStatus, error) {
	// Mock implementation - in production, this would call the payment gateway API
	// to check the actual payment status

	// Simulate different payment statuses based on order ID pattern
	// This is just for demonstration purposes
	if strings.Contains(orderID, "test") {
		return models.PaymentStatusPaid, nil
	}

	// For demo, we'll randomly return different statuses
	// In production, this should call the actual payment gateway
	return models.PaymentStatusPending, nil
}

// ProcessCallback processes payment callback from QRIS provider
func (q *QRISService) ProcessCallback(callbackData map[string]interface{}) (*PaymentCallback, error) {
	// Extract callback data
	orderID, ok := callbackData["order_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing order_id in callback")
	}

	status, ok := callbackData["status"].(string)
	if !ok {
		return nil, fmt.Errorf("missing status in callback")
	}

	amount, ok := callbackData["amount"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing amount in callback")
	}

	// Validate signature (in production)
	// signature := callbackData["signature"].(string)
	// if !q.validateSignature(callbackData, signature) {
	//     return nil, fmt.Errorf("invalid signature")
	// }

	callback := &PaymentCallback{
		OrderID:   orderID,
		Status:    q.mapCallbackStatus(status),
		Amount:    int(amount),
		Timestamp: time.Now(),
		Raw:       callbackData,
	}

	return callback, nil
}

// PaymentCallback represents a payment callback from QRIS provider
type PaymentCallback struct {
	OrderID   string                 `json:"order_id"`
	Status    models.PaymentStatus   `json:"status"`
	Amount    int                    `json:"amount"`
	Timestamp time.Time              `json:"timestamp"`
	Raw       map[string]interface{} `json:"raw"`
}

// mapCallbackStatus maps provider status to our internal status
func (q *QRISService) mapCallbackStatus(providerStatus string) models.PaymentStatus {
	switch strings.ToLower(providerStatus) {
	case "success", "paid", "completed":
		return models.PaymentStatusPaid
	case "pending", "processing":
		return models.PaymentStatusPending
	case "expired", "timeout":
		return models.PaymentStatusExpired
	case "cancelled", "canceled":
		return models.PaymentStatusCancelled
	case "refunded":
		return models.PaymentStatusRefunded
	default:
		return models.PaymentStatusPending
	}
}

// GenerateOrderID generates a unique order ID
func (q *QRISService) GenerateOrderID() string {
	// Generate UUID and take first 8 characters + timestamp
	id := uuid.New().String()
	timestamp := strconv.FormatInt(time.Now().Unix(), 36)
	return fmt.Sprintf("ORD-%s-%s", id[:8], timestamp)
}

// IsExpired checks if a QRIS payment has expired
func (q *QRISService) IsExpired(expiryTime *time.Time) bool {
	if expiryTime == nil {
		return false
	}
	return time.Now().After(*expiryTime)
}

// GetPaymentInstructions returns localized payment instructions
func (q *QRISService) GetPaymentInstructions(orderID string, amount int) string {
	messages := config.GetMessages()
	return fmt.Sprintf(messages.PaymentInstructions,
		models.FormatPrice(amount, q.config.CurrencySymbol),
		q.config.CurrencySymbol,
		orderID)
}

// GetSupportedBanks returns list of banks/e-wallets that support QRIS
func (q *QRISService) GetSupportedBanks() []string {
	return []string{
		"🏦 BCA Mobile",
		"🏦 BNI Mobile Banking",
		"🏦 BRI Mobile",
		"🏦 Mandiri Online",
		"🏦 CIMB Niaga",
		"💳 DANA",
		"💳 OVO", 
		"💳 GoPay",
		"💳 LinkAja",
		"💳 ShopeePay",
		"💳 Jenius",
		"💳 Sakuku",
		"💳 i.saku",
		"💳 DOKU Wallet",
	}
}