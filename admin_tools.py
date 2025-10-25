#!/usr/bin/env python3
"""
Admin tools for managing the Telegram Premium App Sales Bot
"""

import sqlite3
from datetime import datetime
from database import Database
from config import DATABASE_PATH

class AdminTools:
    def __init__(self):
        self.db = Database()
    
    def add_product_cli(self):
        """Add product via command line interface"""
        print("📦 TAMBAH PRODUK BARU")
        print("=" * 50)
        
        name = input("Nama Produk: ")
        description = input("Deskripsi: ")
        
        while True:
            try:
                price = int(input("Harga (Rp): "))
                break
            except ValueError:
                print("❌ Harga harus berupa angka!")
        
        category = input("Kategori: ")
        image_url = input("URL Gambar (opsional): ") or None
        download_link = input("Link Download (opsional): ") or None
        
        # Add to database
        conn = self.db.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            INSERT INTO products (name, description, price, category, image_url, download_link)
            VALUES (?, ?, ?, ?, ?, ?)
        ''', (name, description, price, category, image_url, download_link))
        
        product_id = cursor.lastrowid
        conn.commit()
        conn.close()
        
        print(f"✅ Produk berhasil ditambahkan dengan ID: {product_id}")
    
    def list_products(self):
        """List all products"""
        products = self.db.get_products()
        
        print("📱 DAFTAR PRODUK")
        print("=" * 80)
        print(f"{'ID':<5} {'Nama':<25} {'Harga':<15} {'Kategori':<15} {'Status':<10}")
        print("-" * 80)
        
        for product in products:
            pid, name, desc, price, category, img_url, dl_link, is_active, created = product
            status = "Aktif" if is_active else "Nonaktif"
            name_short = name[:22] + "..." if len(name) > 25 else name
            print(f"{pid:<5} {name_short:<25} Rp {price:,}    {category:<15} {status:<10}")
        
        print(f"\nTotal: {len(products)} produk")
    
    def list_users(self):
        """List all users"""
        conn = self.db.get_connection()
        cursor = conn.cursor()
        
        cursor.execute('''
            SELECT u.user_id, u.username, u.first_name, u.last_name, u.join_date,
                   COUNT(o.id) as total_orders,
                   COALESCE(SUM(o.total_price), 0) as total_spent
            FROM users u
            LEFT JOIN orders o ON u.user_id = o.user_id
            GROUP BY u.user_id
            ORDER BY u.join_date DESC
        ''')
        
        users = cursor.fetchall()
        conn.close()
        
        print("👥 DAFTAR PENGGUNA")
        print("=" * 100)
        print(f"{'User ID':<12} {'Username':<20} {'Nama':<25} {'Total Order':<12} {'Total Belanja':<15}")
        print("-" * 100)
        
        for user in users:
            user_id, username, first_name, last_name, join_date, total_orders, total_spent = user
            full_name = f"{first_name or ''} {last_name or ''}".strip()
            username_display = f"@{username}" if username else "N/A"
            
            print(f"{user_id:<12} {username_display:<20} {full_name:<25} {total_orders:<12} Rp {total_spent:,}")
        
        print(f"\nTotal: {len(users)} pengguna")
    
    def list_orders(self, status=None):
        """List orders by status"""
        conn = self.db.get_connection()
        cursor = conn.cursor()
        
        if status:
            cursor.execute('''
                SELECT o.id, o.user_id, u.first_name, u.last_name, p.name,
                       o.quantity, o.total_price, o.payment_method, o.payment_status, o.order_date
                FROM orders o
                JOIN users u ON o.user_id = u.user_id
                JOIN products p ON o.product_id = p.id
                WHERE o.payment_status = ?
                ORDER BY o.order_date DESC
            ''', (status,))
        else:
            cursor.execute('''
                SELECT o.id, o.user_id, u.first_name, u.last_name, p.name,
                       o.quantity, o.total_price, o.payment_method, o.payment_status, o.order_date
                FROM orders o
                JOIN users u ON o.user_id = u.user_id
                JOIN products p ON o.product_id = p.id
                ORDER BY o.order_date DESC
            ''')
        
        orders = cursor.fetchall()
        conn.close()
        
        status_text = f" - Status: {status.upper()}" if status else ""
        print(f"💰 DAFTAR PESANAN{status_text}")
        print("=" * 120)
        print(f"{'Order ID':<10} {'User ID':<12} {'Nama User':<20} {'Produk':<25} {'Qty':<5} {'Total':<15} {'Status':<10}")
        print("-" * 120)
        
        for order in orders:
            order_id, user_id, first_name, last_name, product_name, quantity, total_price, payment_method, payment_status, order_date = order
            user_name = f"{first_name or ''} {last_name or ''}".strip()
            product_short = product_name[:22] + "..." if len(product_name) > 25 else product_name
            
            print(f"#{order_id:<9} {user_id:<12} {user_name:<20} {product_short:<25} {quantity:<5} Rp {total_price:,}   {payment_status:<10}")
        
        print(f"\nTotal: {len(orders)} pesanan")
    
    def update_order_status(self):
        """Update order payment status"""
        print("💳 UPDATE STATUS PEMBAYARAN")
        print("=" * 50)
        
        # Show pending orders
        self.list_orders('pending')
        print()
        
        try:
            order_id = int(input("Masukkan Order ID: "))
            print("Status yang tersedia: pending, completed, cancelled")
            new_status = input("Status baru: ").lower()
            
            if new_status not in ['pending', 'completed', 'cancelled']:
                print("❌ Status tidak valid!")
                return
            
            self.db.update_payment_status(order_id, new_status)
            print(f"✅ Status order #{order_id} berhasil diubah menjadi {new_status}")
            
        except ValueError:
            print("❌ Order ID harus berupa angka!")
        except Exception as e:
            print(f"❌ Error: {e}")
    
    def show_statistics(self):
        """Show bot statistics"""
        conn = self.db.get_connection()
        cursor = conn.cursor()
        
        # Total users
        cursor.execute("SELECT COUNT(*) FROM users")
        total_users = cursor.fetchone()[0]
        
        # Total products
        cursor.execute("SELECT COUNT(*) FROM products WHERE is_active = 1")
        total_products = cursor.fetchone()[0]
        
        # Total orders
        cursor.execute("SELECT COUNT(*) FROM orders")
        total_orders = cursor.fetchone()[0]
        
        # Total revenue
        cursor.execute("SELECT COALESCE(SUM(total_price), 0) FROM orders WHERE payment_status = 'completed'")
        total_revenue = cursor.fetchone()[0]
        
        # Pending orders
        cursor.execute("SELECT COUNT(*) FROM orders WHERE payment_status = 'pending'")
        pending_orders = cursor.fetchone()[0]
        
        # Today's orders
        cursor.execute("SELECT COUNT(*) FROM orders WHERE DATE(order_date) = DATE('now')")
        today_orders = cursor.fetchone()[0]
        
        conn.close()
        
        print("📊 STATISTIK BOT")
        print("=" * 50)
        print(f"👥 Total Pengguna      : {total_users:,}")
        print(f"📱 Total Produk Aktif  : {total_products:,}")
        print(f"📦 Total Pesanan       : {total_orders:,}")
        print(f"💰 Total Pendapatan    : Rp {total_revenue:,}")
        print(f"⏳ Pesanan Pending     : {pending_orders:,}")
        print(f"📅 Pesanan Hari Ini    : {today_orders:,}")
        print("=" * 50)

def main():
    """Main CLI interface"""
    admin = AdminTools()
    
    while True:
        print("\n🔧 ADMIN TOOLS - BOT TELEGRAM PREMIUM APPS")
        print("=" * 50)
        print("1. Tambah Produk")
        print("2. Lihat Daftar Produk")
        print("3. Lihat Daftar Pengguna")
        print("4. Lihat Semua Pesanan")
        print("5. Lihat Pesanan Pending")
        print("6. Update Status Pembayaran")
        print("7. Statistik Bot")
        print("0. Keluar")
        print("=" * 50)
        
        try:
            choice = input("Pilih menu (0-7): ").strip()
            
            if choice == "1":
                admin.add_product_cli()
            elif choice == "2":
                admin.list_products()
            elif choice == "3":
                admin.list_users()
            elif choice == "4":
                admin.list_orders()
            elif choice == "5":
                admin.list_orders('pending')
            elif choice == "6":
                admin.update_order_status()
            elif choice == "7":
                admin.show_statistics()
            elif choice == "0":
                print("👋 Terima kasih!")
                break
            else:
                print("❌ Pilihan tidak valid!")
                
        except KeyboardInterrupt:
            print("\n👋 Terima kasih!")
            break
        except Exception as e:
            print(f"❌ Error: {e}")

if __name__ == '__main__':
    main()