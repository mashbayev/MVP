package core

import (
	"context"
	"fmt"
	"log"
	"time"
)

// -----------------------------------------------------------------------------
//  ADMIN TOOL DISPATCHER
// -----------------------------------------------------------------------------

func (s *AIService) handleAdminToolCall(
	ctx context.Context,
	name string,
	args map[string]interface{},
) (string, error) {

	log.Printf("🛠️ Admin Tool Call: %s | Args=%v", name, args)

	switch name {

	case "GetSalesDetailTool":
		filters, _ := args["filters"].(string)
		return s.GetSalesDetailTool(ctx, filters)

	case "GetMarketingStatsTool":
		return s.GetMarketingStatsTool(ctx)

	case "GetWeatherTool":
		date, _ := args["date"].(string)
		return s.GetWeatherTool(ctx, date)

	case "GetRevenueByDateRangeTool":
		start, _ := args["start_date"].(string)
		end, _ := args["end_date"].(string)
		return s.GetRevenueByDateRangeTool(ctx, start, end)

	case "GetSalesRecommendationTool":
		return s.GetSalesRecommendationTool(ctx)

	default:
		return fmt.Sprintf("Ошибка: неизвестный инструмент администратора '%s'", name), nil
	}
}

// -----------------------------------------------------------------------------
//  SALES DETAIL (ANALYTICS)
// -----------------------------------------------------------------------------

func (s *AIService) GetSalesDetailTool(ctx context.Context, filters string) (string, error) {
	repo, ok := s.ContextManager.(AnalyticsRepo)
	if !ok {
		return "Ошибка: репозиторий аналитики недоступен.", nil
	}

	data, err := repo.GetSalesDetail(ctx, filters)
	if err != nil {
		return fmt.Sprintf("Ошибка аналитики: %v", err), nil
	}

	// Специальный фильтр "today"
	if filters == "today" {

		total, _ := data["total_bookings"].(int)
		if total == 0 {
			return "Сегодня ещё нет продаж.", nil
		}

		return fmt.Sprintf(
			"Сегодня: %d броней. Популярный час: %s. Брони на 4 места: %d. Средний чек за место: %.0f тг.",
			data["total_bookings"],
			data["popular_hour"],
			data["four_seat_bookings"],
			data["avg_price_per_seat"],
		), nil
	}

	// Default (last 30 days)
	return fmt.Sprintf(
		"30 дней: %d броней, выручка: %.0f тг, средний чек: %.0f тг.",
		data["total_bookings"],
		data["total_revenue"],
		data["avg_check"],
	), nil
}

// -----------------------------------------------------------------------------
//  MARKETING STATS
// -----------------------------------------------------------------------------

func (s *AIService) GetMarketingStatsTool(ctx context.Context) (string, error) {
	return "Маркетинг: Instagram +250 подписчиков за 7 дней. Конверсия WA→бронь: 35%.", nil
}

// -----------------------------------------------------------------------------
//  WEATHER ANALYTICS
// -----------------------------------------------------------------------------

func (s *AIService) GetWeatherTool(ctx context.Context, date string) (string, error) {

	if s.WeatherClient == nil {
		return "Ошибка: Клиент погоды не настроен.", nil
	}

	weatherData, err := s.WeatherClient.GetCurrentWeather(ctx)
	if err != nil {
		return fmt.Sprintf("Ошибка погоды: %v", err), nil
	}

	analysis := "Хорошая погода, ожидается стабильная посещаемость."
	if weatherData.Temp < -10 || weatherData.PrecipProb > 0.5 {
		analysis = "Плохая погода — возможен спад посещаемости."
	} else if weatherData.Temp > 25 {
		analysis = "Жаркая погода — дневная посещаемость может просесть."
	}

	return fmt.Sprintf(
		"Погода: %.1f°C, %s, ветер %.1f м/с. Аналитика: %s",
		weatherData.Temp,
		weatherData.Condition,
		weatherData.WindSpeed,
		analysis,
	), nil
}

// -----------------------------------------------------------------------------
//  REVENUE RANGE ANALYTICS
// -----------------------------------------------------------------------------

func (s *AIService) GetRevenueByDateRangeTool(ctx context.Context, startDate string, endDate string) (string, error) {

	repo, ok := s.ContextManager.(AnalyticsRepo)
	if !ok {
		return "Ошибка: репозиторий аналитики недоступен.", nil
	}

	data, err := repo.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		return fmt.Sprintf("Ошибка аналитики: %v", err), nil
	}

	return fmt.Sprintf(
		"Выручка %s → %s: %.0f тг, %d бронирований, средний чек %.0f тг.",
		startDate, endDate,
		data["total_revenue"],
		data["total_bookings"],
		data["average_check"],
	), nil
}

// -----------------------------------------------------------------------------
//  SALES RECOMMENDATION TOOL
// -----------------------------------------------------------------------------

func (s *AIService) GetSalesRecommendationTool(ctx context.Context) (string, error) {

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	sales, _ := s.GetRevenueByDateRangeTool(ctx, yesterday, yesterday)
	if sales == "" {
		sales = "Нет данных за вчера."
	}

	weather, _ := s.GetWeatherTool(ctx, "today")
	if weather == "" {
		weather = "Погода недоступна."
	}

	return fmt.Sprintf(
		"Комбинированная аналитика: продажи вчера (%s): %s. Погода сегодня: %s.",
		yesterday, sales, weather,
	), nil
}
