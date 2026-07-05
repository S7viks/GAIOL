package llm

import (
	"context"
	"strings"
)

// ChatMessage is one chat turn for provider adapters.
type ChatMessage struct {
	Role    string
	Content string
}

// GenerateParams is input to ProviderAdapter.Generate.
type GenerateParams struct {
	TraceID         string
	Model           string
	Messages        []ChatMessage
	Temperature     *float64
	MaxOutputTokens *int
}

// GenerateUsage is token/cost usage.
type GenerateUsage struct {
	PromptTokens     int
	CompletionTokens int
	CostUsd          float64
}

// GenerateResult is a provider response.
type GenerateResult struct {
	Text      string
	LatencyMs int64
	Usage     *GenerateUsage
}

// ProviderAdapter is the orchestration provider contract.
type ProviderAdapter interface {
	ProviderID() string
	Generate(ctx context.Context, params GenerateParams) (GenerateResult, error)
}

func MessagesToPrompt(messages []ChatMessage) string {
	var b strings.Builder
	for _, m := range messages {
		b.WriteString(m.Role)
		b.WriteString(": ")
		b.WriteString(m.Content)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}
