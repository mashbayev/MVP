package core

import "github.com/google/generative-ai-go/genai"

// -----------------------------
// CLIENT TOOLS
// -----------------------------

func GetClientTools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{

				{
					Name:        "CheckAvailability",
					Description: "Проверяет доступность мест на указанную дату и время.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"date": {
								Type:        genai.TypeString,
								Description: "Дата YYYY-MM-DD",
							},
							"time": {
								Type:        genai.TypeString,
								Description: "Время HH:MM",
							},
							"seats": {
								Type:        genai.TypeInteger,
								Description: "Количество мест",
							},
						},
						Required: []string{"date", "time", "seats"},
					},
				},

				{
					Name:        "GetPrice",
					Description: "Рассчитывает стоимость брони.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"seats": {
								Type:        genai.TypeInteger,
								Description: "Количество мест",
							},
							"hours": {
								Type:        genai.TypeInteger,
								Description: "Количество часов",
							},
							"time": {
								Type:        genai.TypeString,
								Description: "Время начала (опционально)",
							},
						},
						Required: []string{"seats", "hours"},
					},
				},

				{
					Name:        "CreateBooking",
					Description: "Создаёт бронь.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"date": {
								Type:        genai.TypeString,
								Description: "Дата YYYY-MM-DD",
							},
							"time": {
								Type:        genai.TypeString,
								Description: "Время HH:MM",
							},
							"seats": {
								Type:        genai.TypeInteger,
								Description: "Кол-во мест",
							},
							"hours": {
								Type:        genai.TypeInteger,
								Description: "Сколько часов",
							},
						},
						Required: []string{"date", "time", "seats", "hours"},
					},
				},

				{
					Name:        "GeneratePaymentLink",
					Description: "Генерирует ссылку на оплату.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"amount": {
								Type:        genai.TypeNumber,
								Description: "Сумма к оплате",
							},
							"bookingID": {
								Type:        genai.TypeString,
								Description: "ID брони",
							},
						},
						Required: []string{"amount", "bookingID"},
					},
				},
			},
		},
	}
}

// -----------------------------
// ADMIN TOOLS
// -----------------------------

func GetAdminTools() []*genai.Tool {
	return []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{

				{
					Name:        "GetSalesDetailTool",
					Description: "Детальная аналитика продаж по фильтрам.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"filters": {
								Type:        genai.TypeString,
								Description: "today | last30 | client_id:XXX | seats=4 и т.п.",
							},
						},
						Required: []string{"filters"},
					},
				},

				{
					Name:        "GetMarketingStatsTool",
					Description: "Возвращает статистику маркетинга.",
					Parameters:  &genai.Schema{Type: genai.TypeObject},
				},

				{
					Name:        "GetWeatherTool",
					Description: "Получает прогноз или текущую погоду.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"date": {
								Type:        genai.TypeString,
								Description: "Дата или today",
							},
						},
						Required: []string{"date"},
					},
				},

				{
					Name:        "GetRevenueByDateRangeTool",
					Description: "Аналитика выручки за выбранный период.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"start_date": {Type: genai.TypeString},
							"end_date":   {Type: genai.TypeString},
						},
						Required: []string{"start_date", "end_date"},
					},
				},

				{
					Name:        "GetSalesRecommendationTool",
					Description: "Генерирует рекомендации по продажам, используя данные и погоду.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"reason": {
								Type:        genai.TypeString,
								Description: "Причина вызова (optional)",
							},
						},
					},
				},
			},
		},
	}
}
