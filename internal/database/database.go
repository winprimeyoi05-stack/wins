package database

import (
	"database/sql"
	"fmt"
	"time"

	"telegram-premium-store/internal/models"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// Initialize creates and initializes the database
func Initialize(dbPath string) (*DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbWrapper := &DB{db}

	// Run migrations
	if err := dbWrapper.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Insert sample data
	if err := dbWrapper.insertSampleData(); err != nil {
		logrus.Warn("Failed to insert sample data: ", err)
	}

	// Insert default categories
	if err := dbWrapper.insertDefaultCategories(); err != nil {
		logrus.Warn("Failed to insert default categories: ", err)
	}

	logrus.Info("✅ Database initialized successfully")
	return dbWrapper, nil
}

// migrate runs database migrations
func (db *DB) migrate() error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			username TEXT,
			first_name TEXT,
			last_name TEXT,
			join_date DATETIME DEFAULT CURRENT_TIMESTAMP,
			is_admin BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE
		)`,

		// Products table
		`CREATE TABLE IF NOT EXISTS products (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			price INTEGER NOT NULL,
			category TEXT NOT NULL,
			image_url TEXT,
			download_url TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			stock INTEGER DEFAULT 999,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Cart table
		`CREATE TABLE IF NOT EXISTS cart (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER DEFAULT 1,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
			UNIQUE(user_id, product_id)
		)`,

		// Orders table
		`CREATE TABLE IF NOT EXISTS orders (
			id TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			total_amount INTEGER NOT NULL,
			payment_method TEXT DEFAULT 'qris',
			payment_status TEXT DEFAULT 'pending',
			qris_code TEXT,
			qris_expiry DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			completed_at DATETIME,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
		)`,

		// Order items table
		`CREATE TABLE IF NOT EXISTS order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			price INTEGER NOT NULL,
			FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE
		)`,

		// Categories table
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			display_name TEXT NOT NULL,
			icon TEXT DEFAULT '📦',
			is_active BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// User interactions table for broadcast
		`CREATE TABLE IF NOT EXISTS user_interactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			interaction_type TEXT NOT NULL,
			interaction_data TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
		)`,

		// Broadcast messages table
		`CREATE TABLE IF NOT EXISTS broadcasts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			target_type TEXT DEFAULT 'all',
			sent_count INTEGER DEFAULT 0,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			sent_at DATETIME
		)`,

		// Product accounts table (email|password stock)
		`CREATE TABLE IF NOT EXISTS product_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			product_id INTEGER NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			is_sold BOOLEAN DEFAULT FALSE,
			sold_to_user_id INTEGER,
			sold_order_id TEXT,
			sold_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
			FOREIGN KEY (sold_to_user_id) REFERENCES users (user_id) ON DELETE SET NULL
		)`,

		// Payment verifications table (anti-manipulation)
		`CREATE TABLE IF NOT EXISTS payment_verifications (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT NOT NULL,
			expected_amount INTEGER NOT NULL,
			qris_payload TEXT NOT NULL,
			verification_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			verified_at DATETIME,
			FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE
		)`,

		// Sold accounts tracking
		`CREATE TABLE IF NOT EXISTS sold_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_id TEXT NOT NULL,
			product_id INTEGER NOT NULL,
			account_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			sold_price INTEGER NOT NULL,
			sold_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (order_id) REFERENCES orders (id) ON DELETE CASCADE,
			FOREIGN KEY (product_id) REFERENCES products (id) ON DELETE CASCADE,
			FOREIGN KEY (account_id) REFERENCES product_accounts (id) ON DELETE CASCADE,
			FOREIGN KEY (user_id) REFERENCES users (user_id) ON DELETE CASCADE
		)`,

		// Indexes for better performance
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_products_category ON products(category)`,
		`CREATE INDEX IF NOT EXISTS idx_products_active ON products(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_cart_user ON cart(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_user ON orders(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(payment_status)`,
		`CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_categories_active ON categories(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_user_interactions_user ON user_interactions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_user_interactions_type ON user_interactions(interaction_type)`,
		`CREATE INDEX IF NOT EXISTS idx_broadcasts_created ON broadcasts(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_product_accounts_product ON product_accounts(product_id)`,
		`CREATE INDEX IF NOT EXISTS idx_product_accounts_sold ON product_accounts(is_sold)`,
		`CREATE INDEX IF NOT EXISTS idx_payment_verifications_order ON payment_verifications(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sold_accounts_order ON sold_accounts(order_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sold_accounts_product ON sold_accounts(product_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// insertSampleData inserts sample products if the database is empty
func (db *DB) insertSampleData() error {
	// Check if products already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM products").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Data already exists
	}

	sampleProducts := []models.Product{
		{
			Name:        "Spotify Premium 1 Bulan",
			Description: "Akses unlimited musik tanpa iklan, download offline, kualitas audio terbaik. Nikmati jutaan lagu dari seluruh dunia dengan kualitas tinggi.",
			Price:       25000,
			Category:    "music",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/1DB954/FFFFFF?text=Spotify+Premium"),
			DownloadURL: stringPtr("https://spotify.com/premium"),
			Stock:       100,
		},
		{
			Name:        "Netflix Premium 1 Bulan",
			Description: "Streaming film dan series unlimited, 4K Ultra HD, 4 device bersamaan. Akses ke ribuan konten premium dari seluruh dunia.",
			Price:       65000,
			Category:    "entertainment",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/E50914/FFFFFF?text=Netflix+Premium"),
			DownloadURL: stringPtr("https://netflix.com"),
			Stock:       50,
		},
		{
			Name:        "YouTube Premium 1 Bulan",
			Description: "Tanpa iklan, background play, YouTube Music included. Download video untuk ditonton offline kapan saja.",
			Price:       35000,
			Category:    "entertainment",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/FF0000/FFFFFF?text=YouTube+Premium"),
			DownloadURL: stringPtr("https://youtube.com/premium"),
			Stock:       75,
		},
		{
			Name:        "Canva Pro 1 Bulan",
			Description: "Design tool premium dengan template unlimited dan fitur advanced. Buat desain profesional dengan mudah.",
			Price:       45000,
			Category:    "design",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/00C4CC/FFFFFF?text=Canva+Pro"),
			DownloadURL: stringPtr("https://canva.com/pro"),
			Stock:       80,
		},
		{
			Name:        "Adobe Creative Cloud",
			Description: "Photoshop, Illustrator, Premiere Pro, After Effects - Full Package. Suite lengkap untuk kreativitas tanpa batas.",
			Price:       150000,
			Category:    "design",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/FF0000/FFFFFF?text=Adobe+CC"),
			DownloadURL: stringPtr("https://adobe.com/creativecloud"),
			Stock:       30,
		},
		{
			Name:        "Microsoft Office 365",
			Description: "Word, Excel, PowerPoint, Outlook dan aplikasi produktivitas lainnya. Akses cloud storage 1TB OneDrive.",
			Price:       85000,
			Category:    "productivity",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/0078D4/FFFFFF?text=Office+365"),
			DownloadURL: stringPtr("https://office.com"),
			Stock:       60,
		},
		{
			Name:        "Duolingo Plus 1 Bulan",
			Description: "Belajar bahasa asing tanpa iklan, download pelajaran offline, unlimited hearts dan progress tracking.",
			Price:       30000,
			Category:    "education",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/58CC02/FFFFFF?text=Duolingo+Plus"),
			DownloadURL: stringPtr("https://duolingo.com/plus"),
			Stock:       90,
		},
		{
			Name:        "Discord Nitro 1 Bulan",
			Description: "Fitur premium Discord dengan emoji custom, file upload 100MB, server boost, dan kualitas streaming HD.",
			Price:       50000,
			Category:    "gaming",
			ImageURL:    stringPtr("https://via.placeholder.com/300x200/5865F2/FFFFFF?text=Discord+Nitro"),
			DownloadURL: stringPtr("https://discord.com/nitro"),
			Stock:       70,
		},
	}

	for _, product := range sampleProducts {
		_, err := db.Exec(`
			INSERT INTO products (name, description, price, category, image_url, download_url, stock)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, product.Name, product.Description, product.Price, product.Category,
			product.ImageURL, product.DownloadURL, product.Stock)

		if err != nil {
			return fmt.Errorf("failed to insert sample product %s: %w", product.Name, err)
		}
	}

	// Insert sample accounts for products
	if err := db.insertSampleAccounts(); err != nil {
		return fmt.Errorf("failed to insert sample accounts: %w", err)
	}

	logrus.Info("✅ Sample products and accounts inserted successfully")
	return nil
}

// insertSampleAccounts inserts sample email|password accounts for products
func (db *DB) insertSampleAccounts() error {
	// Get all products to add accounts
	products, err := db.GetProducts("", 100, 0)
	if err != nil {
		return err
	}

	sampleAccounts := map[string][]struct {
		Email    string
		Password string
	}{
		"Spotify Premium 1 Bulan": {
			{"spotify.premium1@gmail.com", "SpotifyPass123!"},
			{"premium.spotify2@gmail.com", "MusicLover456@"},
			{"spotify.user3@gmail.com", "Premium789#"},
			{"music.premium4@gmail.com", "SpotifyVIP999$"},
			{"premium.music5@gmail.com", "Spotify2024!@"},
		},
		"Netflix Premium 1 Bulan": {
			{"netflix.premium1@gmail.com", "NetflixHD123!"},
			{"premium.netflix2@gmail.com", "StreamVIP456@"},
			{"netflix.user3@gmail.com", "Premium789#"},
			{"stream.premium4@gmail.com", "NetflixUHD999$"},
			{"premium.stream5@gmail.com", "Netflix2024!@"},
		},
		"YouTube Premium 1 Bulan": {
			{"youtube.premium1@gmail.com", "YTPremium123!"},
			{"premium.youtube2@gmail.com", "VideoVIP456@"},
			{"youtube.user3@gmail.com", "Premium789#"},
			{"video.premium4@gmail.com", "YouTubeVIP999$"},
			{"premium.video5@gmail.com", "YouTube2024!@"},
		},
		"Canva Pro 1 Bulan": {
			{"canva.pro1@gmail.com", "CanvaPro123!"},
			{"pro.canva2@gmail.com", "DesignVIP456@"},
			{"canva.user3@gmail.com", "Premium789#"},
			{"design.pro4@gmail.com", "CanvaVIP999$"},
			{"pro.design5@gmail.com", "Canva2024!@"},
		},
		"Adobe Creative Cloud": {
			{"adobe.cc1@gmail.com", "AdobeCC123!"},
			{"creative.adobe2@gmail.com", "DesignSuite456@"},
			{"adobe.user3@gmail.com", "Creative789#"},
			{"cc.premium4@gmail.com", "AdobeVIP999$"},
			{"premium.adobe5@gmail.com", "Adobe2024!@"},
		},
	}

	for _, product := range products {
		if accounts, exists := sampleAccounts[product.Name]; exists {
			for _, account := range accounts {
				_, err := db.Exec(`
					INSERT INTO product_accounts (product_id, email, password)
					VALUES (?, ?, ?)
				`, product.ID, account.Email, account.Password)
				if err != nil {
					return fmt.Errorf("failed to insert account for product %s: %w", product.Name, err)
				}
			}
		}
	}

	return nil
}

// insertDefaultCategories inserts default categories if not exist
func (db *DB) insertDefaultCategories() error {
	// Check if categories already exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // Categories already exist
	}

	defaultCategories := models.GetDefaultCategories()
	for _, category := range defaultCategories {
		_, err := db.Exec(`
			INSERT INTO categories (name, display_name, icon)
			VALUES (?, ?, ?)
		`, category.Name, category.DisplayName, category.Icon)
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %w", category.Name, err)
		}
	}

	logrus.Info("✅ Default categories inserted successfully")
	return nil
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// User operations
func (db *DB) CreateUser(user *models.User) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO users (user_id, username, first_name, last_name, is_admin)
		VALUES (?, ?, ?, ?, ?)
	`, user.UserID, user.Username, user.FirstName, user.LastName, user.IsAdmin)
	return err
}

func (db *DB) GetUser(userID int64) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(`
		SELECT user_id, username, first_name, last_name, join_date, is_admin, is_active
		FROM users WHERE user_id = ?
	`, userID).Scan(&user.UserID, &user.Username, &user.FirstName, &user.LastName,
		&user.JoinDate, &user.IsAdmin, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// Product operations
func (db *DB) GetProducts(category string, limit, offset int) ([]models.Product, error) {
	var query string
	var args []interface{}

	if category != "" {
		query = `
			SELECT id, name, description, price, category, image_url, download_url, 
				   is_active, stock, created_at, updated_at
			FROM products 
			WHERE is_active = TRUE AND category = ?
			ORDER BY name
			LIMIT ? OFFSET ?
		`
		args = []interface{}{category, limit, offset}
	} else {
		query = `
			SELECT id, name, description, price, category, image_url, download_url,
				   is_active, stock, created_at, updated_at
			FROM products 
			WHERE is_active = TRUE
			ORDER BY category, name
			LIMIT ? OFFSET ?
		`
		args = []interface{}{limit, offset}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.Price, &product.Category, &product.ImageURL,
			&product.DownloadURL, &product.IsActive, &product.Stock,
			&product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

func (db *DB) GetProduct(id int) (*models.Product, error) {
	product := &models.Product{}
	err := db.QueryRow(`
		SELECT id, name, description, price, category, image_url, download_url,
			   is_active, stock, created_at, updated_at
		FROM products WHERE id = ? AND is_active = TRUE
	`, id).Scan(&product.ID, &product.Name, &product.Description,
		&product.Price, &product.Category, &product.ImageURL,
		&product.DownloadURL, &product.IsActive, &product.Stock,
		&product.CreatedAt, &product.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return product, err
}

func (db *DB) GetCategories() ([]models.ProductCategory, error) {
	rows, err := db.Query(`
		SELECT category, COUNT(*) as count
		FROM products 
		WHERE is_active = TRUE
		GROUP BY category
		ORDER BY category
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categoryMap := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, err
		}
		categoryMap[category] = count
	}

	// Get default categories with counts
	categories := models.GetDefaultCategories()
	for i := range categories {
		if count, exists := categoryMap[categories[i].Name]; exists {
			categories[i].Count = count
		}
	}

	return categories, nil
}

// Cart operations
func (db *DB) AddToCart(userID int64, productID, quantity int) error {
	_, err := db.Exec(`
		INSERT INTO cart (user_id, product_id, quantity)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, product_id) 
		DO UPDATE SET quantity = quantity + ?, added_at = CURRENT_TIMESTAMP
	`, userID, productID, quantity, quantity)
	return err
}

func (db *DB) GetCart(userID int64) ([]models.CartItem, error) {
	rows, err := db.Query(`
		SELECT c.id, c.user_id, c.product_id, c.quantity, c.added_at,
			   p.name, p.price, p.image_url
		FROM cart c
		JOIN products p ON c.product_id = p.id
		WHERE c.user_id = ? AND p.is_active = TRUE
		ORDER BY c.added_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.CartItem
	for rows.Next() {
		var item models.CartItem
		err := rows.Scan(&item.ID, &item.UserID, &item.ProductID,
			&item.Quantity, &item.AddedAt, &item.ProductName,
			&item.ProductPrice, &item.ProductImage)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (db *DB) RemoveFromCart(userID int64, productID int) error {
	_, err := db.Exec(`
		DELETE FROM cart WHERE user_id = ? AND product_id = ?
	`, userID, productID)
	return err
}

func (db *DB) ClearCart(userID int64) error {
	_, err := db.Exec(`DELETE FROM cart WHERE user_id = ?`, userID)
	return err
}

// Order operations
func (db *DB) CreateOrder(order *models.Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (id, user_id, total_amount, payment_method, payment_status, qris_code, qris_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, order.ID, order.UserID, order.TotalAmount, order.PaymentMethod,
		order.PaymentStatus, order.QRISCode, order.QRISExpiry)
	if err != nil {
		return err
	}

	// Insert order items
	for _, item := range order.Items {
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES (?, ?, ?, ?)
		`, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetOrder(orderID string) (*models.Order, error) {
	order := &models.Order{}
	err := db.QueryRow(`
		SELECT id, user_id, total_amount, payment_method, payment_status,
			   qris_code, qris_expiry, created_at, updated_at, completed_at
		FROM orders WHERE id = ?
	`, orderID).Scan(&order.ID, &order.UserID, &order.TotalAmount,
		&order.PaymentMethod, &order.PaymentStatus, &order.QRISCode,
		&order.QRISExpiry, &order.CreatedAt, &order.UpdatedAt, &order.CompletedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Get order items
	items, err := db.getOrderItems(orderID)
	if err != nil {
		return nil, err
	}
	order.Items = items

	return order, nil
}

func (db *DB) getOrderItems(orderID string) ([]models.OrderItem, error) {
	rows, err := db.Query(`
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price,
			   p.name, p.description, p.download_url
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = ?
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.OrderItem
	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID,
			&item.Quantity, &item.Price, &item.ProductName,
			&item.ProductDescription, &item.ProductDownloadURL)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (db *DB) UpdateOrderStatus(orderID string, status models.PaymentStatus) error {
	now := time.Now()
	var completedAt *time.Time
	if status == models.PaymentStatusPaid {
		completedAt = &now
	}

	_, err := db.Exec(`
		UPDATE orders 
		SET payment_status = ?, updated_at = ?, completed_at = ?
		WHERE id = ?
	`, status, now, completedAt, orderID)
	return err
}

func (db *DB) GetUserOrders(userID int64, limit, offset int) ([]models.Order, error) {
	rows, err := db.Query(`
		SELECT id, user_id, total_amount, payment_method, payment_status,
			   qris_code, qris_expiry, created_at, updated_at, completed_at
		FROM orders 
		WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.TotalAmount,
			&order.PaymentMethod, &order.PaymentStatus, &order.QRISCode,
			&order.QRISExpiry, &order.CreatedAt, &order.UpdatedAt, &order.CompletedAt)
		if err != nil {
			return nil, err
		}

		// Get order items for each order
		items, err := db.getOrderItems(order.ID)
		if err != nil {
			return nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	return orders, rows.Err()
}

// Stock Management Methods

// UpdateProductStock updates product stock
func (db *DB) UpdateProductStock(productID, newStock int) error {
	_, err := db.Exec(`
		UPDATE products SET stock = ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ?
	`, newStock, productID)
	return err
}

// DecrementProductStock decrements product stock (for purchases)
func (db *DB) DecrementProductStock(productID, quantity int) error {
	_, err := db.Exec(`
		UPDATE products 
		SET stock = stock - ?, updated_at = CURRENT_TIMESTAMP 
		WHERE id = ? AND stock >= ?
	`, quantity, productID, quantity)
	return err
}

// GetLowStockProducts returns products with stock below threshold
func (db *DB) GetLowStockProducts(threshold int) ([]models.Product, error) {
	rows, err := db.Query(`
		SELECT id, name, description, price, category, image_url, download_url,
			   is_active, stock, created_at, updated_at
		FROM products 
		WHERE is_active = TRUE AND stock <= ?
		ORDER BY stock ASC, name
	`, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(&product.ID, &product.Name, &product.Description,
			&product.Price, &product.Category, &product.ImageURL,
			&product.DownloadURL, &product.IsActive, &product.Stock,
			&product.CreatedAt, &product.UpdatedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	return products, rows.Err()
}

// Category Management Methods

// CreateCategory creates a new category
func (db *DB) CreateCategory(name, displayName, icon string) error {
	_, err := db.Exec(`
		INSERT INTO categories (name, display_name, icon)
		VALUES (?, ?, ?)
	`, name, displayName, icon)
	return err
}

// UpdateCategory updates category information
func (db *DB) UpdateCategory(id int, displayName, icon string) error {
	_, err := db.Exec(`
		UPDATE categories 
		SET display_name = ?, icon = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, displayName, icon, id)
	return err
}

// DeleteCategory soft deletes a category
func (db *DB) DeleteCategory(id int) error {
	_, err := db.Exec(`
		UPDATE categories SET is_active = FALSE WHERE id = ?
	`, id)
	return err
}

// GetCategoriesFromDB returns categories from database
func (db *DB) GetCategoriesFromDB() ([]models.ProductCategory, error) {
	rows, err := db.Query(`
		SELECT c.id, c.name, c.display_name, c.icon, COUNT(p.id) as count
		FROM categories c
		LEFT JOIN products p ON c.name = p.category AND p.is_active = TRUE
		WHERE c.is_active = TRUE
		GROUP BY c.id, c.name, c.display_name, c.icon
		ORDER BY c.display_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ProductCategory
	for rows.Next() {
		var category models.ProductCategory
		var id int
		err := rows.Scan(&id, &category.Name, &category.DisplayName, 
			&category.Icon, &category.Count)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// Product Management Methods

// DeleteProduct soft deletes a product
func (db *DB) DeleteProduct(id int) error {
	_, err := db.Exec(`
		UPDATE products SET is_active = FALSE WHERE id = ?
	`, id)
	return err
}

// UpdateProduct updates product information
func (db *DB) UpdateProduct(product *models.Product) error {
	_, err := db.Exec(`
		UPDATE products 
		SET name = ?, description = ?, price = ?, category = ?, 
			image_url = ?, download_url = ?, stock = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, product.Name, product.Description, product.Price, product.Category,
		product.ImageURL, product.DownloadURL, product.Stock, product.ID)
	return err
}

// User Interaction Tracking

// LogUserInteraction logs user interaction for broadcast targeting
func (db *DB) LogUserInteraction(userID int64, interactionType, data string) error {
	_, err := db.Exec(`
		INSERT INTO user_interactions (user_id, interaction_type, interaction_data)
		VALUES (?, ?, ?)
	`, userID, interactionType, data)
	return err
}

// GetActiveUsers returns list of active users for broadcast
func (db *DB) GetActiveUsers(days int) ([]int64, error) {
	rows, err := db.Query(`
		SELECT DISTINCT user_id 
		FROM user_interactions 
		WHERE created_at >= datetime('now', '-' || ? || ' days')
		ORDER BY user_id
	`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// GetAllUsers returns all users who ever interacted with bot
func (db *DB) GetAllUsers() ([]int64, error) {
	rows, err := db.Query(`
		SELECT DISTINCT user_id FROM users ORDER BY user_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var userIDs []int64
	for rows.Next() {
		var userID int64
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, rows.Err()
}

// Broadcast Management

// CreateBroadcast creates a new broadcast message
func (db *DB) CreateBroadcast(message, targetType string, createdBy int64) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO broadcasts (message, target_type, created_by)
		VALUES (?, ?, ?)
	`, message, targetType, createdBy)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// UpdateBroadcastSent updates broadcast sent status
func (db *DB) UpdateBroadcastSent(id int64, sentCount int) error {
	_, err := db.Exec(`
		UPDATE broadcasts 
		SET sent_count = ?, sent_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, sentCount, id)
	return err
}

// Order Management with Stock

// CreateOrderWithStock creates order and decrements stock
func (db *DB) CreateOrderWithStock(order *models.Order) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check stock availability for all items
	for _, item := range order.Items {
		var currentStock int
		err = tx.QueryRow(`
			SELECT stock FROM products WHERE id = ? AND is_active = TRUE
		`, item.ProductID).Scan(&currentStock)
		if err != nil {
			return fmt.Errorf("product not found or inactive: %d", item.ProductID)
		}

		if currentStock < item.Quantity {
			return fmt.Errorf("insufficient stock for product ID %d: available %d, requested %d", 
				item.ProductID, currentStock, item.Quantity)
		}
	}

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (id, user_id, total_amount, payment_method, payment_status, qris_code, qris_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, order.ID, order.UserID, order.TotalAmount, order.PaymentMethod,
		order.PaymentStatus, order.QRISCode, order.QRISExpiry)
	if err != nil {
		return err
	}

	// Insert order items and decrement stock
	for _, item := range order.Items {
		// Insert order item
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES (?, ?, ?, ?)
		`, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return err
		}

		// Decrement stock
		_, err = tx.Exec(`
			UPDATE products 
			SET stock = stock - ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, item.Quantity, item.ProductID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// RestoreStockFromOrder restores stock when order is cancelled
func (db *DB) RestoreStockFromOrder(orderID string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get order items
	rows, err := tx.Query(`
		SELECT product_id, quantity FROM order_items WHERE order_id = ?
	`, orderID)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Restore stock for each item
	for rows.Next() {
		var productID, quantity int
		if err := rows.Scan(&productID, &quantity); err != nil {
			return err
		}

		_, err = tx.Exec(`
			UPDATE products 
			SET stock = stock + ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, quantity, productID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}