package models

import "time"

// -----------------------------------------------------------------------------
// CLIENT PROFILE
// -----------------------------------------------------------------------------

// ClientProfile — расширенная модель клиента, которую получает AI.
type ClientProfile struct {
	ClientID     string  `json:"client_id"`
	Name         string  `json:"name"`
	Lang         string  `json:"lang"`
	LoyaltyLevel string  `json:"loyalty_level"`
	TotalSpent   float64 `json:"total_spent"`

	// История сообщений — JSON-строка, а не структура
	History string `json:"history"`
}

// -----------------------------------------------------------------------------
// BOOKING MODEL (используется ToolsService и ContextRepo)
// -----------------------------------------------------------------------------

type Booking struct {
	BookingID string    `json:"booking_id"`
	ClientID  string    `json:"client_id"`
	Start     time.Time `json:"start"`
	Seats     int       `json:"seats"`
	Hours     int       `json:"hours"`
	Amount    float64   `json:"amount"`
}

// -----------------------------------------------------------------------------
// ANALYTICS LOG MODEL
// -----------------------------------------------------------------------------

// DialogLog — используется для аналитики и ML метрик.
type DialogLog struct {
	ClientID    string    `json:"client_id"`
	Timestamp   time.Time `json:"timestamp"`
	MessageText string    `json:"message_text"`
	Intent      string    `json:"intent"`
	LeadSource  string    `json:"lead_source"`
	Sentiment   string    `json:"sentiment"`
}

// -----------------------------------------------------------------------------
// BUSINESS SETTINGS (фиксированные параметры бизнеса)
// -----------------------------------------------------------------------------

type BusinessSettings struct {
	BusinessName string `json:"business_name"`
	Address      string `json:"address"`
	WorkingHours string `json:"working_hours"`
}
