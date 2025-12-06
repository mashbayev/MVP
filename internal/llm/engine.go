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
	}

	// Convert tools if provided (genai.Tool -> openai.Tool)
	if tools != nil {
		if toolList, ok := tools.([]*genai.Tool); ok {
			req.Tools = genaiToolsToOpenAI(toolList)
		}
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

	for _, p := range resp.Candidates[0].Content.Parts {
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

// genaiToolsToOpenAI converts genai.Tool list to openai.Tool list
func genaiToolsToOpenAI(genaiTools []*genai.Tool) []openai.Tool {
	if len(genaiTools) == 0 {
		return nil
	}

	var result []openai.Tool
	for _, gt := range genaiTools {
		if gt == nil {
			continue
		}

		// Convert each FunctionDeclaration to openai.Tool
		for _, fd := range gt.FunctionDeclarations {
			if fd == nil {
				continue
			}

			tool := openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        fd.Name,
					Description: fd.Description,
					Parameters:  schemaToJSON(fd.Parameters),
				},
			}
			result = append(result, tool)
		}
	}

	return result
}

// schemaToJSON converts genai.Schema to openai-compatible JSON schema
func schemaToJSON(schema *genai.Schema) interface{} {
	if schema == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": "object",
	}

	// genai.Type is not comparable with a plain string in some versions,
	// use fmt.Sprint to safely stringify the enum/value.
	result["type"] = fmt.Sprint(schema.Type)

	if schema.Description != "" {
		result["description"] = schema.Description
	}

	if len(schema.Properties) > 0 {
		props := make(map[string]interface{})
		for key, prop := range schema.Properties {
			props[key] = schemaToJSON(prop)
		}
		result["properties"] = props
	}

	if len(schema.Required) > 0 {
		result["required"] = schema.Required
	}

	return result
}
