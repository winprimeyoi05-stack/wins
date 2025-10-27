# Dockerfile for Telegram Premium App Sales Bot
FROM python:3.9-slim

# Set working directory
WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements first for better caching
COPY requirements.txt .

# Install Python dependencies
RUN pip install --no-cache-dir -r requirements.txt

# Copy application code
COPY . .

# Create directory for database
RUN mkdir -p /app/data

# Set environment variables
ENV PYTHONPATH=/app
ENV DATABASE_PATH=/app/data/bot_database.db

# Create non-root user
RUN useradd -m -u 1000 botuser && chown -R botuser:botuser /app
USER botuser

# Expose port (if needed for webhooks)
EXPOSE 8000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import sqlite3; sqlite3.connect('${DATABASE_PATH}').close()" || exit 1

# Run the bot
CMD ["python", "bot.py"]