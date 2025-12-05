#!/bin/bash

# Telegram Webhook Diagnostics Script

if [ -z "$TELEGRAM_TOKEN" ]; then
    echo "❌ TELEGRAM_TOKEN not set!"
    exit 1
fi

echo "🔍 Telegram Webhook Diagnostics"
echo "================================"
echo ""

# 1. Check webhook info
echo "1️⃣ Current Webhook Configuration:"
curl -s "https://api.telegram.org/bot${TELEGRAM_TOKEN}/getWebhookInfo" | jq '.'
echo ""

# 2. Get bot info
echo "2️⃣ Bot Information:"
curl -s "https://api.telegram.org/bot${TELEGRAM_TOKEN}/getMe" | jq '.'
echo ""

# 3. Check for pending updates
echo "3️⃣ Pending Updates (if any):"
curl -s "https://api.telegram.org/bot${TELEGRAM_TOKEN}/getUpdates?limit=1" | jq '.'
echo ""

echo "✅ Diagnostics complete!"
