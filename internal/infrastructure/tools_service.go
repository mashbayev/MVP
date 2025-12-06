package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"whatsapp-analytics-mvp/internal/models"
)

// ToolsService — основная бизнес-логика для инструментов, вызываемых LLM.
type ToolsService struct {
	DB bookingStore
}

// NewToolsService — создаёт сервис инструментов.
// bookingStore — minimal interface required by ToolsService (booking operations).
type bookingStore interface {
	GetBookingsAt(ctx context.Context, t time.Time) ([]models.Booking, error)
	SaveBooking(ctx context.Context, bookingID, clientID string, start time.Time, seats, hours int, amountStr string) error
}

func NewToolsService(db bookingStore) *ToolsService {
	return &ToolsService{DB: db}
}

// -----------------------------------------------------------------------------
// Проверка доступности
// -----------------------------------------------------------------------------

func (s *ToolsService) CheckAvailability(ctx context.Context, date, timeStr string, seats int) (string, error) {
	if seats <= 0 || seats > 6 {
		return "", fmt.Errorf("количество мест должно быть от 1 до 6")
	}

	parsedTime, err := time.Parse("2006-01-02 15:04", date+" "+timeStr)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты/времени. Используйте YYYY-MM-DD и HH:MM")
	}

	// Считаем брони на выбранный час
	bookings, err := s.DB.GetBookingsAt(ctx, parsedTime)
	if err != nil {
		return "", err
	}

	usedSeats := 0
	for _, b := range bookings {
		usedSeats += b.Seats
	}

	if usedSeats+seats > 6 {
		return "Мест недостаточно", nil
	}

	return "Места доступны", nil
}

// -----------------------------------------------------------------------------
// Расчёт стоимости
// -----------------------------------------------------------------------------

func (s *ToolsService) GetPrice(ctx context.Context, seats, hours int, timeStr string) (string, error) {
	if seats <= 0 || seats > 6 {
		return "", fmt.Errorf("места: 1–6")
	}
	if hours <= 0 || hours > 12 {
		return "", fmt.Errorf("часы: 1–12")
	}

	// Простая тарифная логика
	basePrice := 2000.0
	nightMultiplier := 1.0

	parsedTime, err := time.Parse("15:04", timeStr)
	if err == nil {
		// После 22:00 — цена выше
		if parsedTime.Hour() >= 22 {
			nightMultiplier = 1.25
		}
	}

	total := basePrice * float64(seats) * float64(hours) * nightMultiplier
	return fmt.Sprintf("%.0f", total), nil
}

// -----------------------------------------------------------------------------
// Создание брони
// -----------------------------------------------------------------------------

func (s *ToolsService) CreateBooking(ctx context.Context, clientID, date, timeStr string, seats, hours int) (string, error) {
	if seats <= 0 || seats > 6 {
		return "", fmt.Errorf("места: 1–6")
	}
	if hours <= 0 || hours > 12 {
		return "", fmt.Errorf("часы: 1–12")
	}

	startTime, err := time.Parse("2006-01-02 15:04", date+" "+timeStr)
	if err != nil {
		return "", fmt.Errorf("неверная дата/время")
	}

	bookingID := fmt.Sprintf("bk_%d", time.Now().UnixNano())

	priceStr, _ := s.GetPrice(ctx, seats, hours, timeStr)

	err = s.DB.SaveBooking(ctx, bookingID, clientID, startTime, seats, hours, priceStr)
	if err != nil {
		return "", err
	}

	log.Printf("✓ Booking created: %s (%s %s) seats=%d hours=%d", bookingID, date, timeStr, seats, hours)

	return bookingID, nil
}

// -----------------------------------------------------------------------------
// Генерация платёжной ссылки (MVP stub)
// -----------------------------------------------------------------------------

func (s *ToolsService) GeneratePaymentLink(ctx context.Context, amount float64, bookingID string) (string, error) {
	if amount <= 0 {
		return "", fmt.Errorf("некорректная сумма")
	}

	// В боевом продукте будет интеграция Kaspi / Stripe / WalletPay
	return fmt.Sprintf("https://pay.example.com/?booking=%s&amount=%.0f", bookingID, amount), nil
}
