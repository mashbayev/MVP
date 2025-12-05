package main

import (
	"context"
	"log"
	"net/http"

	"whatsapp-analytics-mvp/internal/api"
	"whatsapp-analytics-mvp/internal/config"
	"whatsapp-analytics-mvp/internal/core"
	"whatsapp-analytics-mvp/internal/data"
	"whatsapp-analytics-mvp/internal/infrastructure"
	"whatsapp-analytics-mvp/internal/llm"
	"whatsapp-analytics-mvp/internal/weather"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const geminiModelName = "gemini-1.5-flash" // стабильная модель для tools

// ------------------------------
// MAIN
// ------------------------------

func main() {
	ctx := context.Background()

	// 1) Load Config
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2) Init SQLite
	contextManager, err := data.NewSQLiteContextRepo("whatsapp_analytics.db")
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}
	analyticsRepo := contextManager
	settingsRepo := contextManager

	// 3) Init Gemini client (используем только для инструментов)
	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(cfg.API.GeminiAPIKey))
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	// 4) Init Hybrid LLM Engine (OpenAI → fallback Gemini)
	llmEngine := llm.NewLLMEngine(
		geminiClient,
		cfg.API.OpenAIKey, // добавь в config.yaml поле openai_key
	)

	// 5) Init Weather
	weatherClient := weather.NewClient(
		cfg.API.OpenWeatherMapKey,
		cfg.Location.AstanaLat,
		cfg.Location.AstanaLon,
	)

	// 6) Init Tools Provider
	toolsProvider := infrastructure.NewToolsService(contextManager.DB, weatherClient)

	// 7) Init Senders, Notifiers, Events, Tasks
	telegramSender := infrastructure.NewTelegramSender(cfg.API.TelegramToken)
	wazzupSender := infrastructure.NewWazzupSender(cfg.API.WazzupAPIKey)
	notifier := infrastructure.NewDefaultNotifier()
	eventBus := &infrastructure.MockEventBus{}
	taskManager := &infrastructure.MockTaskManager{}

	// 8) Init AIService (теперь с гибридным движком)
	aiService := core.NewAIService(
		llmEngine,
		geminiModelName,
		analyticsRepo,
		settingsRepo,
		nil, // transcriber добавите позже
		notifier,
		contextManager,
		eventBus,
		taskManager,
		toolsProvider,
		weatherClient,
		geminiClient,
	)

	// 9) API router
	apiHandler := api.NewAPIHandler(
		aiService,
		wazzupSender,
		telegramSender,
		nil, // transcriber (если пока нет)
	)

	router := api.SetupRouter(apiHandler)

	// 10) Start HTTP Server
	log.Printf("Server running on %s...", cfg.App.Port)
	if err := http.ListenAndServe(cfg.App.Port, router); err != nil {
		log.Fatal(err)
	}
}
