#!/bin/bash

# Script to check Telegram updates via Long Polling
# Usage: ./check-updates.sh

if [ -z "$TELEGRAM_TOKEN" ]; then
    export TELEGRAM_TOKEN="8203439505:AAE5RfnnBgHhNPJLMGqcA8GbPwJkDnkl-t8"
fi

echo "🔍 Checking for updates via Long Polling..."
echo "Bot Token: ...${TELEGRAM_TOKEN: -5}"
echo ""

# 1. Ensure webhook is deleted (required for getUpdates)
curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/deleteWebhook" > /dev/null

# 2. Get updates
RESPONSE=$(curl -s "https://api.telegram.org/bot$TELEGRAM_TOKEN/getUpdates?limit=5")

echo "📥 Response from Telegram:"
echo "$RESPONSE" | jq '.'
