package infrastructure

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// ============================================================================
// TELEGRAM SENDER
// ============================================================================

type TelegramSender struct {
	Token  string
	Client *http.Client
}

func NewTelegramSender(token string) *TelegramSender {
	return &TelegramSender{
		Token:  token,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (t *TelegramSender) Send(chatID int64, text string) error {
	if t.Token == "" {
		log.Println("[TelegramSender] ⚠️ Token not set — message skipped")
		return nil
	}

	log.Printf("[TelegramSender] → Send to %d: %s", chatID, text)
	// Реальный API вызов можно добавить позже
	return nil
}

func (t *TelegramSender) SendTyping(chatID int64) error {
	log.Printf("[TelegramSender] … typing to %d", chatID)
	return nil
}

func (t *TelegramSender) GetFileDirectURL(fileID string) (string, error) {
	if fileID == "" {
		return "", fmt.Errorf("fileID пустой")
	}
	// Реализация простая — для MVP достаточно
	return fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", t.Token, fileID), nil
}

// ============================================================================
// WAZZUP (WHATSAPP) SENDER
// ============================================================================

type WazzupSender struct {
	APIKey string
}

func NewWazzupSender(apiKey string) *WazzupSender {
	return &WazzupSender{APIKey: apiKey}
}

func (w *WazzupSender) Send(channelID, clientPhone, messageText string) error {
	if w.APIKey == "" {
		log.Println("[WazzupSender] ⚠️ API Key not set — skipping send")
		return nil
	}

	log.Printf("[WazzupSender] → Send to %s (%s): %s",
		clientPhone, channelID, messageText)
	return nil
}

// ============================================================================
// NOTIFIER (ADMIN ALERTING)
// ============================================================================

type DefaultNotifier struct{}

func (n *DefaultNotifier) NotifyAdmin(message string) error {
	log.Printf("🔔 ADMIN NOTIFY: %s", message)
	return nil
}

// ============================================================================
// EVENT BUS (STUB)
// ============================================================================

type MockEventBus struct{}

func (m *MockEventBus) Publish(topic string, payload interface{}) {
	log.Printf("[EventBus] %s → %v", topic, payload)
}

// ============================================================================
// TASK MANAGER (STUB)
// ============================================================================

type MockTaskManager struct{}

func (m *MockTaskManager) Schedule(taskName string, f func()) {
	log.Printf("[TaskManager] Schedule: %s", taskName)
}

// ============================================================================
// TRANSCRIPTION (STUB FOR MVP)
// ============================================================================

type MockTranscriber struct{}

func (t *MockTranscriber) Transcribe(url string) (string, error) {
	log.Printf("[MockTranscriber] Transcribing: %s", url)
	return "[транскрипция недоступна в MVP]", nil
}
