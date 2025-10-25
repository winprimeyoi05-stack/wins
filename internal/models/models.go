package models

import (
	"database/sql/driver"
	"encoding/json"
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

import "fmt"

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