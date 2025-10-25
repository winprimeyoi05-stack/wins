import logging
from telegram import Update, InlineKeyboardButton, InlineKeyboardMarkup
from telegram.ext import Application, CommandHandler, CallbackQueryHandler, ContextTypes, MessageHandler, filters
from database import Database
from config import BOT_TOKEN, ADMIN_IDS, MESSAGES, PAYMENT_METHODS
import qrcode
import io
from datetime import datetime

# Enable logging
logging.basicConfig(
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    level=logging.INFO
)
logger = logging.getLogger(__name__)

# Initialize database
db = Database()

class TelegramBot:
    def __init__(self):
        self.application = Application.builder().token(BOT_TOKEN).build()
        self.setup_handlers()
    
    def setup_handlers(self):
        """Setup all command and callback handlers"""
        # Command handlers
        self.application.add_handler(CommandHandler("start", self.start_command))
        self.application.add_handler(CommandHandler("help", self.help_command))
        self.application.add_handler(CommandHandler("catalog", self.catalog_command))
        self.application.add_handler(CommandHandler("cart", self.cart_command))
        self.application.add_handler(CommandHandler("history", self.history_command))
        self.application.add_handler(CommandHandler("contact", self.contact_command))
        
        # Admin commands
        self.application.add_handler(CommandHandler("admin", self.admin_command))
        self.application.add_handler(CommandHandler("addproduct", self.add_product_command))
        self.application.add_handler(CommandHandler("users", self.users_command))
        
        # Callback query handlers
        self.application.add_handler(CallbackQueryHandler(self.handle_callback))
        
        # Message handlers
        self.application.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, self.handle_message))
    
    async def start_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /start command"""
        user = update.effective_user
        
        # Add user to database
        db.add_user(user.id, user.username, user.first_name, user.last_name)
        
        keyboard = [
            [InlineKeyboardButton("📱 Lihat Katalog", callback_data="catalog")],
            [InlineKeyboardButton("🛒 Keranjang", callback_data="cart"),
             InlineKeyboardButton("📞 Kontak", callback_data="contact")],
            [InlineKeyboardButton("ℹ️ Bantuan", callback_data="help")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await update.message.reply_text(
            MESSAGES['welcome'],
            reply_markup=reply_markup
        )
    
    async def help_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /help command"""
        await update.message.reply_text(MESSAGES['help'])
    
    async def catalog_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /catalog command"""
        await self.show_catalog(update, context)
    
    async def show_catalog(self, update: Update, context: ContextTypes.DEFAULT_TYPE, category=None):
        """Show product catalog"""
        products = db.get_products(category)
        categories = db.get_categories()
        
        if not products:
            text = "🚫 Tidak ada produk yang tersedia saat ini."
            keyboard = [[InlineKeyboardButton("🔙 Kembali", callback_data="start")]]
        else:
            text = f"📱 **KATALOG APLIKASI PREMIUM**\n"
            if category:
                text += f"Kategori: {category}\n"
            text += "\n"
            
            keyboard = []
            
            # Category filter buttons
            if not category:
                text += "🏷️ **Filter berdasarkan kategori:**\n"
                cat_buttons = []
                for cat in categories:
                    cat_buttons.append(InlineKeyboardButton(f"📂 {cat}", callback_data=f"category_{cat}"))
                    if len(cat_buttons) == 2:
                        keyboard.append(cat_buttons)
                        cat_buttons = []
                if cat_buttons:
                    keyboard.append(cat_buttons)
                keyboard.append([InlineKeyboardButton("📋 Semua Produk", callback_data="catalog")])
                text += "\n"
            
            # Product list
            for product in products:
                product_id, name, desc, price, cat, img_url, dl_link, is_active, created = product
                text += f"🔸 **{name}**\n"
                text += f"💰 Rp {price:,}\n"
                text += f"📝 {desc[:50]}{'...' if len(desc) > 50 else ''}\n"
                
                keyboard.append([
                    InlineKeyboardButton(f"👁️ Detail", callback_data=f"product_{product_id}"),
                    InlineKeyboardButton(f"🛒 Beli", callback_data=f"buy_{product_id}")
                ])
                text += "\n"
            
            # Navigation buttons
            nav_buttons = []
            if category:
                nav_buttons.append(InlineKeyboardButton("🔙 Semua Kategori", callback_data="catalog"))
            nav_buttons.append(InlineKeyboardButton("🏠 Menu Utama", callback_data="start"))
            keyboard.append(nav_buttons)
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        if update.callback_query:
            await update.callback_query.edit_message_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
        else:
            await update.message.reply_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
    
    async def show_product_detail(self, update: Update, context: ContextTypes.DEFAULT_TYPE, product_id):
        """Show detailed product information"""
        product = db.get_product(product_id)
        
        if not product:
            await update.callback_query.answer("❌ Produk tidak ditemukan!")
            return
        
        pid, name, desc, price, category, img_url, dl_link, is_active, created = product
        
        text = f"📱 **{name}**\n\n"
        text += f"📝 **Deskripsi:**\n{desc}\n\n"
        text += f"💰 **Harga:** Rp {price:,}\n"
        text += f"🏷️ **Kategori:** {category}\n\n"
        text += f"✅ **Status:** Tersedia\n"
        
        keyboard = [
            [InlineKeyboardButton("🛒 Tambah ke Keranjang", callback_data=f"addcart_{product_id}")],
            [InlineKeyboardButton("💳 Beli Sekarang", callback_data=f"buynow_{product_id}")],
            [InlineKeyboardButton("🔙 Kembali ke Katalog", callback_data="catalog")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await update.callback_query.edit_message_text(
            text, reply_markup=reply_markup, parse_mode='Markdown'
        )
    
    async def cart_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /cart command"""
        await self.show_cart(update, context)
    
    async def show_cart(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Show user's shopping cart"""
        user_id = update.effective_user.id
        cart_items = db.get_cart(user_id)
        
        if not cart_items:
            text = "🛒 **KERANJANG BELANJA**\n\n"
            text += "Keranjang Anda kosong.\n"
            text += "Silakan pilih produk dari katalog terlebih dahulu."
            
            keyboard = [
                [InlineKeyboardButton("📱 Lihat Katalog", callback_data="catalog")],
                [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
            ]
        else:
            text = "🛒 **KERANJANG BELANJA**\n\n"
            total_price = 0
            
            for item in cart_items:
                cart_id, quantity, name, price, product_id = item
                subtotal = price * quantity
                total_price += subtotal
                
                text += f"🔸 **{name}**\n"
                text += f"   Jumlah: {quantity} x Rp {price:,} = Rp {subtotal:,}\n\n"
            
            text += f"💰 **Total: Rp {total_price:,}**\n"
            
            keyboard = [
                [InlineKeyboardButton("💳 Checkout", callback_data="checkout")],
                [InlineKeyboardButton("🗑️ Kosongkan Keranjang", callback_data="clearcart")],
                [InlineKeyboardButton("📱 Lanjut Belanja", callback_data="catalog")],
                [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
            ]
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        if update.callback_query:
            await update.callback_query.edit_message_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
        else:
            await update.message.reply_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
    
    async def history_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /history command"""
        user_id = update.effective_user.id
        orders = db.get_user_orders(user_id)
        
        if not orders:
            text = "📋 **RIWAYAT PEMBELIAN**\n\n"
            text += "Anda belum memiliki riwayat pembelian."
            
            keyboard = [
                [InlineKeyboardButton("📱 Mulai Belanja", callback_data="catalog")],
                [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
            ]
        else:
            text = "📋 **RIWAYAT PEMBELIAN**\n\n"
            
            for order in orders:
                order_id, quantity, total_price, payment_method, status, order_date, product_name = order
                
                # Format date
                date_obj = datetime.strptime(order_date, "%Y-%m-%d %H:%M:%S")
                formatted_date = date_obj.strftime("%d/%m/%Y %H:%M")
                
                status_emoji = "✅" if status == "completed" else "⏳" if status == "pending" else "❌"
                
                text += f"🔸 **{product_name}**\n"
                text += f"   Order ID: #{order_id}\n"
                text += f"   Jumlah: {quantity}\n"
                text += f"   Total: Rp {total_price:,}\n"
                text += f"   Pembayaran: {payment_method}\n"
                text += f"   Status: {status_emoji} {status.title()}\n"
                text += f"   Tanggal: {formatted_date}\n\n"
            
            keyboard = [
                [InlineKeyboardButton("📱 Belanja Lagi", callback_data="catalog")],
                [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
            ]
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        if update.callback_query:
            await update.callback_query.edit_message_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
        else:
            await update.message.reply_text(
                text, reply_markup=reply_markup, parse_mode='Markdown'
            )
    
    async def contact_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /contact command"""
        keyboard = [
            [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        if update.callback_query:
            await update.callback_query.edit_message_text(
                MESSAGES['contact'], reply_markup=reply_markup
            )
        else:
            await update.message.reply_text(
                MESSAGES['contact'], reply_markup=reply_markup
            )
    
    async def show_checkout(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Show checkout options"""
        user_id = update.effective_user.id
        cart_items = db.get_cart(user_id)
        
        if not cart_items:
            await update.callback_query.answer("❌ Keranjang kosong!")
            return
        
        total_price = sum(item[2] * item[1] for item in cart_items)  # price * quantity
        
        text = "💳 **CHECKOUT**\n\n"
        text += f"💰 Total Pembayaran: **Rp {total_price:,}**\n\n"
        text += "Pilih metode pembayaran:\n"
        
        keyboard = []
        for method, details in PAYMENT_METHODS.items():
            keyboard.append([InlineKeyboardButton(f"💳 {details.split(':')[0]}", callback_data=f"pay_{method}")])
        
        keyboard.append([InlineKeyboardButton("🔙 Kembali ke Keranjang", callback_data="cart")])
        
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await update.callback_query.edit_message_text(
            text, reply_markup=reply_markup, parse_mode='Markdown'
        )
    
    async def process_payment(self, update: Update, context: ContextTypes.DEFAULT_TYPE, payment_method):
        """Process payment for cart items"""
        user_id = update.effective_user.id
        cart_items = db.get_cart(user_id)
        
        if not cart_items:
            await update.callback_query.answer("❌ Keranjang kosong!")
            return
        
        total_price = sum(item[2] * item[1] for item in cart_items)  # price * quantity
        payment_details = PAYMENT_METHODS[payment_method]
        
        # Create orders for each cart item
        order_ids = []
        for item in cart_items:
            cart_id, quantity, name, price, product_id = item
            subtotal = price * quantity
            order_id = db.create_order(user_id, product_id, quantity, subtotal, payment_method)
            order_ids.append(order_id)
        
        # Clear cart
        db.clear_cart(user_id)
        
        text = "💳 **PEMBAYARAN**\n\n"
        text += f"💰 Total: **Rp {total_price:,}**\n"
        text += f"🏦 Metode: {payment_details}\n\n"
        text += "📋 **Instruksi Pembayaran:**\n"
        text += f"1. Transfer ke: {payment_details}\n"
        text += f"2. Nominal: Rp {total_price:,}\n"
        text += f"3. Kirim bukti transfer ke admin\n"
        text += f"4. Tunggu konfirmasi dari admin\n\n"
        text += f"🆔 Order ID: {', '.join([f'#{oid}' for oid in order_ids])}\n\n"
        text += "⚠️ **Penting:** Simpan Order ID untuk referensi Anda!"
        
        keyboard = [
            [InlineKeyboardButton("📞 Hubungi Admin", callback_data="contact")],
            [InlineKeyboardButton("📋 Riwayat Pembelian", callback_data="history")],
            [InlineKeyboardButton("🏠 Menu Utama", callback_data="start")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await update.callback_query.edit_message_text(
            text, reply_markup=reply_markup, parse_mode='Markdown'
        )
    
    async def admin_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /admin command"""
        user_id = update.effective_user.id
        
        if user_id not in ADMIN_IDS:
            await update.message.reply_text("❌ Anda tidak memiliki akses admin!")
            return
        
        text = "👨‍💼 **PANEL ADMIN**\n\n"
        text += "Pilih menu admin:"
        
        keyboard = [
            [InlineKeyboardButton("📊 Statistik", callback_data="admin_stats")],
            [InlineKeyboardButton("👥 Daftar User", callback_data="admin_users")],
            [InlineKeyboardButton("📦 Kelola Produk", callback_data="admin_products")],
            [InlineKeyboardButton("💰 Kelola Pesanan", callback_data="admin_orders")]
        ]
        reply_markup = InlineKeyboardMarkup(keyboard)
        
        await update.message.reply_text(text, reply_markup=reply_markup, parse_mode='Markdown')
    
    async def add_product_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /addproduct command (simplified for demo)"""
        user_id = update.effective_user.id
        
        if user_id not in ADMIN_IDS:
            await update.message.reply_text("❌ Anda tidak memiliki akses admin!")
            return
        
        await update.message.reply_text(
            "📝 **TAMBAH PRODUK BARU**\n\n"
            "Format: /addproduct <nama> | <deskripsi> | <harga> | <kategori>\n\n"
            "Contoh:\n"
            "/addproduct Discord Nitro | Premium Discord features | 50000 | Gaming"
        )
    
    async def users_command(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle /users command"""
        user_id = update.effective_user.id
        
        if user_id not in ADMIN_IDS:
            await update.message.reply_text("❌ Anda tidak memiliki akses admin!")
            return
        
        # This would show user statistics in a real implementation
        await update.message.reply_text("👥 **STATISTIK PENGGUNA**\n\nFitur ini akan dikembangkan lebih lanjut.")
    
    async def handle_callback(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle all callback queries"""
        query = update.callback_query
        await query.answer()
        
        data = query.data
        
        if data == "start":
            await self.start_command(update, context)
        elif data == "help":
            await query.edit_message_text(MESSAGES['help'])
        elif data == "catalog":
            await self.show_catalog(update, context)
        elif data.startswith("category_"):
            category = data.split("_", 1)[1]
            await self.show_catalog(update, context, category)
        elif data.startswith("product_"):
            product_id = int(data.split("_")[1])
            await self.show_product_detail(update, context, product_id)
        elif data.startswith("addcart_"):
            product_id = int(data.split("_")[1])
            user_id = update.effective_user.id
            db.add_to_cart(user_id, product_id)
            await query.answer("✅ Produk ditambahkan ke keranjang!")
        elif data.startswith("buy_") or data.startswith("buynow_"):
            product_id = int(data.split("_")[1])
            user_id = update.effective_user.id
            db.add_to_cart(user_id, product_id)
            await self.show_cart(update, context)
        elif data == "cart":
            await self.show_cart(update, context)
        elif data == "clearcart":
            user_id = update.effective_user.id
            db.clear_cart(user_id)
            await query.answer("🗑️ Keranjang dikosongkan!")
            await self.show_cart(update, context)
        elif data == "checkout":
            await self.show_checkout(update, context)
        elif data.startswith("pay_"):
            payment_method = data.split("_")[1]
            await self.process_payment(update, context, payment_method)
        elif data == "history":
            await self.history_command(update, context)
        elif data == "contact":
            await self.contact_command(update, context)
    
    async def handle_message(self, update: Update, context: ContextTypes.DEFAULT_TYPE):
        """Handle text messages"""
        # This can be expanded to handle product search, etc.
        await update.message.reply_text(
            "ℹ️ Gunakan menu atau ketik /help untuk melihat perintah yang tersedia."
        )
    
    def run(self):
        """Start the bot"""
        print("🤖 Bot Telegram Penjualan Aplikasi Premium dimulai...")
        self.application.run_polling()

if __name__ == '__main__':
    bot = TelegramBot()
    bot.run()