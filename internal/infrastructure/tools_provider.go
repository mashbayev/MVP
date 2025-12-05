package infrastructure

import (
	"context"
	"whatsapp-analytics-mvp/internal/core"
)

// ToolsProviderAdapter адаптирует ToolsService под интерфейс core.ToolsProvider.
type ToolsProviderAdapter struct {
	svc *ToolsService
}

// NewToolsProviderAdapter создаёт адаптер.
func NewToolsProviderAdapter(svc *ToolsService) core.ToolsProvider {
	return &ToolsProviderAdapter{svc: svc}
}

// -----------------------------------------------------------------------------
// Реализация методов интерфейса ToolsProvider
// -----------------------------------------------------------------------------

func (a *ToolsProviderAdapter) CheckAvailability(ctx context.Context, date, time string, seats int) (string, error) {
	return a.svc.CheckAvailability(ctx, date, time, seats)
}

func (a *ToolsProviderAdapter) GetPrice(ctx context.Context, seats, hours int, time string) (string, error) {
	return a.svc.GetPrice(ctx, seats, hours, time)
}

func (a *ToolsProviderAdapter) CreateBooking(ctx context.Context, clientID, date, time string, seats, hours int) (string, error) {
	return a.svc.CreateBooking(ctx, clientID, date, time, seats, hours)
}

func (a *ToolsProviderAdapter) GeneratePaymentLink(ctx context.Context, amount float64, bookingID string) (string, error) {
	return a.svc.GeneratePaymentLink(ctx, amount, bookingID)
}
