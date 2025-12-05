package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App struct {
		Port string `yaml:"port"`
	} `yaml:"app"`

	API struct {
		OpenAIKey         string `yaml:"openai_api_key"`
		OpenWeatherMapKey string `yaml:"openweathermap_key"`
		TelegramToken     string `yaml:"telegram_token"`
		WazzupAPIKey      string `yaml:"wazzup_api_key"`
	} `yaml:"api"`

	Location struct {
		AstanaLat float64 `yaml:"astana_lat"`
		AstanaLon float64 `yaml:"astana_lon"`
	} `yaml:"location"`
}

func LoadConfig(configPath string) (*Config, error) {
	if !filepath.IsAbs(configPath) {
		if wd, _ := os.Getwd(); wd != "" {
			configPath = filepath.Join(wd, configPath)
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфига (%s): %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("ошибка парсинга YAML: %w", err)
	}

	// Вычитываем из окружения, если нужно
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.API.OpenAIKey = v
	}

	if cfg.API.OpenAIKey == "" {
		return nil, fmt.Errorf("КРИТИЧЕСКАЯ ОШИБКА: OpenAI API key не указан (api.openai_api_key или env OPENAI_API_KEY)")
	}

	if cfg.App.Port == "" {
		cfg.App.Port = ":8080"
		log.Println("[CONFIG] ⚠️ Port не указан, использован :8080")
	}
	if cfg.App.Port[0] != ':' {
		cfg.App.Port = ":" + cfg.App.Port
	}

	if cfg.API.OpenWeatherMapKey == "" {
		log.Println("[CONFIG] ⚠️ OpenWeatherMap API key отсутствует. Модуль погоды работать не будет.")
	}
	if cfg.API.TelegramToken == "" {
		log.Println("[CONFIG] ⚠️ Telegram Token отсутствует. Telegram webhook работать не будет.")
	}

	return &cfg, nil
}
