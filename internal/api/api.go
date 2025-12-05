package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"whatsapp-analytics-mvp/internal/core"

	"github.com/go-chi/chi/v5"
)

// ADMIN TELEGRAM ID (если нужно)
const ADMIN_TELEGRAM_ID int64 = 779270468

// ==========================================================
// STRUCTS
// ==========================================================

// WhatsApp (Wazzup) webhook format
type WazzupWebhook struct {
	ChannelID string `json:"channelId"`
	Messages  []struct {
		Text      string `json:"text"`
		ChatID    string `json:"chatID"`
		Direction string `json:"direction"`
		Type      string `json:"type"`
		AudioURL  string `json:"audioUrl,omitempty"`
	} `json:"messages"`
}

// Telegram update
type TelegramUpdate struct {
	UpdateID int64 `json:"update_id"`
	Message  *struct {
		MessageID int64 `json:"message_id"`
		From      *struct {
			ID int64 `json:"id"`
		} `json:"from,omitempty"`

		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`

		Text  string `json:"text"`
		Voice *struct {
			FileID string `json:"file_id"`
		} `json:"voice,omitempty"`
	} `json:"message,omitempty"`
}

// ==========================================================
// API Handler
// ==========================================================

type APIHandler struct {
	Service        *core.AIService
	WazzupSender   core.WhatsAppSender
	TelegramSender core.TelegramSender
	Transcriber    core.TranscriptionProvider
}

func NewAPIHandler(
	service *core.AIService,
	wazzupSender core.WhatsAppSender,
	telegramSender core.TelegramSender,
	transcriber core.TranscriptionProvider,
) *APIHandler {
	return &APIHandler{
		Service:        service,
		WazzupSender:   wazzupSender,
		TelegramSender: telegramSender,
		Transcriber:    transcriber,
	}
}

// ==========================================================
// ROUTER
// ==========================================================

func SetupRouter(h *APIHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hybrid AI Engine is running."))
	})

	r.Post("/webhook/wazzup", h.HandleWazzupWebhook)
	r.Post("/webhook/telegram", h.HandleTelegramWebhook)

	return r
}

// ==========================================================
// TELEGRAM HANDLER
// ==========================================================

func (h *APIHandler) HandleTelegramWebhook(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	log.Printf("📩 TG RAW: %s\n", string(body))

	var upd TelegramUpdate
	if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
		log.Printf("❌ Telegram decode error: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	if upd.Message == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	chatID := upd.Message.Chat.ID
	clientID := fmt.Sprintf("TG-%d", chatID)
	userMessage := upd.Message.Text

	// Voice message
	if upd.Message.Voice != nil && upd.Message.Voice.FileID != "" {
		log.Println("🎤 TG voice message detected")

		fileURL, err := h.TelegramSender.GetFileDirectURL(upd.Message.Voice.FileID)
		if err == nil {
			text, err := h.Transcriber.Transcribe(fileURL)
			if err == nil {
				userMessage = text
			}
		}
	}

	isAdmin := upd.Message.From != nil && upd.Message.From.ID == ADMIN_TELEGRAM_ID

	// Answer immediately, process async
	w.WriteHeader(http.StatusOK)

	go func() {
		_ = h.TelegramSender.SendTyping(chatID)

		reply, err := h.Service.ProcessMessage(clientID, userMessage, isAdmin)
		if err != nil {
			reply = "Ошибка сервера. Попробуйте позже."
		}

		_ = h.TelegramSender.Send(chatID, reply)
	}()
}

// ==========================================================
// WAZZUP (WHATSAPP) HANDLER
// ==========================================================

func (h *APIHandler) HandleWazzupWebhook(w http.ResponseWriter, r *http.Request) {
	var wh WazzupWebhook

	if err := json.NewDecoder(r.Body).Decode(&wh); err != nil {
		log.Printf("❌ Wazzup decode error: %v", err)
		w.WriteHeader(http.StatusOK)
		return
	}

	for _, m := range wh.Messages {
		if m.Direction != "inbound" {
			continue
		}

		clientID := "WA-" + m.ChatID
		userMessage := m.Text

		if m.Type == "audio" && m.AudioURL != "" {
			log.Println("🎤 WA audio message detected")

			text, err := h.Transcriber.Transcribe(m.AudioURL)
			if err == nil {
				userMessage = text
			}
		}

		reply, err := h.Service.ProcessMessage(clientID, userMessage, false)
		if err != nil {
			reply = "Ошибка сервера. Попробуйте позже."
		}

		if err := h.WazzupSender.Send(wh.ChannelID, m.ChatID, reply); err != nil {
			log.Printf("❌ Wazzup send error: %v", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}
