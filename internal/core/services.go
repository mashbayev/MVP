package core

import (
	"context"
	"log"
	"strings"
	"time"

	"whatsapp-analytics-mvp/internal/models"

	"github.com/google/generative-ai-go/genai"
)

// ================================
// AIService
// ================================

// AIService — основной сервис обработки сообщений.
type AIService struct {
	// --- Гибридный движок (OpenAI → fallback на Gemini) ---
	LLMEngine LLMProvider
	ModelName string // имя модели Gemini для инструментов

	// --- Хранилища/адаптеры ---
	AnalyticsRepo AnalyticsRepo
	Transcription TranscriptionProvider
	Notifier      NotificationProvider

	// --- NEW ARCHITECTURE DEPENDENCIES ---
	ContextManager ContextManager  // Retrieves enriched client profile
	EventBus       EventBus        // Publishes events
	TaskManager    TaskManager     // Schedules tasks
	ToolsProvider  ToolsProvider   // Executes business logic tools
	WeatherClient  WeatherProvider // Provides weather data for analytics

	// --- (опционально) прямой доступ к Gemini для инструментов ---
	// Если твой LLMEngine внутри уже содержит genai.Client — можно удалить это поле.
	// Но тут оставляем: инструменты по-прежнему вызываем через Gemini.
	Client *genai.Client
}

// NewAIService — конструктор.
func NewAIService(
	llmEngine LLMProvider, // <— гибридный движок
	modelName string,
	transcriber TranscriptionProvider,
	notifier NotificationProvider,
	contextManager ContextManager,
	eventBus EventBus,
	taskManager TaskManager,
	toolsProvider ToolsProvider,
	weatherClient WeatherProvider,
	// Дополнительно: явный Gemini client для инструментов
	geminiClient *genai.Client,
) *AIService {
	return &AIService{
		LLMEngine:      llmEngine,
		ModelName:      modelName,
		Transcription:  transcriber,
		Notifier:       notifier,
		ContextManager: contextManager,
		EventBus:       eventBus,
		TaskManager:    taskManager,
		ToolsProvider:  toolsProvider,
		WeatherClient:  weatherClient,
		Client:         geminiClient,
	}
}

// ProcessMessage — ядро контроллера. Возвращает ответ агента.
func (s *AIService) ProcessMessage(clientID, userMessage string, isAdmin bool) (string, error) {
	ctx := context.Background()

	// 1) Persist incoming message (best-effort)
	if repo, ok := s.ContextManager.(interface {
		SaveMessage(ctx context.Context, clientID, sender, text string) error
	}); ok {
		_ = repo.SaveMessage(ctx, clientID, "client", userMessage)
	}

	// 2) System prompt selection
	var systemInstruction string
	if isAdmin {
		systemInstruction = s.getAdminSystemPrompt()
	} else {
		systemInstruction = s.getClientSystemPrompt()
	}

	// 3) Get tools for this request
	var tools any
	if isAdmin {
		tools = GetAdminTools()
	} else {
		tools = GetClientTools()
	}

	// 4) Call hybrid LLM engine with tools
	if s.LLMEngine != nil {
		reply, err, wasOpenAI := s.LLMEngine.Generate(ctx, systemInstruction, userMessage, tools)
		if err == nil && strings.TrimSpace(reply) != "" {
			// Save bot response
			if repo, ok := s.ContextManager.(interface {
				SaveMessage(ctx context.Context, clientID, sender, text string) error
			}); ok {
				_ = repo.SaveMessage(ctx, clientID, "bot", reply)
			}
			go s.saveAnalyticsLog(clientID, userMessage, reply)
			log.Printf("[AI] Reply via %s", map[bool]string{true: "OpenAI", false: "Gemini"}[wasOpenAI])
			return reply, nil
		}
		log.Printf("[AI] LLM error: %v", err)
	}

	// 5) Fallback: generic response
	return "Извини, сейчас не могу обработать запрос. Попробуй позже.", nil
}

// saveAnalyticsLog — best-effort лог.
func (s *AIService) saveAnalyticsLog(clientID, userMessage, reply string) {
	if s.AnalyticsRepo == nil {
		return
	}
	entry := models.DialogLog{
		ClientID:    clientID,
		Timestamp:   time.Now(),
		MessageText: userMessage,
		Intent:      "unknown",
		LeadSource:  "whatsapp",
		Sentiment:   "neutral",
	}
	_ = s.AnalyticsRepo.SaveLog(context.Background(), entry)
}

// notify — best-effort уведомление администратора.
func (s *AIService) notify(msg string) {
	if s.Notifier != nil {
		_ = s.Notifier.NotifyAdmin(msg)
	}
}
