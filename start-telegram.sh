#!/bin/bash

# Telegram Bot Quick Start Script

echo "🤖 WhatsApp Analytics MVP - Telegram Bot"
echo "========================================"
echo ""

# Check if TELEGRAM_TOKEN is set
if [ -z "$TELEGRAM_TOKEN" ]; then
    echo "⚠️  TELEGRAM_TOKEN not set, using known working token..."
    export TELEGRAM_TOKEN="8203439505:AAE5RfnnBgHhNPJLMGqcA8GbPwJkDnkl-t8"
fi

echo "✓ Telegram token found"
echo ""

# Build the application
echo "🔨 Building application..."
go build -o whatsapp-analytics ./cmd/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✓ Build successful"
echo ""

# Create db directory if it doesn't exist
mkdir -p db

echo "🚀 Starting server on port 8080..."
echo ""
echo "📝 Next steps:"
echo "1. In another terminal, run: ngrok http 8080"
echo "2. Copy the ngrok HTTPS URL"
echo "3. Set webhook: curl -X POST \"https://api.telegram.org/bot\$TELEGRAM_TOKEN/setWebhook?url=https://YOUR_NGROK_URL/webhook/telegram\""
echo ""
echo "Press Ctrl+C to stop the server"
echo "========================================"
echo ""

# Run the application
./whatsapp-analytics
