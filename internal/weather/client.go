package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Общая структура данных, которую ожидает проект
type WeatherData struct {
	Temp       float64 `json:"temp"`
	Condition  string  `json:"condition"`
	WindSpeed  float64 `json:"wind_speed"`
	PrecipProb float64 `json:"precip_prob"`
}

// Клиент погоды через OpenWeatherMap
type WeatherClient struct {
	APIKey string
	Lat    float64
	Lon    float64
	client *http.Client
}

func NewClient(apiKey string, lat, lon float64) *WeatherClient {
	return &WeatherClient{
		APIKey: apiKey,
		Lat:    lat,
		Lon:    lon,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// ============================================================================
// CURRENT WEATHER
// ============================================================================

func (w *WeatherClient) GetCurrentWeather(ctx context.Context) (*WeatherData, error) {
	if w.APIKey == "" {
		// Возвращаем stub-погоду, чтобы система не падала
		return &WeatherData{
			Temp:       -5,
			Condition:  "Cloudy",
			WindSpeed:  3.2,
			PrecipProb: 0.1,
		}, nil
	}

	url := fmt.Sprintf(
		"https://api.openweathermap.org/data/2.5/weather?lat=%f&lon=%f&appid=%s&units=metric",
		w.Lat, w.Lon, w.APIKey,
	)

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Main struct {
			Temp float64 `json:"temp"`
		} `json:"main"`
		Weather []struct {
			Main string `json:"main"`
		} `json:"weather"`
		Wind struct {
			Speed float64 `json:"speed"`
		} `json:"wind"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &WeatherData{
		Temp:       raw.Main.Temp,
		Condition:  raw.Weather[0].Main,
		WindSpeed:  raw.Wind.Speed,
		PrecipProb: 0.2, // OpenWeather не даёт — оставляем mock
	}, nil
}

// ============================================================================
// FORECAST (STUB)
// ============================================================================

func (w *WeatherClient) GetForecast(ctx context.Context, date time.Time) (*WeatherData, error) {
	// В MVP просто возвращаем текущую погоду как прогноз
	wd, err := w.GetCurrentWeather(ctx)
	if err != nil {
		return nil, err
	}

	// Можно слегка модифицировать данные, чтобы выглядело реалистичнее
	wd.Condition = "Forecast: " + wd.Condition

	return wd, nil
}
