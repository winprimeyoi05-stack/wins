package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// User represents a Telegram user
type User struct {
	UserID    int64     `json:"user_id" db:"user_id"`
	Username  *string   `json:"username" db:"username"`
	FirstName *string   `json:"first_name" db:"first_name"`
	LastName  *string   `json:"last_name" db:"last_name"`
	JoinDate  time.Time `json:"join_date" db:"join_date"`
	IsAdmin   bool      `json:"is_admin" db:"is_admin"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Product represents a premium application for sale
type Product struct {
	ID          int       `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       int       `json:"price" db:"price"` // Price in smallest currency unit (e.g., cents)
	Category    string    `json:"category" db:"category"`
	ImageURL    *string   `json:"image_url" db:"image_url"`
	DownloadURL *string   `json:"download_url" db:"download_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Stock       int       `json:"stock" db:"stock"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CartItem represents an item in user's shopping cart
type CartItem struct {
	ID        int       `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	ProductID int       `json:"product_id" db:"product_id"`
	Quantity  int       `json:"quantity" db:"quantity"`
	AddedAt   time.Time `json:"added_at" db:"added_at"`
	
	// Joined fields from Product
	ProductName  string  `json:"product_name,omitempty" db:"product_name"`
	ProductPrice int     `json:"product_price,omitempty" db:"product_price"`
	ProductImage *string `json:"product_image,omitempty" db:"product_image"`
}

// Order represents a purchase order
type Order struct {
	ID            string       `json:"id" db:"id"` // UUID
	UserID        int64        `json:"user_id" db:"user_id"`
	TotalAmount   int          `json:"total_amount" db:"total_amount"`
	PaymentMethod string       `json:"payment_method" db:"payment_method"`
	PaymentStatus PaymentStatus `json:"payment_status" db:"payment_status"`
	QRISCode      *string      `json:"qris_code" db:"qris_code"`
	QRISExpiry    *time.Time   `json:"qris_expiry" db:"qris_expiry"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	CompletedAt   *time.Time   `json:"completed_at" db:"completed_at"`
	
	// Joined fields
	Items []OrderItem `json:"items,omitempty"`
}

// OrderItem represents individual items in an order
type OrderItem struct {
	ID        int    `json:"id" db:"id"`
	OrderID   string `json:"order_id" db:"order_id"`
	ProductID int    `json:"product_id" db:"product_id"`
	Quantity  int    `json:"quantity" db:"quantity"`
	Price     int    `json:"price" db:"price"` // Price at time of purchase
	
	// Joined fields from Product
	ProductName        string  `json:"product_name,omitempty" db:"product_name"`
	ProductDescription string  `json:"product_description,omitempty" db:"product_description"`
	ProductDownloadURL *string `json:"product_download_url,omitempty" db:"product_download_url"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusExpired   PaymentStatus = "expired"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Value implements the driver.Valuer interface for PaymentStatus
func (ps PaymentStatus) Value() (driver.Value, error) {
	return string(ps), nil
}

// Scan implements the sql.Scanner interface for PaymentStatus
func (ps *PaymentStatus) Scan(value interface{}) error {
	if value == nil {
		*ps = PaymentStatusPending
		return nil
	}
	if str, ok := value.(string); ok {
		*ps = PaymentStatus(str)
		return nil
	}
	return nil
}

// QRISPayment represents QRIS payment details
type QRISPayment struct {
	OrderID       string    `json:"order_id"`
	Amount        int       `json:"amount"`
	MerchantID    string    `json:"merchant_id"`
	MerchantName  string    `json:"merchant_name"`
	City          string    `json:"city"`
	CountryCode   string    `json:"country_code"`
	CurrencyCode  string    `json:"currency_code"`
	QRString      string    `json:"qr_string"`
	ExpiryTime    time.Time `json:"expiry_time"`
}

// ProductCategory represents product categories
type ProductCategory struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Count       int    `json:"count,omitempty"`
}

// GetDefaultCategories returns predefined product categories
func GetDefaultCategories() []ProductCategory {
	return []ProductCategory{
		{Name: "music", DisplayName: "🎵 Musik & Audio", Icon: "🎵"},
		{Name: "entertainment", DisplayName: "🎬 Hiburan", Icon: "🎬"},
		{Name: "design", DisplayName: "🎨 Design & Kreativitas", Icon: "🎨"},
		{Name: "productivity", DisplayName: "💼 Produktivitas", Icon: "💼"},
		{Name: "education", DisplayName: "📚 Edukasi", Icon: "📚"},
		{Name: "gaming", DisplayName: "🎮 Gaming", Icon: "🎮"},
		{Name: "social", DisplayName: "💬 Sosial Media", Icon: "💬"},
		{Name: "utility", DisplayName: "🔧 Utilitas", Icon: "🔧"},
	}
}

// CartSummary represents a summary of cart contents
type CartSummary struct {
	TotalItems int `json:"total_items"`
	TotalPrice int `json:"total_price"`
	Items      []CartItem `json:"items"`
}

// OrderSummary represents order statistics
type OrderSummary struct {
	TotalOrders    int `json:"total_orders"`
	PendingOrders  int `json:"pending_orders"`
	CompletedOrders int `json:"completed_orders"`
	TotalRevenue   int `json:"total_revenue"`
	TodayOrders    int `json:"today_orders"`
	TodayRevenue   int `json:"today_revenue"`
}

// UserStats represents user statistics
type UserStats struct {
	TotalUsers   int `json:"total_users"`
	ActiveUsers  int `json:"active_users"`
	NewToday     int `json:"new_today"`
	TopBuyers    []User `json:"top_buyers,omitempty"`
}

// ProductStats represents product statistics  
type ProductStats struct {
	TotalProducts    int `json:"total_products"`
	ActiveProducts   int `json:"active_products"`
	OutOfStock      int `json:"out_of_stock"`
	TopSelling      []Product `json:"top_selling,omitempty"`
}

// FormatPrice formats price with currency symbol
func FormatPrice(price int, symbol string) string {
	return fmt.Sprintf("%s %s", symbol, formatNumber(price))
}

// formatNumber formats number with thousand separators
func formatNumber(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}
	
	var result []rune
	for i, r := range []rune(str) {
		if i > 0 && (len(str)-i)%3 == 0 {
			result = append(result, '.')
		}
		result = append(result, r)
	}
	return string(result)
}

// ProductContentType represents different types of product content
type ProductContentType string

const (
	ContentTypeAccount ProductContentType = "account" // email | password
	ContentTypeLink    ProductContentType = "link"    // URL/link
	ContentTypeCode    ProductContentType = "code"    // redeem code/voucher
	ContentTypeCustom  ProductContentType = "custom"  // custom text format
)

// ProductAccount represents different product delivery formats (account/link/code/custom)
type ProductAccount struct {
	ID          int                `json:"id" db:"id"`
	ProductID   int                `json:"product_id" db:"product_id"`
	ContentType ProductContentType `json:"content_type" db:"content_type"`
	ContentData string             `json:"content_data" db:"content_data"`
	// Legacy fields for backward compatibility (deprecated, use ContentData instead)
	Email        *string    `json:"email,omitempty" db:"email"`
	Password     *string    `json:"password,omitempty" db:"password"`
	IsSold       bool       `json:"is_sold" db:"is_sold"`
	SoldToUserID *int64     `json:"sold_to_user_id" db:"sold_to_user_id"`
	SoldOrderID  *string    `json:"sold_order_id" db:"sold_order_id"`
	SoldAt       *time.Time `json:"sold_at" db:"sold_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// SoldAccount represents sold account tracking (supports multiple content formats)
type SoldAccount struct {
	ID          int                `json:"id" db:"id"`
	OrderID     string             `json:"order_id" db:"order_id"`
	ProductID   int                `json:"product_id" db:"product_id"`
	AccountID   int                `json:"account_id" db:"account_id"`
	UserID      int64              `json:"user_id" db:"user_id"`
	ContentType ProductContentType `json:"content_type" db:"content_type"`
	ContentData string             `json:"content_data" db:"content_data"`
	// Legacy fields for backward compatibility
	Email       *string    `json:"email,omitempty" db:"email"`
	Password    *string    `json:"password,omitempty" db:"password"`
	SoldPrice   int        `json:"sold_price" db:"sold_price"`
	SoldAt      time.Time  `json:"sold_at" db:"sold_at"`
	
	// Joined fields
	ProductName    string  `json:"product_name,omitempty" db:"product_name"`
	BuyerFirstName *string `json:"buyer_first_name,omitempty" db:"first_name"`
	BuyerLastName  *string `json:"buyer_last_name,omitempty" db:"last_name"`
	BuyerUsername  *string `json:"buyer_username,omitempty" db:"username"`
}

// PaymentVerification represents payment verification for anti-manipulation
type PaymentVerification struct {
	ID               int       `json:"id" db:"id"`
	OrderID          string    `json:"order_id" db:"order_id"`
	ExpectedAmount   int       `json:"expected_amount" db:"expected_amount"`
	QRISPayload      string    `json:"qris_payload" db:"qris_payload"`
	VerificationHash string    `json:"verification_hash" db:"verification_hash"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	VerifiedAt       *time.Time `json:"verified_at" db:"verified_at"`
}

// StockSummary represents stock summary for a product
type StockSummary struct {
	ProductID      int `json:"product_id"`
	AvailableStock int `json:"available_stock"`
	SoldStock      int `json:"sold_stock"`
	TotalStock     int `json:"total_stock"`
}

// AccountCredentials represents formatted account credentials
type AccountCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Format   string `json:"format"` // "email | password"
}

// FormatContent formats product content based on type
func (a *ProductAccount) FormatContent() string {
	// If ContentData is set, use it
	if a.ContentData != "" {
		return a.ContentData
	}
	// Fallback to legacy format for backward compatibility
	if a.Email != nil && a.Password != nil {
		return fmt.Sprintf("%s | %s", *a.Email, *a.Password)
	}
	return ""
}

// GetContentLabel returns a label for the content type
func (a *ProductAccount) GetContentLabel() string {
	switch a.ContentType {
	case ContentTypeAccount:
		return "🔐 Akun"
	case ContentTypeLink:
		return "🔗 Link"
	case ContentTypeCode:
		return "🎫 Kode"
	case ContentTypeCustom:
		return "📝 Data"
	default:
		return "📦 Produk"
	}
}

// FormatAccountCredentials formats account as "email | password" (deprecated)
func (a *ProductAccount) FormatAccountCredentials() string {
	return a.FormatContent()
}

// FormatContent formats sold account content based on type
func (s *SoldAccount) FormatContent() string {
	// If ContentData is set, use it
	if s.ContentData != "" {
		return s.ContentData
	}
	// Fallback to legacy format for backward compatibility
	if s.Email != nil && s.Password != nil {
		return fmt.Sprintf("%s | %s", *s.Email, *s.Password)
	}
	return ""
}

// GetContentLabel returns a label for the content type
func (s *SoldAccount) GetContentLabel() string {
	switch s.ContentType {
	case ContentTypeAccount:
		return "🔐 Akun"
	case ContentTypeLink:
		return "🔗 Link"
	case ContentTypeCode:
		return "🎫 Kode"
	case ContentTypeCustom:
		return "📝 Data"
	default:
		return "📦 Produk"
	}
}

// GetBuyerName returns formatted buyer name
func (s *SoldAccount) GetBuyerName() string {
	var name string
	if s.BuyerFirstName != nil {
		name = *s.BuyerFirstName
	}
	if s.BuyerLastName != nil {
		if name != "" {
			name += " " + *s.BuyerLastName
		} else {
			name = *s.BuyerLastName
		}
	}
	if name == "" {
		name = "Unknown"
	}
	return name
}

// JSONMap is a helper type for storing JSON data in database
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}
	
	return json.Unmarshal(bytes, j)
}