package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gaiol/internal/models"
	"gaiol/internal/models/adapters"
	"gaiol/internal/uaip"
)

// ModelAdapterBridge wraps a models.ModelAdapter for orchestration.
type ModelAdapterBridge struct {
	Provider string
	Adapter  models.ModelAdapter
}

func (b *ModelAdapterBridge) ProviderID() string { return b.Provider }

func (b *ModelAdapterBridge) Generate(ctx context.Context, params GenerateParams) (GenerateResult, error) {
	start := time.Now()
	prompt := MessagesToPrompt(params.Messages)
	req := &uaip.UAIPRequest{}
	req.Payload.Input.Data = prompt
	if params.MaxOutputTokens != nil && *params.MaxOutputTokens > 0 {
		req.Payload.OutputRequirements.MaxTokens = *params.MaxOutputTokens
	}
	if params.Temperature != nil {
		req.Payload.OutputRequirements.Temperature = *params.Temperature
	}
	resp, err := b.Adapter.GenerateText(ctx, params.Model, req)
	if err != nil {
		return GenerateResult{}, err
	}
	if resp == nil {
		return GenerateResult{}, fmt.Errorf("adapter returned nil response")
	}
	latency := time.Since(start).Milliseconds()
	pt, ct := tokensFromUAIPResult(resp.Result)
	return GenerateResult{
		Text:      resp.Result.Data,
		LatencyMs: latency,
		Usage: &GenerateUsage{
			PromptTokens:     pt,
			CompletionTokens: ct,
			CostUsd:          0,
		},
	}, nil
}

func tokensFromUAIPResult(r uaip.Result) (prompt, completion int) {
	if r.Metadata != nil {
		if v, ok := r.Metadata["prompt_tokens"]; ok {
			prompt = intFromMeta(v)
		}
		if v, ok := r.Metadata["completion_tokens"]; ok {
			completion = intFromMeta(v)
		}
	}
	if prompt == 0 && completion == 0 && r.TokensUsed > 0 {
		completion = r.TokensUsed
	}
	return prompt, completion
}

func intFromMeta(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case int32:
		return int(n)
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

// NewOpenAICompatible creates an OpenAI-compatible orchestration adapter.
func NewOpenAICompatible(providerID, baseURL, apiKey string) ProviderAdapter {
	return &ModelAdapterBridge{
		Provider: providerID,
		Adapter:  adapters.NewOpenAICompatibleAdapter(providerID, baseURL, "Authorization", "Bearer", apiKey),
	}
}

// NewAnthropic creates an Anthropic orchestration adapter.
func NewAnthropic(apiKey, baseURL string) ProviderAdapter {
	a := adapters.NewAnthropicAdapter("anthropic", baseURL, apiKey)
	return &ModelAdapterBridge{Provider: "anthropic", Adapter: a}
}

// NewGemini creates a Gemini orchestration adapter.
func NewGemini(apiKey string) ProviderAdapter {
	return &ModelAdapterBridge{
		Provider: "google",
		Adapter:  adapters.NewGeminiAdapter(apiKey),
	}
}

// NewOllama creates an Ollama OpenAI-compatible adapter.
func NewOllama(baseURL string) ProviderAdapter {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "http://localhost:11434"
	}
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	return NewOpenAICompatible("ollama", base, "ollama")
}

// AdapterForCredential builds a provider adapter from credential metadata.
func AdapterForCredential(id, kind, apiKey, baseURL string) (ProviderAdapter, error) {
	switch kind {
	case "openai_compatible":
		if strings.TrimSpace(apiKey) == "" {
			return nil, fmt.Errorf("missing api key for %s", id)
		}
		b := strings.TrimSpace(baseURL)
		if id == "openrouter" && b == "" {
			b = "https://openrouter.ai/api/v1"
		}
		return NewOpenAICompatible(id, b, apiKey), nil
	case "anthropic_messages":
		if strings.TrimSpace(apiKey) == "" {
			return nil, fmt.Errorf("missing api key for anthropic")
		}
		return NewAnthropic(apiKey, baseURL), nil
	case "gemini":
		if strings.TrimSpace(apiKey) == "" {
			return nil, fmt.Errorf("missing api key for gemini")
		}
		return NewGemini(apiKey), nil
	case "ollama":
		return NewOllama(baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported credential kind %q", kind)
	}
}
