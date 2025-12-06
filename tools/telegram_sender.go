// Package tools contains helper functions for development and testing tools
package tools

import (
	"fmt"
	"log"
	"os"

	"whatsapp-analytics-mvp/internal/infrastructure"
)

// RunTelegramSenderTest runs Telegram sender tests
func RunTelegramSenderTest() {
	// Get token from environment
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set in environment")
	}

	// Get chat ID from environment (optional, use a test one if not set)
	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("TELEGRAM_CHAT_ID is not set in environment")
	}

	// Parse chat ID
	var chatID int64
	_, err := fmt.Sscanf(chatIDStr, "%d", &chatID)
	if err != nil {
		log.Fatalf("Invalid TELEGRAM_CHAT_ID: %v", err)
	}

	// Create TelegramSender
	sender := infrastructure.NewTelegramSender(token)

	// Test 1: Send typing action
	log.Println("\n=== TEST 1: SendTyping ===")
	err = sender.SendTyping(chatID)
	if err != nil {
		log.Printf("❌ SendTyping failed: %v\n", err)
	} else {
		log.Println("✅ SendTyping succeeded")
	}

	// Test 2: Send text message
	log.Println("=== TEST 2: Send text message ===")
	err = sender.Send(chatID, "Test message via Codex v2 - with proper error handling")
	if err != nil {
		log.Printf("❌ Send failed: %v\n", err)
	} else {
		log.Println("✅ Send succeeded")
	}

	// Test 3: Send another message
	log.Println("=== TEST 3: Send another message ===")
	err = sender.Send(chatID, "Validation and error handling are now working!")
	if err != nil {
		log.Printf("❌ Send failed: %v\n", err)
	} else {
		log.Println("✅ Send succeeded")
	}
}
