#!/usr/bin/env python3
"""
Simple runner script for the Telegram Premium App Sales Bot
"""

import sys
import os

# Add current directory to Python path
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))

try:
    from bot import TelegramBot
    
    if __name__ == '__main__':
        print("🚀 Starting Telegram Premium App Sales Bot...")
        print("📋 Make sure you have:")
        print("   1. Created a .env file with BOT_TOKEN and ADMIN_IDS")
        print("   2. Installed all dependencies: pip install -r requirements.txt")
        print("   3. Set up your bot with @BotFather")
        print()
        
        bot = TelegramBot()
        bot.run()
        
except ImportError as e:
    print(f"❌ Import Error: {e}")
    print("📦 Please install dependencies: pip install -r requirements.txt")
    sys.exit(1)
except Exception as e:
    print(f"❌ Error starting bot: {e}")
    print("🔧 Please check your configuration in .env file")
    sys.exit(1)