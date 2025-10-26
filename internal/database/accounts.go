package database

import (
	"fmt"

	"telegram-premium-store/internal/models"
)

// Product Account Management

// GetAvailableAccountCount returns count of available accounts for a product
func (db *DB) GetAvailableAccountCount(productID int) (int, error) {
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM product_accounts 
		WHERE product_id = ? AND is_sold = FALSE
	`, productID).Scan(&count)
	return count, err
}

// GetAvailableAccounts returns available accounts for a product
func (db *DB) GetAvailableAccounts(productID int) ([]models.ProductAccount, error) {
	rows, err := db.Query(`
		SELECT id, product_id, email, password, created_at
		FROM product_accounts 
		WHERE product_id = ? AND is_sold = FALSE
		ORDER BY created_at ASC
	`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.ProductAccount
	for rows.Next() {
		var account models.ProductAccount
		err := rows.Scan(&account.ID, &account.ProductID, &account.Email, 
			&account.Password, &account.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// CreateOrderWithAccounts creates order and assigns accounts to buyer
func (db *DB) CreateOrderWithAccounts(order *models.Order) ([]models.ProductAccount, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var assignedAccounts []models.ProductAccount

	// Check account availability for all items
	for _, item := range order.Items {
		var availableAccounts int
		err = tx.QueryRow(`
			SELECT COUNT(*) FROM product_accounts 
			WHERE product_id = ? AND is_sold = FALSE
		`, item.ProductID).Scan(&availableAccounts)
		if err != nil {
			return nil, err
		}

		if availableAccounts < item.Quantity {
			return nil, fmt.Errorf("insufficient accounts for product ID %d: available %d, requested %d", 
				item.ProductID, availableAccounts, item.Quantity)
		}
	}

	// Insert order
	_, err = tx.Exec(`
		INSERT INTO orders (id, user_id, total_amount, payment_method, payment_status, qris_code, qris_expiry)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, order.ID, order.UserID, order.TotalAmount, order.PaymentMethod,
		order.PaymentStatus, order.QRISCode, order.QRISExpiry)
	if err != nil {
		return nil, err
	}

	// Insert order items and assign accounts
	for _, item := range order.Items {
		// Insert order item
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES (?, ?, ?, ?)
		`, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return nil, err
		}

		// Get and assign accounts
		rows, err := tx.Query(`
			SELECT id, email, password FROM product_accounts 
			WHERE product_id = ? AND is_sold = FALSE
			ORDER BY created_at ASC
			LIMIT ?
		`, item.ProductID, item.Quantity)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var account models.ProductAccount
			err := rows.Scan(&account.ID, &account.Email, &account.Password)
			if err != nil {
				rows.Close()
				return nil, err
			}

			// Mark account as sold
			_, err = tx.Exec(`
				UPDATE product_accounts 
				SET is_sold = TRUE, sold_to_user_id = ?, sold_order_id = ?, sold_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, order.UserID, order.ID, account.ID)
			if err != nil {
				rows.Close()
				return nil, err
			}

			// Add to sold accounts tracking
			_, err = tx.Exec(`
				INSERT INTO sold_accounts (order_id, product_id, account_id, user_id, email, password, sold_price)
				VALUES (?, ?, ?, ?, ?, ?, ?)
			`, order.ID, item.ProductID, account.ID, order.UserID, account.Email, account.Password, item.Price)
			if err != nil {
				rows.Close()
				return nil, err
			}

			account.ProductID = item.ProductID
			assignedAccounts = append(assignedAccounts, account)
		}
		rows.Close()
	}

	return assignedAccounts, tx.Commit()
}

// GetProductAccountsForOrder returns accounts assigned to an order
func (db *DB) GetProductAccountsForOrder(orderID string) ([]models.SoldAccount, error) {
	rows, err := db.Query(`
		SELECT sa.id, sa.order_id, sa.product_id, sa.user_id, sa.email, sa.password, 
			   sa.sold_price, sa.sold_at, p.name as product_name,
			   u.first_name, u.last_name, u.username
		FROM sold_accounts sa
		JOIN products p ON sa.product_id = p.id
		LEFT JOIN users u ON sa.user_id = u.user_id
		WHERE sa.order_id = ?
		ORDER BY sa.sold_at ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.SoldAccount
	for rows.Next() {
		var account models.SoldAccount
		err := rows.Scan(&account.ID, &account.OrderID, &account.ProductID,
			&account.UserID, &account.Email, &account.Password, &account.SoldPrice,
			&account.SoldAt, &account.ProductName, &account.BuyerFirstName,
			&account.BuyerLastName, &account.BuyerUsername)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// GetSoldAccountsByProduct returns sold accounts for a specific product
func (db *DB) GetSoldAccountsByProduct(productID int, limit, offset int) ([]models.SoldAccount, error) {
	rows, err := db.Query(`
		SELECT sa.id, sa.order_id, sa.product_id, sa.user_id, sa.email, sa.password,
			   sa.sold_price, sa.sold_at, p.name as product_name,
			   u.first_name, u.last_name, u.username
		FROM sold_accounts sa
		JOIN products p ON sa.product_id = p.id
		LEFT JOIN users u ON sa.user_id = u.user_id
		WHERE sa.product_id = ?
		ORDER BY sa.sold_at DESC
		LIMIT ? OFFSET ?
	`, productID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.SoldAccount
	for rows.Next() {
		var account models.SoldAccount
		err := rows.Scan(&account.ID, &account.OrderID, &account.ProductID,
			&account.UserID, &account.Email, &account.Password, &account.SoldPrice,
			&account.SoldAt, &account.ProductName, &account.BuyerFirstName,
			&account.BuyerLastName, &account.BuyerUsername)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

// AddProductAccount adds new account to product stock
func (db *DB) AddProductAccount(productID int, email, password string) error {
	_, err := db.Exec(`
		INSERT INTO product_accounts (product_id, email, password)
		VALUES (?, ?, ?)
	`, productID, email, password)
	return err
}

// GetProductStockSummary returns stock summary including available and sold accounts
func (db *DB) GetProductStockSummary(productID int) (*models.StockSummary, error) {
	var summary models.StockSummary
	
	// Get available accounts count
	err := db.QueryRow(`
		SELECT COUNT(*) FROM product_accounts 
		WHERE product_id = ? AND is_sold = FALSE
	`, productID).Scan(&summary.AvailableStock)
	if err != nil {
		return nil, err
	}

	// Get sold accounts count
	err = db.QueryRow(`
		SELECT COUNT(*) FROM product_accounts 
		WHERE product_id = ? AND is_sold = TRUE
	`, productID).Scan(&summary.SoldStock)
	if err != nil {
		return nil, err
	}

	summary.TotalStock = summary.AvailableStock + summary.SoldStock
	summary.ProductID = productID

	return &summary, nil
}