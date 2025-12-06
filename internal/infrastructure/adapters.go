package infrastructure

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		return fmt.Errorf("telegram token not set")
	}

	// Ensure http client with timeout
	if t.Client == nil {
		t.Client = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.Token)

	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[TelegramSender] NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		log.Printf("[TelegramSender] HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("[TelegramSender] Non-200 response: status=%d raw=%s", resp.StatusCode, string(raw))
		return fmt.Errorf("telegram send failed: status %d", resp.StatusCode)
	}

	// Parse Telegram API response { ok: bool, description: string, ... }
	var tr struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &tr); err != nil {
		log.Printf("[TelegramSender] Response unmarshal failed: %v, raw=%s", err, string(raw))
		return err
	}
	if !tr.Ok {
		log.Printf("[TelegramSender] Telegram API rejected message: %s", tr.Description)
		return fmt.Errorf("telegram api error: %s", tr.Description)
	}

	log.Printf("[TelegramSender] → Sent to %d: %s", chatID, text)
	return nil
}

func (t *TelegramSender) SendTyping(chatID int64) error {
	if t.Token == "" {
		log.Println("[TelegramSender] ⚠️ Token not set — sendTyping skipped")
		return fmt.Errorf("telegram token not set")
	}

	if t.Client == nil {
		t.Client = &http.Client{Timeout: 10 * time.Second}
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendChatAction", t.Token)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"action":  "typing",
	}
	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[TelegramSender] sendTyping NewRequest failed: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.Client.Do(req)
	if err != nil {
		log.Printf("[TelegramSender] sendTyping HTTP failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		log.Printf("[TelegramSender] sendTyping Non-200 response: status=%d raw=%s", resp.StatusCode, string(raw))
		return fmt.Errorf("telegram sendChatAction failed: status %d", resp.StatusCode)
	}
	var tr struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &tr); err != nil {
		log.Printf("[TelegramSender] sendTyping unmarshal failed: %v, raw=%s", err, string(raw))
		return err
	}
	if !tr.Ok {
		log.Printf("[TelegramSender] sendTyping api rejected: %s", tr.Description)
		return fmt.Errorf("telegram api error: %s", tr.Description)
	}
	log.Printf("[TelegramSender] … typing sent to %d", chatID)
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

func (m *MockEventBus) Publish(topic string, payload interface{}) error {
	log.Printf("[EventBus] %s → %v", topic, payload)
	return nil
}

// ============================================================================
// TASK MANAGER (STUB)
// ============================================================================

type MockTaskManager struct{}

func (m *MockTaskManager) Schedule(taskName string, when time.Time, f func()) error {
	log.Printf("[TaskManager] Schedule: %s at %v", taskName, when)
	return nil
}

// ============================================================================
// TRANSCRIPTION (STUB FOR MVP)
// ============================================================================

type MockTranscriber struct{}

func (t *MockTranscriber) Transcribe(url string) (string, error) {
	log.Printf("[MockTranscriber] Transcribing: %s", url)
	return "[транскрипция недоступна в MVP]", nil
}
