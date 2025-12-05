package core

import (
	"context"
	"time"

	"whatsapp-analytics-mvp/internal/models"
	"whatsapp-analytics-mvp/internal/weather"
)

//
// ============================================================================
//  LLM / AI ENGINES
// ============================================================================
//

// LLMProvider — единый интерфейс для всех движков (OpenAI, Gemini, Hybrid).
type LLMProvider interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

//
// ============================================================================
//  TOOLS PROVIDER (BUSINESS LOGIC)
// ============================================================================
//

// ToolsProvider — интерфейс доступа к бизнес-операциям (бронь, цена, слоты).
type ToolsProvider interface {
	CheckAvailability(ctx context.Context, date, time string, seats int) (string, error)
	GetPrice(ctx context.Context, seats, hours int, time string) (string, error)
	CreateBooking(ctx context.Context, clientID, date, time string, seats, hours int) (string, error)
	GeneratePaymentLink(ctx context.Context, amount float64, bookingID string) (string, error)
}

//
// ============================================================================
//  WEATHER PROVIDER
// ============================================================================
//

// WeatherProvider — абстракция клиента погоды.
type WeatherProvider interface {
	GetCurrentWeather(ctx context.Context) (*weather.WeatherData, error)
	GetForecast(ctx context.Context, date time.Time) (*weather.WeatherData, error)
}

//
// ============================================================================
//  REPOSITORIES
// ============================================================================
//

// AnalyticsRepo — аналитический репозиторий (продажи, отчеты).
type AnalyticsRepo interface {
	GetSalesReport(ctx context.Context, startDate, endDate string) (map[string]interface{}, error)
	GetSalesDetail(ctx context.Context, filter string) (map[string]interface{}, error)
	SaveLog(ctx context.Context, entry models.DialogLog) error
}

// ContextManager — управление клиентом, сессией и историей переписки.
type ContextManager interface {
	GetProfile(ctx context.Context, clientID string) (*models.ClientProfile, error)
	GetChatHistory(ctx context.Context, clientID string) ([]map[string]string, error)

	SaveMessage(ctx context.Context, clientID, sender, text string) error
	CreateOrUpdateSession(ctx context.Context, clientID string, bookingID *string) error
}

//
// ============================================================================
//  NOTIFIER / EVENTS / TASKS
// ============================================================================
//

// NotificationProvider — отправка админ-логов и предупреждений.
type NotificationProvider interface {
	NotifyAdmin(message string) error
}

// EventBus — шина событий (для будущих маркетинговых триггеров).
type EventBus interface {
	Publish(eventName string, payload interface{}) error
}

// TaskManager — планировщик заданий.
type TaskManager interface {
	Schedule(taskName string, when time.Time, fn func()) error
}

//
// ============================================================================
//  TRANSCRIPTION PROVIDER
// ============================================================================
//

// TranscriptionProvider — обработка голосовых сообщений.
type TranscriptionProvider interface {
	Transcribe(audioURL string) (string, error)
}

//
// ============================================================================
//  SENDERS (WHATSAPP / TELEGRAM)
// ============================================================================
//

// WhatsAppSender — отправка сообщений WA (через Wazzup).
type WhatsAppSender interface {
	Send(channelID, clientPhone, messageText string) error
}

// TelegramSender — отправка сообщений телеграм-клиенту.
type TelegramSender interface {
	Send(chatID int64, text string) error
	SendTyping(chatID int64) error
	GetFileDirectURL(fileID string) (string, error)
}
