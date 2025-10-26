package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"telegram-premium-store/internal/models"
)

// PaymentVerifier handles payment verification and anti-manipulation
type PaymentVerifier struct {
	secretKey string
}

// NewPaymentVerifier creates a new payment verifier
func NewPaymentVerifier(secretKey string) *PaymentVerifier {
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}
	return &PaymentVerifier{
		secretKey: secretKey,
	}
}

// GenerateVerificationHash generates verification hash for QRIS payment
func (v *PaymentVerifier) GenerateVerificationHash(orderID string, amount int, qrisPayload string) string {
	// Create data string for hashing
	data := fmt.Sprintf("%s:%d:%s", orderID, amount, qrisPayload)
	
	// Generate HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte(v.secretKey))
	h.Write([]byte(data))
	hash := hex.EncodeToString(h.Sum(nil))
	
	return hash[:16] // Take first 16 characters for brevity
}

// VerifyQRISPayment verifies QRIS payment against manipulation
func (v *PaymentVerifier) VerifyQRISPayment(orderID string, expectedAmount int, qrisPayload string, providedHash string) error {
	// Generate expected hash
	expectedHash := v.GenerateVerificationHash(orderID, expectedAmount, qrisPayload)
	
	// Compare hashes
	if expectedHash != providedHash {
		return fmt.Errorf("payment verification failed: hash mismatch")
	}

	// Extract amount from QRIS payload
	extractedAmount, err := v.extractAmountFromQRIS(qrisPayload)
	if err != nil {
		return fmt.Errorf("failed to extract amount from QRIS: %w", err)
	}

	// Verify amount matches
	if extractedAmount != expectedAmount {
		return fmt.Errorf("amount manipulation detected: expected %d, found %d in QRIS", expectedAmount, extractedAmount)
	}

	return nil
}

// extractAmountFromQRIS extracts transaction amount from QRIS payload
func (v *PaymentVerifier) extractAmountFromQRIS(qrisPayload string) (int, error) {
	// QRIS format: ...54[length][amount]...
	// Find transaction amount field (ID "54")
	amountIndex := strings.Index(qrisPayload, "54")
	if amountIndex == -1 {
		return 0, fmt.Errorf("transaction amount field not found in QRIS")
	}

	// Skip "54" and get length
	if len(qrisPayload) < amountIndex+4 {
		return 0, fmt.Errorf("invalid QRIS format: insufficient length")
	}

	lengthStr := qrisPayload[amountIndex+2 : amountIndex+4]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return 0, fmt.Errorf("invalid amount length in QRIS: %s", lengthStr)
	}

	// Extract amount value
	if len(qrisPayload) < amountIndex+4+length {
		return 0, fmt.Errorf("invalid QRIS format: amount field truncated")
	}

	amountStr := qrisPayload[amountIndex+4 : amountIndex+4+length]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return 0, fmt.Errorf("invalid amount value in QRIS: %s", amountStr)
	}

	return amount, nil
}

// ValidateQRISIntegrity validates QRIS payload integrity
func (v *PaymentVerifier) ValidateQRISIntegrity(qrisPayload string) error {
	// Basic QRIS format validation
	if len(qrisPayload) < 50 {
		return fmt.Errorf("QRIS payload too short")
	}

	// Check for required fields
	requiredFields := []string{
		"00", // Payload Format Indicator
		"01", // Point of Initiation Method
		"52", // Merchant Category Code
		"53", // Transaction Currency
		"54", // Transaction Amount
		"58", // Country Code
		"59", // Merchant Name
		"63", // CRC
	}

	for _, field := range requiredFields {
		if !strings.Contains(qrisPayload, field) {
			return fmt.Errorf("required QRIS field %s not found", field)
		}
	}

	return nil
}

// DetectQRISManipulation detects potential QRIS manipulation
func (v *PaymentVerifier) DetectQRISManipulation(originalPayload, receivedPayload string) []string {
	var manipulations []string

	// Extract amounts from both payloads
	originalAmount, err1 := v.extractAmountFromQRIS(originalPayload)
	receivedAmount, err2 := v.extractAmountFromQRIS(receivedPayload)

	if err1 != nil || err2 != nil {
		manipulations = append(manipulations, "Failed to extract amounts for comparison")
		return manipulations
	}

	// Check amount manipulation
	if originalAmount != receivedAmount {
		manipulations = append(manipulations, 
			fmt.Sprintf("Amount changed: %d → %d", originalAmount, receivedAmount))
	}

	// Check payload length changes (might indicate other manipulations)
	if len(originalPayload) != len(receivedPayload) {
		manipulations = append(manipulations, 
			fmt.Sprintf("Payload length changed: %d → %d", len(originalPayload), len(receivedPayload)))
	}

	// Check for common manipulation patterns
	if strings.Count(originalPayload, "54") != strings.Count(receivedPayload, "54") {
		manipulations = append(manipulations, "Transaction amount field count mismatch")
	}

	return manipulations
}

// CreateSecurePaymentToken creates secure payment token for order
func (v *PaymentVerifier) CreateSecurePaymentToken(orderID string, amount int, userID int64) string {
	data := fmt.Sprintf("%s:%d:%d:%d", orderID, amount, userID, time.Now().Unix())
	h := hmac.New(sha256.New, []byte(v.secretKey))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))[:24] // 24 character token
}

// ValidatePaymentToken validates payment token
func (v *PaymentVerifier) ValidatePaymentToken(token, orderID string, amount int, userID int64) bool {
	// For demo purposes, we'll accept tokens created within last hour
	// In production, you'd store and validate against database
	
	// Generate expected token for current time (simplified validation)
	currentTime := time.Now().Unix()
	for i := 0; i < 3600; i++ { // Check last hour (3600 seconds)
		testTime := currentTime - int64(i)
		data := fmt.Sprintf("%s:%d:%d:%d", orderID, amount, userID, testTime)
		h := hmac.New(sha256.New, []byte(v.secretKey))
		h.Write([]byte(data))
		expectedToken := hex.EncodeToString(h.Sum(nil))[:24]
		
		if expectedToken == token {
			return true
		}
	}
	
	return false
}