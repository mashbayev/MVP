package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	openai "github.com/sashabaranov/go-openai"
)

// LLMEngine — гибридный движок OpenAI → Gemini (fallback)
type LLMEngine struct {
	openaiClient *openai.Client
	geminiClient *genai.Client
	modelOpenAI  string
	modelGemini  string
}

// NewLLMEngine создаёт движок
func NewLLMEngine(
	openaiKey string,
	openAIModel string,
	geminiClient *genai.Client,
	geminiModel string,
) *LLMEngine {

	var oc *openai.Client
	if openaiKey != "" && !strings.Contains(openaiKey, "placeholder") {
		oc = openai.NewClient(openaiKey)
		log.Println("[LLM Engine] OpenAI client initialized.")
	} else {
		log.Println("[LLM Engine] ⚠️ OpenAI key missing, fallback-only mode active.")
	}

	return &LLMEngine{
		openaiClient: oc,
		geminiClient: geminiClient,
		modelOpenAI:  openAIModel,
		modelGemini:  geminiModel,
	}
}

// -------------------------------
// PUBLIC API
// -------------------------------

// Generate — единая точка входа
func (e *LLMEngine) Generate(
	ctx context.Context,
	systemPrompt string,
	userPrompt string,
	tools any, // список инструментов (OpenAI tools)
) (string, error, bool) {
	// bool → был ли ответ от OpenAI (true) или Gemini fallback (false)

	// ------------------------------------
	// 1) Сначала пытаемся OpenAI
	// ------------------------------------
	if e.openaiClient != nil {
		reply, err := e.callOpenAI(ctx, systemPrompt, userPrompt, tools)
		if err == nil {
			return reply, nil, true // успех OpenAI
		}

		if isQuotaError(err) || isRateLimit(err) || isUnavailable(err) {
			log.Printf("[LLM Engine] OpenAI overload/quota → fallback to Gemini. ERR: %v", err)
		} else {
			log.Printf("[LLM Engine] OpenAI unexpected error → fallback: %v", err)
		}
	}

	// ------------------------------------
	// 2) Fallback → Gemini
	// ------------------------------------
	reply, err := e.callGemini(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", fmt.Errorf("ни OpenAI, ни Gemini не смогли ответить: %w", err), false
	}

	return reply, nil, false
}

// -------------------------------
// INTERNAL: OpenAI
// -------------------------------

func (e *LLMEngine) callOpenAI(ctx context.Context, systemPrompt, userPrompt string, tools any) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: e.modelOpenAI,
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Tools: tools,
	}

	resp, err := e.openaiClient.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("OpenAI вернул пустой ответ")
	}

	return resp.Choices[0].Message.Content, nil
}

// -------------------------------
// INTERNAL: Gemini fallback
// -------------------------------

func (e *LLMEngine) callGemini(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	model := e.geminiClient.GenerativeModel(e.modelGemini)

	resp, err := model.GenerateContent(ctx,
		genai.Text(systemPrompt+"\n"+userPrompt),
	)
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("Gemini вернул пустой fallback ответ")
	}

	for _, p := range resp.Candidates[0].Content.Pparts {
		if txt, ok := p.(genai.Text); ok {
			return string(txt), nil
		}
	}

	return "", errors.New("Gemini fallback: не удалось извлечь текст")
}

// -------------------------------
// ERROR HELPERS
// -------------------------------

func isQuotaError(err error) bool {
	return strings.Contains(err.Error(), "quota")
}

func isRateLimit(err error) bool {
	return strings.Contains(err.Error(), "429")
}

func isUnavailable(err error) bool {
	return strings.Contains(err.Error(), "unavailable") ||
		strings.Contains(err.Error(), "try again") ||
		strings.Contains(err.Error(), "timeout")
}
