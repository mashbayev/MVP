#!/bin/bash

# Development environment setup script
# This sets up minimal environment variables for local development

export GEMINI_API_KEY="${GEMINI_API_KEY:-dev-gemini-key-placeholder}"
export OPENWEATHER_API_KEY="${OPENWEATHER_API_KEY:-dev-weather-key-placeholder}"
export TELEGRAM_BOT_TOKEN="${TELEGRAM_BOT_TOKEN:-dev-telegram-token-placeholder}"
export WAZZUP_API_KEY="${WAZZUP_API_KEY:-dev-wazzup-key-placeholder}"
export OPENAI_API_KEY="${OPENAI_API_KEY:-dev-openai-key-placeholder}"

echo "✅ Development environment variables set:"
echo "   GEMINI_API_KEY: ${GEMINI_API_KEY:0:30}..."
echo "   OPENWEATHER_API_KEY: ${OPENWEATHER_API_KEY:0:30}..."
echo "   TELEGRAM_BOT_TOKEN: ${TELEGRAM_BOT_TOKEN:0:30}..."
echo "   WAZZUP_API_KEY: ${WAZZUP_API_KEY:0:30}..."
echo "   OPENAI_API_KEY: ${OPENAI_API_KEY:0:30}..."
echo ""
echo "⚠️  These are placeholder values for development only!"
echo "Replace them with actual API keys for production."
echo ""
echo "Starting server..."
cd "$(dirname "$0")" || exit 1
go run cmd/main.go
