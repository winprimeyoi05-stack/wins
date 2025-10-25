#!/usr/bin/env python3
"""
Setup script for Telegram Premium App Sales Bot
"""

import os
import sys
import subprocess

def check_python_version():
    """Check if Python version is compatible"""
    if sys.version_info < (3, 7):
        print("❌ Python 3.7 atau lebih baru diperlukan!")
        print(f"   Versi Python Anda: {sys.version}")
        return False
    return True

def install_dependencies():
    """Install required dependencies"""
    print("📦 Installing dependencies...")
    try:
        subprocess.check_call([sys.executable, "-m", "pip", "install", "-r", "requirements.txt"])
        print("✅ Dependencies berhasil diinstall!")
        return True
    except subprocess.CalledProcessError:
        print("❌ Gagal menginstall dependencies!")
        return False

def create_env_file():
    """Create .env file from template"""
    if os.path.exists('.env'):
        print("⚠️  File .env sudah ada, tidak akan ditimpa.")
        return True
    
    if os.path.exists('.env.example'):
        print("📄 Membuat file .env dari template...")
        with open('.env.example', 'r') as src, open('.env', 'w') as dst:
            dst.write(src.read())
        print("✅ File .env berhasil dibuat!")
        print("⚠️  PENTING: Edit file .env dan isi BOT_TOKEN serta ADMIN_IDS Anda!")
        return True
    else:
        print("❌ File .env.example tidak ditemukan!")
        return False

def setup_database():
    """Initialize database"""
    print("🗄️  Initializing database...")
    try:
        from database import Database
        db = Database()
        print("✅ Database berhasil diinisialisasi!")
        return True
    except Exception as e:
        print(f"❌ Gagal menginisialisasi database: {e}")
        return False

def show_next_steps():
    """Show next steps to user"""
    print("\n" + "="*60)
    print("🎉 SETUP SELESAI!")
    print("="*60)
    print("📋 Langkah selanjutnya:")
    print()
    print("1. 🤖 Buat bot Telegram:")
    print("   - Chat dengan @BotFather di Telegram")
    print("   - Ketik /newbot dan ikuti instruksi")
    print("   - Copy token yang diberikan")
    print()
    print("2. 🆔 Dapatkan User ID Anda:")
    print("   - Chat dengan @userinfobot di Telegram")
    print("   - Copy User ID Anda")
    print()
    print("3. ⚙️  Edit file .env:")
    print("   - Buka file .env dengan text editor")
    print("   - Ganti YOUR_BOT_TOKEN_HERE dengan token bot Anda")
    print("   - Ganti 123456789 dengan User ID Anda")
    print()
    print("4. 🚀 Jalankan bot:")
    print("   python bot.py")
    print("   atau")
    print("   python run.py")
    print()
    print("5. 🔧 Kelola bot (opsional):")
    print("   python admin_tools.py")
    print()
    print("="*60)
    print("📚 Baca README.md untuk dokumentasi lengkap!")
    print("="*60)

def main():
    """Main setup function"""
    print("🚀 SETUP BOT TELEGRAM PENJUALAN APLIKASI PREMIUM")
    print("="*60)
    
    # Check Python version
    if not check_python_version():
        sys.exit(1)
    
    # Install dependencies
    if not install_dependencies():
        sys.exit(1)
    
    # Create .env file
    if not create_env_file():
        sys.exit(1)
    
    # Setup database
    if not setup_database():
        sys.exit(1)
    
    # Show next steps
    show_next_steps()

if __name__ == '__main__':
    main()