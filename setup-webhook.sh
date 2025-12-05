#!/bin/bash

# Telegram Webhook Setup Script

if [ -z "$1" ]; then
    echo "Usage: ./setup-webhook.sh <ngrok_https_url>"
    echo "Example: ./setup-webhook.sh https://abcd-12-34-56-78.ngrok-free.app"
    exit 1
fi

if [ -z "$TELEGRAM_TOKEN" ]; then
    echo "❌ TELEGRAM_TOKEN not set!"
    echo "Please run: export TELEGRAM_TOKEN=\"your_bot_token\""
    exit 1
fi

NGROK_URL=$1
WEBHOOK_URL="${NGROK_URL}/webhook/telegram"

echo "🔧 Setting up Telegram webhook..."
echo "Webhook URL: $WEBHOOK_URL"
echo ""

# Set webhook
RESPONSE=$(curl -s -X POST "https://api.telegram.org/bot${TELEGRAM_TOKEN}/setWebhook?url=${WEBHOOK_URL}")
echo "Response: $RESPONSE"
echo ""

# Get webhook info
echo "📊 Webhook info:"
curl -s "https://api.telegram.org/bot${TELEGRAM_TOKEN}/getWebhookInfo" | jq '.'
