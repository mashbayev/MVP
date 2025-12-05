package core

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"whatsapp-analytics-mvp/internal/models"

	"github.com/google/generative-ai-go/genai"
)

// ================================
// LLMProvider (интерфейс движка)
// ================================

// Вариант минимального интерфейса, чтобы не создавать отдельный файл.
// Если у тебя уже есть определение — оставь своё, это совместимо.
type LLMProvider interface {
	// Generate возвращает (ответ, ошибка, wasOpenAI)
	Generate(ctx context.Context, systemPrompt, userPrompt string, tools any) (string, error, bool)
}

// ================================
// AIService
// ================================

// AIService — основной сервис обработки сообщений.
type AIService struct {
	// --- Гибридный движок (OpenAI → fallback на Gemini) ---
	LLMEngine LLMProvider
	ModelName string // имя модели Gemini для инструментов

	// --- Хранилища/адаптеры ---
	AnalyticsRepo AnalyticsAdapter
	SettingsRepo  SettingsRepository
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
	analyticsRepo AnalyticsAdapter,
	settingsRepo SettingsRepository,
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
		AnalyticsRepo:  analyticsRepo,
		SettingsRepo:   settingsRepo,
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
	if repo, ok := s.ContextManager.(interface {
		CreateOrUpdateSession(ctx context.Context, clientID string, bookingID *string) error
	}); ok {
		_ = repo.CreateOrUpdateSession(ctx, clientID, nil)
	}

	// 2) System prompt + tools
	var systemInstruction string
	if isAdmin {
		systemInstruction = s.getAdminSystemPrompt()
	} else {
		systemInstruction = s.getClientSystemPrompt()
	}

	// 3) ЛЁГКИЙ ПУТЬ: сначала пробуем гибридный LLMEngine (OpenAI → Gemini-fallback)
	// ВАЖНО: здесь мы не передаем инструменты, это быстрый ответ для "простых" реплик.
	// Если нужно вызвать инструменты — перейдем к инструментальному пайплайну ниже.
	if s.LLMEngine != nil {
		reply, err, wasOpenAI := s.LLMEngine.Generate(ctx, systemInstruction, userMessage, nil)
		if err == nil && strings.TrimSpace(reply) != "" {
			// Сохраняем и выходим
			if repo, ok := s.ContextManager.(interface {
				SaveMessage(ctx context.Context, clientID, sender, text string) error
			}); ok {
				_ = repo.SaveMessage(ctx, clientID, "bot", reply)
			}
			go s.saveAnalyticsLog(clientID, userMessage, reply)
			log.Printf("[AI] Quick reply via %s", map[bool]string{true: "OpenAI", false: "Gemini-fallback"}[wasOpenAI])
			return reply, nil
		}
		// Если движок не дал валидный ответ — продолжаем штатный пайплайн с инструментами.
		log.Printf("[AI] Quick reply not used (err=%v). Continue with tools...", err)
	}

	// 4) ИНСТРУМЕНТАЛЬНЫЙ ПАЙПЛАЙН (через Gemini Tools)
	//    Он нужен для цен/броней/линков и админ-аналитики.
	if s.Client == nil {
		// Без прямого Gemini-клиента инструменты недоступны: мягкий ответ
		return "Сейчас не могу обработать запрос полностью. Напиши: на когда, сколько мест и на сколько часов?", nil
	}

	model := s.Client.GenerativeModel(s.ModelName)
	// Tools
	if isAdmin {
		model.Tools = getAdminTools()
	} else {
		model.Tools = getClientTools()
	}

	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemInstruction)}}
	chat := model.StartChat()

	// 5) История (если доступна)
	if repo, ok := s.ContextManager.(interface {
		GetChatHistory(ctx context.Context, clientID string) ([]map[string]string, error)
	}); ok {
		if hist, err := repo.GetChatHistory(ctx, clientID); err == nil && len(hist) > 0 {
			h := make([]*genai.Content, 0, len(hist))
			for _, m := range hist {
				role := m["role"]
				txt := m["text"]
				if txt == "" {
					continue
				}
				h = append(h, &genai.Content{
					Role:  role,
					Parts: []genai.Part{genai.Text(txt)},
				})
			}
			chat.History = h
			log.Printf("📚 Injected history for %s: %d msgs", clientID, len(h))
		}
	}

	// 6) Первичный ответ модели (Gemini)
	log.Printf("🔧 DEBUG: Send → Gemini (admin=%v, tools=%d)", isAdmin, len(model.Tools))
	resp, err := chat.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		// Попробуем честно отреагировать: уведомим и вернём мягкий ответ
		s.notify(fmt.Sprintf("Gemini API error (initial) for %s: %v", clientID, err))
		return "Извини, сейчас перегрузка. Попробуй через минуту.", fmt.Errorf("gemini send failed: %w", err)
	}

	// 7) Tool loop — обрабатываем FunctionCall → FunctionResponse
	finalResp, err := s.handleToolLoop(ctx, chat, resp, isAdmin, clientID)
	if err != nil {
		s.notify(fmt.Sprintf("Tool loop error for %s: %v", clientID, err))
		return "Произошёл технический сбой. Давай начнём с простого: на когда нужна бронь и на сколько мест?", nil
	}

	// 8) Достаём финальный текст
	text := extractFirstText(finalResp)
	if strings.TrimSpace(text) == "" {
		text = "Продолжим. На какое время, сколько мест и на сколько часов планируешь?"
	}

	// 9) Сохранение и лог
	if repo, ok := s.ContextManager.(interface {
		SaveMessage(ctx context.Context, clientID, sender, text string) error
	}); ok {
		_ = repo.SaveMessage(ctx, clientID, "bot", text)
	}
	go s.saveAnalyticsLog(clientID, userMessage, text)

	return text, nil
}

