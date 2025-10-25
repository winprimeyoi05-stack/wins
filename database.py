import sqlite3
import json
from datetime import datetime
from config import DATABASE_PATH

class Database:
    def __init__(self):
        self.init_database()
    
    def get_connection(self):
        return sqlite3.connect(DATABASE_PATH)
    
    def init_database(self):
        """Initialize database tables"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        # Users table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS users (
                user_id INTEGER PRIMARY KEY,
                username TEXT,
                first_name TEXT,
                last_name TEXT,
                join_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                is_admin BOOLEAN DEFAULT FALSE
            )
        ''')
        
        # Products table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS products (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                description TEXT,
                price INTEGER NOT NULL,
                category TEXT,
                image_url TEXT,
                download_link TEXT,
                is_active BOOLEAN DEFAULT TRUE,
                created_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        ''')
        
        # Orders table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS orders (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER,
                product_id INTEGER,
                quantity INTEGER DEFAULT 1,
                total_price INTEGER,
                payment_method TEXT,
                payment_status TEXT DEFAULT 'pending',
                order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users (user_id),
                FOREIGN KEY (product_id) REFERENCES products (id)
            )
        ''')
        
        # Cart table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS cart (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                user_id INTEGER,
                product_id INTEGER,
                quantity INTEGER DEFAULT 1,
                added_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users (user_id),
                FOREIGN KEY (product_id) REFERENCES products (id)
            )
        ''')
        
        conn.commit()
        conn.close()
        
        # Insert sample products
        self.insert_sample_products()
    
    def insert_sample_products(self):
        """Insert sample products for demo"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        # Check if products already exist
        cursor.execute("SELECT COUNT(*) FROM products")
        if cursor.fetchone()[0] > 0:
            conn.close()
            return
        
        sample_products = [
            {
                'name': 'Spotify Premium 1 Bulan',
                'description': 'Akses unlimited musik tanpa iklan, download offline, kualitas audio terbaik',
                'price': 25000,
                'category': 'Music',
                'image_url': 'https://via.placeholder.com/300x200?text=Spotify+Premium',
                'download_link': 'https://example.com/spotify'
            },
            {
                'name': 'Netflix Premium 1 Bulan',
                'description': 'Streaming film dan series unlimited, 4K Ultra HD, 4 device bersamaan',
                'price': 65000,
                'category': 'Entertainment',
                'image_url': 'https://via.placeholder.com/300x200?text=Netflix+Premium',
                'download_link': 'https://example.com/netflix'
            },
            {
                'name': 'YouTube Premium 1 Bulan',
                'description': 'Tanpa iklan, background play, YouTube Music included',
                'price': 35000,
                'category': 'Entertainment',
                'image_url': 'https://via.placeholder.com/300x200?text=YouTube+Premium',
                'download_link': 'https://example.com/youtube'
            },
            {
                'name': 'Canva Pro 1 Bulan',
                'description': 'Design tool premium dengan template unlimited dan fitur advanced',
                'price': 45000,
                'category': 'Design',
                'image_url': 'https://via.placeholder.com/300x200?text=Canva+Pro',
                'download_link': 'https://example.com/canva'
            },
            {
                'name': 'Adobe Creative Suite',
                'description': 'Photoshop, Illustrator, Premiere Pro, After Effects - Full Package',
                'price': 150000,
                'category': 'Design',
                'image_url': 'https://via.placeholder.com/300x200?text=Adobe+Suite',
                'download_link': 'https://example.com/adobe'
            }
        ]
        
        for product in sample_products:
            cursor.execute('''
                INSERT INTO products (name, description, price, category, image_url, download_link)
                VALUES (?, ?, ?, ?, ?, ?)
            ''', (product['name'], product['description'], product['price'], 
                  product['category'], product['image_url'], product['download_link']))
        
        conn.commit()
        conn.close()
    
    def add_user(self, user_id, username=None, first_name=None, last_name=None):
        """Add or update user"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            INSERT OR REPLACE INTO users (user_id, username, first_name, last_name)
            VALUES (?, ?, ?, ?)
        ''', (user_id, username, first_name, last_name))
        
        conn.commit()
        conn.close()
    
    def get_products(self, category=None):
        """Get all active products or by category"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        if category:
            cursor.execute('''
                SELECT * FROM products WHERE is_active = 1 AND category = ?
                ORDER BY name
            ''', (category,))
        else:
            cursor.execute('''
                SELECT * FROM products WHERE is_active = 1
                ORDER BY category, name
            ''')
        
        products = cursor.fetchall()
        conn.close()
        return products
    
    def get_product(self, product_id):
        """Get single product by ID"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('SELECT * FROM products WHERE id = ? AND is_active = 1', (product_id,))
        product = cursor.fetchone()
        conn.close()
        return product
    
    def add_to_cart(self, user_id, product_id, quantity=1):
        """Add product to user's cart"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        # Check if product already in cart
        cursor.execute('''
            SELECT id, quantity FROM cart WHERE user_id = ? AND product_id = ?
        ''', (user_id, product_id))
        
        existing = cursor.fetchone()
        
        if existing:
            # Update quantity
            new_quantity = existing[1] + quantity
            cursor.execute('''
                UPDATE cart SET quantity = ? WHERE id = ?
            ''', (new_quantity, existing[0]))
        else:
            # Add new item
            cursor.execute('''
                INSERT INTO cart (user_id, product_id, quantity)
                VALUES (?, ?, ?)
            ''', (user_id, product_id, quantity))
        
        conn.commit()
        conn.close()
    
    def get_cart(self, user_id):
        """Get user's cart with product details"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            SELECT c.id, c.quantity, p.name, p.price, p.id as product_id
            FROM cart c
            JOIN products p ON c.product_id = p.id
            WHERE c.user_id = ?
        ''', (user_id,))
        
        cart_items = cursor.fetchall()
        conn.close()
        return cart_items
    
    def clear_cart(self, user_id):
        """Clear user's cart"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('DELETE FROM cart WHERE user_id = ?', (user_id,))
        conn.commit()
        conn.close()
    
    def create_order(self, user_id, product_id, quantity, total_price, payment_method):
        """Create new order"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            INSERT INTO orders (user_id, product_id, quantity, total_price, payment_method)
            VALUES (?, ?, ?, ?, ?)
        ''', (user_id, product_id, quantity, total_price, payment_method))
        
        order_id = cursor.lastrowid
        conn.commit()
        conn.close()
        return order_id
    
    def get_user_orders(self, user_id):
        """Get user's order history"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            SELECT o.id, o.quantity, o.total_price, o.payment_method, 
                   o.payment_status, o.order_date, p.name
            FROM orders o
            JOIN products p ON o.product_id = p.id
            WHERE o.user_id = ?
            ORDER BY o.order_date DESC
        ''', (user_id,))
        
        orders = cursor.fetchall()
        conn.close()
        return orders
    
    def update_payment_status(self, order_id, status):
        """Update order payment status"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            UPDATE orders SET payment_status = ? WHERE id = ?
        ''', (status, order_id))
        
        conn.commit()
        conn.close()
    
    def get_categories(self):
        """Get all product categories"""
        conn = self.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            SELECT DISTINCT category FROM products WHERE is_active = 1
            ORDER BY category
        ''')
        
        categories = [row[0] for row in cursor.fetchall()]
        conn.close()
        return categories