import os
from dotenv import load_dotenv

load_dotenv()

# Bot Configuration
BOT_TOKEN = os.getenv('BOT_TOKEN', 'YOUR_BOT_TOKEN_HERE')
ADMIN_IDS = [int(x) for x in os.getenv('ADMIN_IDS', '').split(',') if x.strip()]

# Database Configuration
DATABASE_PATH = 'bot_database.db'

# Payment Configuration (Mock for demo)
PAYMENT_METHODS = {
    'dana': 'DANA: 081234567890',
    'gopay': 'GoPay: 081234567890', 
    'ovo': 'OVO: 081234567890',
    'bank': 'BCA: 1234567890 a.n. Toko Premium Apps'
}

# Messages in Indonesian
MESSAGES = {
    'welcome': """
🎉 Selamat datang di Bot Penjualan Aplikasi Premium! 🎉

Kami menyediakan berbagai aplikasi premium berkualitas tinggi dengan harga terjangkau.

📱 Fitur Bot:
• Katalog aplikasi premium lengkap
• Sistem pembayaran mudah
• Support 24/7
• Garansi aplikasi

Ketik /help untuk melihat semua perintah yang tersedia.
""",
    
    'help': """
📋 DAFTAR PERINTAH:

🏠 /start - Mulai menggunakan bot
📱 /catalog - Lihat katalog aplikasi
🛒 /cart - Lihat keranjang belanja
💰 /history - Riwayat pembelian
📞 /contact - Hubungi admin
ℹ️ /help - Bantuan

👨‍💼 ADMIN ONLY:
/admin - Panel admin
/addproduct - Tambah produk baru
/removeproduct - Hapus produk
/users - Lihat daftar pengguna
""",
    
    'contact': """
📞 HUBUNGI KAMI:

👨‍💼 Admin: @admin_username
📧 Email: admin@premiumapps.com
⏰ Jam Operasional: 08:00 - 22:00 WIB

Untuk pertanyaan lebih lanjut, silakan hubungi admin di atas.
"""
}