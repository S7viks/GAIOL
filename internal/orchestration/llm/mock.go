package llm

import (
	"context"
	"time"
)

// MockAdapter echoes subtask content for local dev/tests.
type MockAdapter struct {
	Provider string
	Prefix   string
	Latency  int64
}

func NewMockAdapter(providerID string) *MockAdapter {
	return &MockAdapter{Provider: providerID, Prefix: "[mock]", Latency: 2}
}

func (m *MockAdapter) ProviderID() string {
	if m.Provider == "" {
		return "mock"
	}
	return m.Provider
}

func (m *MockAdapter) Generate(_ context.Context, params GenerateParams) (GenerateResult, error) {
	latency := m.Latency
	if latency <= 0 {
		latency = 2
	}
	lastUser := ""
	for i := len(params.Messages) - 1; i >= 0; i-- {
		if params.Messages[i].Role == "user" {
			lastUser = params.Messages[i].Content
			break
		}
	}
	text := m.Prefix + " " + params.Model + ": " + lastUser
	pt, ct := 10, 20
	return GenerateResult{
		Text:      text,
		LatencyMs: latency,
		Usage: &GenerateUsage{
			PromptTokens:     pt,
			CompletionTokens: ct,
			CostUsd:          0,
		},
	}, nil
}

// SleepAdapter simulates latency (unused by default).
type SleepAdapter struct {
	Provider string
	Delay    time.Duration
}

func (s *SleepAdapter) ProviderID() string { return s.Provider }

func (s *SleepAdapter) Generate(ctx context.Context, params GenerateParams) (GenerateResult, error) {
	if s.Delay > 0 {
		select {
		case <-ctx.Done():
			return GenerateResult{}, ctx.Err()
		case <-time.After(s.Delay):
		}
	}
	m := NewMockAdapter(s.Provider)
	return m.Generate(ctx, params)
}
