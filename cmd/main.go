package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

// const geminiModelName = "gemini-1.5-flash" // модель для инструментов
const geminiModelName = "gemini-2.5-flash" // модель для инструментов

func main() {
	ctx := context.Background()

	// --------------------------
	// 1) LOAD CONFIG
	// --------------------------
	cfg, err := config.LoadConfig("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// --------------------------
	// 2) SQLITE
	// --------------------------
	contextManager, err := data.NewSQLiteContextRepo("whatsapp_analytics.db")
	if err != nil {
		log.Fatalf("DB init failed: %v", err)
	}

	// --------------------------
	// 3) GEMINI client (tools)
	// --------------------------
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Fatal("GEMINI_API_KEY is not set in environment")
	}

	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(geminiKey))
	if err != nil {
		log.Fatalf("Gemini client init failed: %v", err)
	}

	// --------------------------
	// 4) HYBRID LLM ENGINE
	// --------------------------
	llmEngine := llm.NewLLMEngine(
		cfg.API.OpenAIKey, // openaiKey
		"gpt-4o-mini",     // openAI model
		geminiClient,      // Gemini client
		geminiModelName,   // Gemini model
	)

	// --------------------------
	// 5) WEATHER CLIENT
	// --------------------------
	weatherClient := weather.NewClient(
		cfg.API.OpenWeatherMapKey,
		cfg.Location.AstanaLat,
		cfg.Location.AstanaLon,
	)

	// --------------------------
	// 6) TOOLS PROVIDER
	// --------------------------
	toolsProvider := infrastructure.NewToolsService(contextManager)

	// --------------------------
	// 7) SENDERS & INFRA
	// --------------------------
	telegramSender := infrastructure.NewTelegramSender(cfg.API.TelegramToken)
	wazzupSender := infrastructure.NewWazzupSender(cfg.API.WazzupAPIKey)
	notifier := &infrastructure.DefaultNotifier{}
	eventBus := &infrastructure.MockEventBus{}
	taskManager := &infrastructure.MockTaskManager{}

	// --------------------------
	// 8) AI SERVICE
	// --------------------------
	aiService := core.NewAIService(
		llmEngine,       // LLMProvider
		geminiModelName, // modelName
		nil,             // TranscriptionProvider
		notifier,        // NotificationProvider
		contextManager,  // ContextManager
		eventBus,        // EventBus
		taskManager,     // TaskManager
		toolsProvider,   // ToolsProvider
		weatherClient,   // WeatherProvider
		geminiClient,    // *genai.Client
	)

	// --------------------------
	// 9) API ROUTER
	// --------------------------
	apiHandler := api.NewAPIHandler(
		aiService,
		wazzupSender,
		telegramSender,
		nil, // Transcriber (пока нет)
	)

	router := api.SetupRouter(apiHandler)

	// --------------------------
	// 10) START SERVER
	// --------------------------
	log.Printf("Server running on %s...", cfg.App.Port)
	if err := http.ListenAndServe(cfg.App.Port, router); err != nil {
		log.Fatal(err)
	}
}