// handleToolLoop — цикл обработки function_call → function_response
func (s *AIService) handleToolLoop(
	ctx context.Context,
	chat *genai.ChatSession,
	firstResp *genai.GenerateContentResponse,
	isAdmin bool,
	clientID string,
) (*genai.GenerateContentResponse, error) {

	const maxSteps = 3
	resp := firstResp

	for step := 0; step < maxSteps; step++ {
		call := firstFunctionCall(resp)
		if call == nil {
			return resp, nil
		}

		// Выполняем инструмент
		var toolOutput string
		if isAdmin {
			out, _ := s.handleAdminToolCall(ctx, call.Name, call.Args) // ошибки → в текст
			toolOutput = out
		} else {
			out, _ := s.dispatchClientTool(ctx, call.Name, call.Args, clientID)
			toolOutput = out
		}

		fnResp := genai.FunctionResponse{
			Name: call.Name,
			Response: map[string]any{
				"result": toolOutput,
			},
		}

		next, err := chat.SendMessage(ctx, fnResp)
		if err != nil {
			s.notify(fmt.Sprintf("Gemini API error (tool step) for %s: %v", clientID, err))
			return nil, fmt.Errorf("gemini function response failed: %w", err)
		}

		resp = next
	}

	return resp, nil
}

// dispatchClientTool — маршрутизация клиентских инструментов к ToolsProvider.
func (s *AIService) dispatchClientTool(ctx context.Context, name string, args map[string]any, clientID string) (string, error) {
	switch name {
	case "CheckAvailability":
		date, _ := strArg(args, "date")
		tm, _ := strArg(args, "time")
		seats, _ := floatArg(args, "seats")
		return s.ToolsProvider.CheckAvailability(ctx, date, tm, int(seats))

	case "GetPrice":
		seats, _ := floatArg(args, "seats")
		hours, _ := floatArg(args, "hours")
		tm, _ := strArg(args, "time")
		return s.ToolsProvider.GetPrice(ctx, int(seats), int(hours), tm)

	case "CreateBooking":
		date, _ := strArg(args, "date")
		tm, _ := strArg(args, "time")
		seats, _ := floatArg(args, "seats")
		hours, _ := floatArg(args, "hours")
		return s.ToolsProvider.CreateBooking(ctx, clientID, date, tm, int(seats), int(hours))

	case "GeneratePaymentLink":
		amount, _ := floatArg(args, "amount")
		bookingID, _ := strArg(args, "bookingID")
		return s.ToolsProvider.GeneratePaymentLink(ctx, amount, bookingID)

	default:
		return fmt.Sprintf("Ошибка: неизвестный инструмент '%s'", name), nil
	}
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

// -----------------------------
// Helpers
// -----------------------------

func (s *AIService) notify(msg string) {
	if s.Notifier != nil {
		_ = s.Notifier.NotifyAdmin(msg)
	}
}

type fnCall struct {
	Name string
	Args map[string]any
}

func firstFunctionCall(resp *genai.GenerateContentResponse) *fnCall {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil
	}
	for _, p := range resp.Candidates[0].Content.Parts {
		if fc, ok := p.(genai.FunctionCall); ok {
			return &fnCall{Name: fc.Name, Args: fc.Args}
		}
	}
	return nil
}

func extractFirstText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}
	for _, p := range resp.Candidates[0].Content.Parts {
		if t, ok := p.(genai.Text); ok {
			return string(t)
		}
	}
	return ""
}

func strArg(m map[string]any, key string) (string, bool) {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s, true
		}
	}
	return "", false
}

func floatArg(m map[string]any, key string) (float64, bool) {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f, true
		}
		// Иногда модели присылают int
		if i, ok := v.(int); ok {
			return float64(i), true
		}
	}
	return 0, false
}
